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

	// ===================== GET =====================
	g.GET("", middleware.RequireScopes("content-reviews:read"), func(c *gin.Context) {

		var entities []model.Content

		status := c.Query("status")
		category := c.Query("category")
		from := c.Query("from")
		to := c.Query("to")

		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")
		sort := c.DefaultQuery("sort", "created desc")

		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}

		limit, _ := strconv.Atoi(limitStr)
		if limit < 1 || limit > 200 {
			limit = 20
		}

		offset := (page - 1) * limit

		query := db.Model(&model.Content{})

		if status != "" {
			query = query.Where("status = ?", status)
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

		var total int64
		query.Count(&total)

		if err := query.
			Order(sort).
			Limit(limit).
			Offset(offset).
			Find(&entities).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := make([]ContentReview, 0, len(entities))
		for _, e := range entities {
			response = append(response, ContentReview{
				ID:               e.ID,
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

	// ===================== POST (CREATE) =====================
	g.POST("", middleware.RequireScopes("content-reviews:write"), func(c *gin.Context) {
		var input struct {
			Title            string `json:"title"`
			ShortDescription string `json:"short_description"`
			Message          string `json:"message"`
			Status           string `json:"status"`
			Category         string `json:"category"`
			SubCategory      string `json:"sub_category"`
			ImageURL         string `json:"image_url"`
			ImagePrompt      string `json:"image_prompt"`
			Slug             string `json:"slug"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
			return
		}

		if input.Status == "" {
			input.Status = "PENDING"
		}

		now := time.Now()

		entity := model.Content{
			Title:            input.Title,
			ShortDescription: input.ShortDescription,
			Message:          input.Message,
			Status:           input.Status,
			Category:         input.Category,
			SubCategory:      input.SubCategory,
			ImageURL:         input.ImageURL,
			ImagePrompt:      input.ImagePrompt,
			Slug:             input.Slug,
			Created:          now,
			LastUpdated:      now,
		}

		if err := db.Create(&entity).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, ContentReview{
			ID:               entity.ID,
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

	// ===================== DELETE =====================
	g.DELETE("/:id", middleware.RequireScopes("content-reviews:write"), func(c *gin.Context) {
		id := c.Param("id")

		if err := db.Delete(&model.Content{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	})

	// ===================== PUT (UPDATE) =====================
	g.PUT("/:id", middleware.RequireScopes("content-reviews:write"), func(c *gin.Context) {
		var entity model.Content
		id := c.Param("id")

		var input ContentReview
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.First(&entity, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
			return
		}

		entity.Title = input.Title
		entity.ShortDescription = input.ShortDescription
		entity.Message = input.Message
		entity.Status = input.Status
		entity.Category = input.Category
		entity.SubCategory = input.SubCategory
		entity.ImageURL = input.ImageURL
		entity.ImagePrompt = input.ImagePrompt
		entity.Slug = input.Slug
		entity.LastUpdated = time.Now()

		if err := db.Save(&entity).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, entity)
	})
}
