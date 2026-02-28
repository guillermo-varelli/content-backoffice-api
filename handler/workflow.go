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

func RegisterWorkflowRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	g := r.Group("/workflows")
	g.Use(middleware.AuthMiddleware(cfg))

	// ==============================
	// GET /workflows
	// ==============================
	g.GET("", middleware.RequireScopes("workflows:read"), func(c *gin.Context) {

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

		enabledParam := c.Query("enabled")

		var workflows []model.Workflow
		query := db.Model(&model.Workflow{})

		// Filtro opcional por enabled
		if enabledParam != "" {
			enabled, err := strconv.ParseBool(enabledParam)
			if err == nil {
				query = query.Where("enabled = ?", enabled)
			}
		}

		if err := query.
			Scopes(service.Paginate(page, size)).
			Find(&workflows).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch workflows"})
			return
		}

		c.JSON(http.StatusOK, workflows)
	})

	// ==============================
	// POST /workflows
	// ==============================
	g.POST("", middleware.RequireScopes("workflows:write"), func(c *gin.Context) {

		var w model.Workflow
		if err := c.ShouldBindJSON(&w); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Create(&w).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workflow"})
			return
		}

		c.JSON(http.StatusCreated, w)
	})

	// ==============================
	// PUT /workflows/:id
	// ==============================
	g.PUT("/:id", middleware.RequireScopes("workflows:write"), func(c *gin.Context) {

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var existing model.Workflow
		if err := db.First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}

		var input model.Workflow
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Usamos map para asegurar que false se actualice
		updates := map[string]interface{}{
			"name":        input.Name,
			"description": input.Description,
			"enabled":     input.Enabled,
		}

		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workflow"})
			return
		}

		// Refetch actualizado
		db.First(&existing, id)

		c.JSON(http.StatusOK, existing)
	})

	// ==============================
	// PATCH /workflows/:id/enabled
	// ==============================
	g.PATCH("/:id/enabled", middleware.RequireScopes("workflows:write"), func(c *gin.Context) {

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var body struct {
			Enabled bool `json:"enabled"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := db.Model(&model.Workflow{}).
			Where("id = ?", id).
			Update("enabled", body.Enabled)

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update enabled state"})
			return
		}

		c.Status(http.StatusOK)
	})

	// ==============================
	// DELETE /workflows/:id
	// ==============================
	g.DELETE("/:id", middleware.RequireScopes("workflows:write"), func(c *gin.Context) {

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		result := db.Delete(&model.Workflow{}, id)

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workflow"})
			return
		}

		c.Status(http.StatusNoContent)
	})
}
