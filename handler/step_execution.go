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
	g := r.Group("/step-executions")
	// Aplicar autenticación JWT a todas las rutas
	g.Use(middleware.AuthMiddleware(cfg))

	g.GET("", middleware.RequireScopes("step-executions:read"), func(c *gin.Context) {

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

		query := db.Model(&model.StepExecution{})

		var err error
		query, err = service.ApplyStepExecutionFilters(query, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameter"})
			return
		}

		var list []model.StepExecution
		if err := query.
			Scopes(service.Paginate(page, size)).
			Preload("Execution").
			Preload("Step").
			Find(&list).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch step executions"})
			return
		}

		c.JSON(http.StatusOK, list)
	})

	// PUT: Actualizar step execution usando query params (execution_id y step_id)
	g.PUT("", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {
		// Obtener step exec id del query param
		executionIDStr := c.Query("id")
		if executionIDStr == "" {
			// También aceptar executionId para compatibilidad
			executionIDStr = c.Query("id")
		}

		// Validar que al menos execution_id esté presente
		if executionIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execution_id query parameter is required"})
			return
		}

		stepExecutionID, err := strconv.ParseUint(executionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution_id. Must be a number"})
			return
		}

		// Construir la query de búsqueda
		query := db.Where("id = ?", stepExecutionID)

		// Buscar la step execution
		var se model.StepExecution
		if err := query.First(&se).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found for the given id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Parsear el JSON como un mapa para extraer solo los campos que nos interesan
		var jsonData map[string]interface{}
		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Preparar campos para actualizar usando Updates() para asegurar que se persista
		updateFields := make(map[string]interface{})

		// Extraer status si existe en el JSON
		if statusValue, exists := jsonData["status"]; exists && statusValue != nil {
			if statusStr, ok := statusValue.(string); ok && statusStr != "" {
				updateFields["status"] = statusStr
			}
		}

		// Extraer output si existe en el JSON
		if outputValue, exists := jsonData["output"]; exists && outputValue != nil {
			if outputStr, ok := outputValue.(string); ok {
				updateFields["output"] = outputStr
			}
		}

		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update. Expected 'status' or 'output'"})
			return
		}

		// Usar Updates() en lugar de Save() para asegurar que los cambios se persistan
		// Updates() solo actualiza los campos especificados y no requiere campos zero values
		result := db.Model(&se).Updates(updateFields)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution", "details": result.Error.Error()})
			return
		}

		// Verificar que se actualizó al menos una fila
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found or no changes made"})
			return
		}

		// Cargar relaciones actualizadas
		db.Preload("Step").Preload("Execution").First(&se, se.ID)

		c.JSON(http.StatusOK, se)
	})

	// PATCH: Actualización parcial usando query params (execution_id y step_id)
	g.PATCH("", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {
		// Obtener execution_id del query param
		executionIDStr := c.Query("execution_id")
		if executionIDStr == "" {
			executionIDStr = c.Query("executionId")
		}

		// Obtener step_id del query param
		stepIDStr := c.Query("step_id")
		if stepIDStr == "" {
			stepIDStr = c.Query("stepId")
		}

		// Validar que al menos execution_id esté presente
		if executionIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "execution_id query parameter is required"})
			return
		}

		executionID, err := strconv.ParseUint(executionIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution_id. Must be a number"})
			return
		}

		// Construir la query de búsqueda
		query := db.Where("execution_id = ?", executionID)

		// Si step_id está presente, agregarlo a la búsqueda
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

		// Permitir actualizar solo status y output
		allowedFields := map[string]bool{
			"status": true,
			"output": true,
		}

		updateFields := make(map[string]interface{})
		for key, value := range updateData {
			if allowedFields[key] {
				// Asegurar que no enviamos nil o valores inválidos
				if value != nil {
					updateFields[key] = value
				}
			}
		}

		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
			return
		}

		// Usar Updates() que siempre persiste los cambios
		if err := db.Model(&se).Updates(updateFields).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution", "details": err.Error()})
			return
		}

		// Cargar relaciones actualizadas
		db.Preload("Step").Preload("Execution").First(&se, se.ID)

		c.JSON(http.StatusOK, se)
	})
}
