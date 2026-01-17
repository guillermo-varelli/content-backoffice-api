package handler

import (
	"net/http"
	"strconv"

	"example.com/workflowapi/config"
	"example.com/workflowapi/middleware"
	"example.com/workflowapi/model"
	"example.com/workflowapi/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterStepExecutionRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	g := r.Group("/step-executions-grouped")
	// Aplicar autenticación JWT a todas las rutas
	g.Use(middleware.AuthMiddleware(cfg))

	g.GET("", middleware.RequireScopes("step-executions:read"), func(c *gin.Context) {

		query := db.Model(&model.StepExecution{})

		var err error
		query, err = service.ApplyStepExecutionFilters(query, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameter"})
			return
		}

		var list []model.StepExecution
		if err := query.
			Preload("Execution").
			Preload("Execution.Workflow"). // Preload del Workflow
			Preload("Step").
			Order("execution_id, id").
			Find(&list).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch step executions"})
			return
		}

		// Agrupar correctamente
		groups := make(map[uint64]*model.StepExecutionGroupResponse)

		for _, se := range list {
			group, exists := groups[se.ExecutionID]
			if !exists {
				group = &model.StepExecutionGroupResponse{
					ExecutionID: se.ExecutionID,
					Execution: model.ExecutionResponse{
						ID:     se.Execution.ID,
						Status: se.Execution.Status,
						Workflow: model.WorkflowResponse{
							ID:          se.Execution.Workflow.ID,
							Name:        se.Execution.Workflow.Name,
							Description: se.Execution.Workflow.Description,
						},
					},
					Steps: []model.StepExecution{},
				}
				groups[se.ExecutionID] = group
			}

			group.Steps = append(group.Steps, se)
		}

		// Convertir map → slice
		response := make([]model.StepExecutionGroupResponse, 0, len(groups))
		for _, group := range groups {
			response = append(response, *group)
		}

		c.JSON(http.StatusOK, response)
	})

	// PUT: Actualizar step execution usando query params (execution_id y step_id)
	g.PUT("", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {
		executionIDStr := c.Query("id")
		if executionIDStr == "" {
			executionIDStr = c.Query("id")
		}

		if executionIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execution_id query parameter is required"})
			return
		}

		stepExecutionID, err := strconv.ParseUint(executionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution_id. Must be a number"})
			return
		}

		query := db.Where("id = ?", stepExecutionID)

		var se model.StepExecution
		if err := query.First(&se).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found for the given id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var jsonData map[string]interface{}
		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateFields := make(map[string]interface{})

		if statusValue, exists := jsonData["status"]; exists && statusValue != nil {
			if statusStr, ok := statusValue.(string); ok && statusStr != "" {
				updateFields["status"] = statusStr
			}
		}

		if outputValue, exists := jsonData["output"]; exists && outputValue != nil {
			if outputStr, ok := outputValue.(string); ok {
				updateFields["output"] = outputStr
			}
		}

		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update. Expected 'status' or 'output'"})
			return
		}

		result := db.Model(&se).Updates(updateFields)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution", "details": result.Error.Error()})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found or no changes made"})
			return
		}

		db.Preload("Step").Preload("Execution").First(&se, se.ID)

		c.JSON(http.StatusOK, se)
	})

	// PATCH: Actualización parcial usando query params (execution_id y step_id)
	g.PATCH("", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {
		executionIDStr := c.Query("execution_id")
		if executionIDStr == "" {
			executionIDStr = c.Query("executionId")
		}

		stepIDStr := c.Query("step_id")
		if stepIDStr == "" {
			stepIDStr = c.Query("stepId")
		}

		if executionIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execution_id query parameter is required"})
			return
		}

		executionID, err := strconv.ParseUint(executionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution_id. Must be a number"})
			return
		}

		query := db.Where("execution_id = ?", executionID)

		if stepIDStr != "" {
			stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step_id. Must be a number"})
				return
			}
			query = query.Where("step_id = ?", stepID)
		}

		var se model.StepExecution
		if err := query.First(&se).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found for the given execution_id and step_id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		allowedFields := map[string]bool{
			"status": true,
			"output": true,
		}

		updateFields := make(map[string]interface{})
		for key, value := range updateData {
			if allowedFields[key] {
				if value != nil {
					updateFields[key] = value
				}
			}
		}

		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
			return
		}

		if err := db.Model(&se).Updates(updateFields).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution", "details": err.Error()})
			return
		}

		db.Preload("Step").Preload("Execution").First(&se, se.ID)

		c.JSON(http.StatusOK, se)
	})
}
