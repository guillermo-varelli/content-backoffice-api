package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"example.com/workflowapi/config"
	"example.com/workflowapi/middleware"
	"example.com/workflowapi/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContentReview struct {
	ID               uint64    `json:"id"`
	ExecutionID      uint64    `json:"execution_id"`
	Title            string    `json:"title"`
	ShortDescription string    `json:"short_description"`
	Message          string    `json:"message"`
	Status           string    `json:"status"`
	Category         string    `json:"category"`
	SubCategory      string    `json:"sub_category"`
	ImageURL         string    `json:"image_url"`
	ImagePrompt      string    `json:"image_prompt"`
	Slug             string    `json:"slug"`
	CreatedAt        time.Time `json:"created"`
	UpdatedAt        time.Time `json:"last_updated"`
}

func RegisterContentReviewRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	g := r.Group("/content-reviews")
	g.Use(middleware.AuthMiddleware(cfg))

	g.GET("", middleware.RequireScopes("content-reviews:read"), func(c *gin.Context) {

		var entities []model.Content

		// 🔎 Query params
		status := c.Query("status")
		executionID := c.Query("execution_id")
		category := c.Query("category")
		from := c.Query("from")
		to := c.Query("to")

		// 📄 Paginación
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")
		sort := c.DefaultQuery("sort", "created desc")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 200 {
			limit = 20
		}

		offset := (page - 1) * limit

		query := db.Model(&model.Content{})

		// 🟢 Filtros
		if status != "" {
			query = query.Where("status = ?", status)
		}

		if executionID != "" {
			query = query.Where("execution_id = ?", executionID)
		}

		if category != "" {
			query = query.Where("category = ?", category)
		}

		if from != "" {
			if fromTime, err := time.Parse("2006-01-02", from); err == nil {
				query = query.Where("created >= ?", fromTime)
			}
		}

		if to != "" {
			if toTime, err := time.Parse("2006-01-02", to); err == nil {
				toTime = toTime.Add(24 * time.Hour)
				query = query.Where("created < ?", toTime)
			}
		}

		// 🔢 Total count antes de paginar
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 📦 Aplicar orden + paginación
		if err := query.
			Order(sort).
			Limit(limit).
			Offset(offset).
			Find(&entities).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 🧾 Mapear respuesta
		response := make([]ContentReview, 0, len(entities))
		for _, e := range entities {
			response = append(response, ContentReview{
				ID:               e.ID,
				ExecutionID:      e.ExecutionID,
				Title:            e.Title,
				ShortDescription: e.ShortDescription,
				Message:          e.Message,
				Status:           e.Status,
				Category:         e.Category,
				SubCategory:      e.SubCategory,
				ImageURL:         e.ImageURL,
				ImagePrompt:      e.ImagePrompt,
				Slug:             e.Slug,
				CreatedAt:        e.Created,
				UpdatedAt:        e.LastUpdated,
			})
		}

		// 📤 Respuesta estructurada
		c.JSON(http.StatusOK, gin.H{
			"data": response,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"totalPages": int(math.Ceil(float64(total) / float64(limit))),
			},
		})
	})
	g.DELETE("/:id", middleware.RequireScopes("content-reviews:write"), func(c *gin.Context) {
		id := c.Param("id")
		result := db.Where("id = ?", id).Delete(&model.Content{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	g.PUT("/:id", middleware.RequireScopes("content-reviews:write"), func(c *gin.Context) {
		var entity model.Content

		// 1️⃣ Obtener ID de la URL
		id := c.Param("id")

		// 2️⃣ Buscar registro existente
		if err := db.First(&entity, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "content review not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 3️⃣ Bind JSON de entrada
		var input ContentReview
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 4️⃣ Actualizar campos permitidos
		entity.Title = input.Title
		entity.ShortDescription = input.ShortDescription
		entity.Message = input.Message
		entity.Status = input.Status
		entity.Category = input.Category
		entity.SubCategory = input.SubCategory
		entity.ImageURL = input.ImageURL
		entity.ImagePrompt = input.ImagePrompt
		entity.LastUpdated = time.Now()

		// 5️⃣ Guardar cambios
		if err := db.Save(&entity).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 6️⃣ Responder actualizado
		c.JSON(http.StatusOK, ContentReview{
			ID:               entity.ID,
			ExecutionID:      entity.ExecutionID,
			Title:            entity.Title,
			ShortDescription: entity.ShortDescription,
			Message:          entity.Message,
			Status:           entity.Status,
			Category:         entity.Category,
			SubCategory:      entity.SubCategory,
			ImageURL:         entity.ImageURL,
			ImagePrompt:      entity.ImagePrompt,
			Slug:             entity.Slug,
			CreatedAt:        entity.Created,
			UpdatedAt:        entity.LastUpdated,
		})
	})
}
