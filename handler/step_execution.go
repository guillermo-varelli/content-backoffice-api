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

	// =====================================================
	// GET: List grouped step executions (FULL FILTERED)
	// =====================================================
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

		// =====================================================
		// BASE QUERY
		// =====================================================
		baseQuery := db.
			Model(&model.StepExecution{}).
			Select("step_executions.execution_id").
			Joins("JOIN executions ON executions.id = step_executions.execution_id").
			Joins("JOIN workflows ON workflows.id = executions.workflow_id")

		// =====================================================
		// FILTERS
		// =====================================================

		// status (execution status)
		if status := c.Query("status"); status != "" {
			baseQuery = baseQuery.Where("executions.status = ?", status)
		}

		// workflow name
		if name := c.Query("name"); name != "" {
			baseQuery = baseQuery.Where(
				"LOWER(workflows.name) LIKE LOWER(?)",
				"%"+name+"%",
			)
		}

		// workflowId
		if workflowID := c.Query("workflowId"); workflowID != "" {
			if id, err := strconv.ParseUint(workflowID, 10, 64); err == nil {
				baseQuery = baseQuery.Where("executions.workflow_id = ?", id)
			}
		}

		// execution_id
		if executionID := c.Query("execution_id"); executionID != "" {
			if id, err := strconv.ParseUint(executionID, 10, 64); err == nil {
				baseQuery = baseQuery.Where("step_executions.execution_id = ?", id)
			}
		}

		// stepId
		if stepID := c.Query("stepId"); stepID != "" {
			if id, err := strconv.ParseUint(stepID, 10, 64); err == nil {
				baseQuery = baseQuery.Where("step_executions.step_id = ?", id)
			}
		}

		// date from
		if from := c.Query("from"); from != "" {
			t, err := time.Parse(time.RFC3339, from)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from"})
				return
			}
			baseQuery = baseQuery.Where("step_executions.created_at >= ?", t)
		}

		// date to
		if to := c.Query("to"); to != "" {
			t, err := time.Parse(time.RFC3339, to)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to"})
				return
			}
			baseQuery = baseQuery.Where("step_executions.created_at <= ?", t)
		}

		// =====================================================
		// COUNT DISTINCT EXECUTIONS
		// =====================================================
		var total int64
		if err := db.
			Table("(?) as filtered", baseQuery.Distinct()).
			Count(&total).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "count error"})
			return
		}

		// =====================================================
		// GET PAGINATED EXECUTION IDS
		// =====================================================
		var executionIDs []uint64
		if err := baseQuery.
			Distinct().
			Order("step_executions.execution_id DESC").
			Limit(pageSize).
			Offset(offset).
			Pluck("step_executions.execution_id", &executionIDs).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "id fetch error"})
			return
		}

		if len(executionIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"data": []model.StepExecutionGroupResponse{},
				"pagination": gin.H{
					"page":       page,
					"pageSize":   pageSize,
					"total":      total,
					"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
				},
			})
			return
		}

		// =====================================================
		// FETCH ALL STEPS FOR THOSE EXECUTIONS
		// =====================================================
		var list []model.StepExecution
		if err := db.
			Preload("Execution").
			Preload("Execution.Workflow").
			Preload("Step").
			Where("execution_id IN ?", executionIDs).
			Order("execution_id DESC, id ASC").
			Find(&list).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		// =====================================================
		// GROUP
		// =====================================================
		groups := map[uint64]*model.StepExecutionGroupResponse{}

		for _, id := range executionIDs {
			groups[id] = &model.StepExecutionGroupResponse{
				ExecutionID: id,
				Steps:       []model.StepExecution{},
			}
		}

		for _, se := range list {
			group := groups[se.ExecutionID]

			group.Execution = model.ExecutionResponse{
				ID:     se.Execution.ID,
				Status: se.Execution.Status,
				Workflow: model.WorkflowResponse{
					ID:          se.Execution.Workflow.ID,
					Name:        se.Execution.Workflow.Name,
					Description: se.Execution.Workflow.Description,
				},
			}
			group.Created = se.CreatedAt
			group.Steps = append(group.Steps, se)
		}

		resp := make([]model.StepExecutionGroupResponse, 0, len(executionIDs))
		for _, id := range executionIDs {
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

	// =====================================================
	// PUT: Update step execution
	// =====================================================
	g.PUT("", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {

		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id query parameter is required"})
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
			return
		}

		var se model.StepExecution
		if err := db.First(&se, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
			return
		}

		if err := db.Model(&se).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution"})
			return
		}

		db.Preload("Step").Preload("Execution").First(&se, se.ID)
		c.JSON(http.StatusOK, se)
	})
}
