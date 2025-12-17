package handler

import (
    "net/http"
    "strconv"

    "example.com/workflowapi/model"
    "example.com/workflowapi/service"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func RegisterWorkflowRoutes(r *gin.Engine, db *gorm.DB) {
    g := r.Group("/workflows")

    g.POST("", func(c *gin.Context) {
        var w model.Workflow
        if err := c.ShouldBindJSON(&w); err != nil {
            c.JSON(http.StatusBadRequest, err)
            return
        }
        db.Create(&w)
        c.JSON(http.StatusCreated, w)
    })

    g.GET("", func(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

        var list []model.Workflow
        db.Scopes(service.Paginate(page, size)).Find(&list)
        c.JSON(http.StatusOK, list)
    })

    g.PUT("/:id", func(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

        var w model.Workflow
        if err := db.First(&w, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
            return
        }

        var input model.Workflow
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, err)
            return
        }

        db.Model(&w).Updates(input)
        c.JSON(http.StatusOK, w)
    })

    g.DELETE("/:id", func(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
        db.Delete(&model.Workflow{}, id)
        c.Status(http.StatusNoContent)
    })
}
