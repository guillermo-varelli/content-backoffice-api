package handler

import (
	"net/http"
	"strconv"

	"example.com/workflowapi/config"
	"example.com/workflowapi/middleware"
	"example.com/workflowapi/model"
	dto "example.com/workflowapi/model/dto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterStepRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	g := r.Group("/steps")
	g.Use(middleware.AuthMiddleware(cfg))

	// ======================
	// GET /steps
	// ======================
	g.GET("", middleware.RequireScopes("steps:read"), func(c *gin.Context) {
		var steps []model.Step

		if err := db.
			Preload("Workflow", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "name", "description")
			}).
			Preload("Agent").
			Find(&steps).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := make([]dto.StepResponseDto, 0, len(steps))
		for _, s := range steps {
			response = append(response, dto.ToStepResponse(s))
		}

		c.JSON(http.StatusOK, response)
	})

	// ======================
	// GET /steps/by-workflow/:workflowId
	// ======================
	g.GET("/by-workflow/:workflowId", middleware.RequireScopes("steps:read"), func(c *gin.Context) {
		workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow id"})
			return
		}

		var steps []model.Step

		if err := db.
			Where("workflow_id = ?", workflowID).
			Preload("Workflow", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "name", "description")
			}).
			Preload("Agent").
			Order("order_index ASC").
			Find(&steps).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := make([]dto.StepResponseDto, 0, len(steps))
		for _, s := range steps {
			response = append(response, dto.ToStepResponse(s))
		}

		c.JSON(http.StatusOK, response)
	})

	// ======================
	// POST /steps
	// AgentID OPCIONAL (0 = null)
	// ======================
	g.POST("", middleware.RequireScopes("steps:write"), func(c *gin.Context) {
		var input dto.StepInputDto
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		// validar workflow (obligatorio)
		var workflow model.Workflow
		if err := db.First(&workflow, input.WorkflowID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workflow not found"})
			return
		}

		// Agent opcional
		var agentID *uint64 = nil

		if input.AgentID != 0 {
			var agent model.Agent
			if err := db.First(&agent, input.AgentID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "agent not found"})
				return
			}
			agentID = &agent.ID
		}

		step := model.Step{
			OrderIndex:    input.OrderIndex,
			Name:          input.Name,
			OperationType: input.OperationType,
			Prompt:        input.Prompt,
			WorkflowID:    workflow.ID,
			AgentID:       agentID, // nil si AgentID == 0
		}

		if err := db.Create(&step).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		db.
			Preload("Workflow", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "name", "description")
			}).
			Preload("Agent").
			First(&step, step.ID)

		c.JSON(http.StatusCreated, dto.ToStepResponse(step))
	})

	// ======================
	// PUT /steps/:id
	// AgentID OBLIGATORIO (como antes)
	// ======================
	g.PUT("/:id", middleware.RequireScopes("steps:write"), func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
			return
		}

		var existing model.Step
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "step not found"})
			return
		}

		var input dto.StepInputDto
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		// validar workflow
		if err := db.First(&model.Workflow{}, input.WorkflowID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "workflow not found"})
			return
		}

		// validar agent (OBLIGATORIO en update)
		if err := db.First(&model.Agent{}, input.AgentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "agent not found"})
			return
		}

		updates := map[string]interface{}{
			"order_index":    input.OrderIndex,
			"name":           input.Name,
			"operation_type": input.OperationType,
			"prompt":         input.Prompt,
			"workflow_id":    input.WorkflowID,
			"agent_id":       input.AgentID,
		}

		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		db.
			Preload("Workflow", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "name", "description")
			}).
			Preload("Agent").
			First(&existing, existing.ID)

		c.JSON(http.StatusOK, dto.ToStepResponse(existing))
	})

	// ======================
	// DELETE /steps/:id
	// ======================
	g.DELETE("/:id", middleware.RequireScopes("steps:write"), func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step id"})
			return
		}

		if err := db.Delete(&model.Step{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	})
}
