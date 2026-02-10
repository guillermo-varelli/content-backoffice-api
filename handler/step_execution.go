package handler

import (
	"net/http"
	"strconv"
	"time"

	"example.com/workflowapi/config"
	"example.com/workflowapi/middleware"
	"example.com/workflowapi/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterStepExecutionRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	g := r.Group("/step-executions-grouped")
	g.Use(middleware.AuthMiddleware(cfg))

	// =========================
	// GET: List grouped step executions (with pagination)
	// =========================
	g.GET("", middleware.RequireScopes("step-executions:read"), func(c *gin.Context) {

		// ===== pagination =====
		page := 1
		pageSize := 30

		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}

		if ps := c.Query("pageSize"); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil && v > 0 {
				pageSize = v
			}
		}

		offset := (page - 1) * pageSize

		query := db.
			Model(&model.StepExecution{}).
			Joins("JOIN executions ON executions.id = step_executions.execution_id").
			Joins("JOIN workflows ON workflows.id = executions.workflow_id")

		// ===== filters =====
		if status := c.Query("status"); status != "" {
			query = query.Where("executions.status = ?", status)
		}

		if name := c.Query("name"); name != "" {
			query = query.Where(
				"LOWER(workflows.name) LIKE LOWER(?)",
				"%"+name+"%",
			)
		}

		// ===== date range =====
		if from := c.Query("from"); from != "" {
			t, err := time.Parse(time.RFC3339, from)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
				return
			}
			query = query.Where("step_executions.created_at >= ?", t)
		}

		if to := c.Query("to"); to != "" {
			t, err := time.Parse(time.RFC3339, to)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
				return
			}
			query = query.Where("step_executions.created_at <= ?", t)
		}

		// ===== total count =====
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "count error"})
			return
		}

		// ===== fetch =====
		var list []model.StepExecution
		if err := query.
			Preload("Execution").
			Preload("Execution.Workflow").
			Preload("Step").
			Order("step_executions.execution_id DESC, step_executions.id DESC").
			Limit(pageSize).
			Offset(offset).
			Find(&list).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		// ===== group =====
		groups := map[uint64]*model.StepExecutionGroupResponse{}
		order := []uint64{}

		for _, se := range list {
			if _, ok := groups[se.ExecutionID]; !ok {
				groups[se.ExecutionID] = &model.StepExecutionGroupResponse{
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
				order = append(order, se.ExecutionID)
			}
			groups[se.ExecutionID].Steps = append(
				groups[se.ExecutionID].Steps,
				se,
			)
		}

		resp := make([]model.StepExecutionGroupResponse, 0, len(order))
		for _, id := range order {
			resp = append(resp, *groups[id])
		}

		c.JSON(http.StatusOK, gin.H{
			"data": resp,
			"pagination": gin.H{
				"page":       page,
				"pageSize":   pageSize,
				"total":      total,
				"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	})

	// =========================
	// PUT: Update step execution by ID
	// =========================
	g.PUT("", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {

		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "id query parameter is required",
			})
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid id",
			})
			return
		}

		var se model.StepExecution
		if err := db.First(&se, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Step execution not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error",
			})
			return
		}

		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		updates := map[string]interface{}{}

		if v, ok := payload["status"].(string); ok && v != "" {
			updates["status"] = v
		}
		if v, ok := payload["output"].(string); ok {
			updates["output"] = v
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No valid fields to update",
			})
			return
		}

		if err := db.Model(&se).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update step execution",
			})
			return
		}

		db.Preload("Step").Preload("Execution").First(&se, se.ID)
		c.JSON(http.StatusOK, se)
	})
}
