package handler

import (
    "net/http"
    "strconv"

    "example.com/workflowapi/model"
    "example.com/workflowapi/service"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func RegisterStepRoutes(r *gin.Engine, db *gorm.DB) {
    g := r.Group("/steps")

    g.POST("", func(c *gin.Context) {
        var s model.Step
        if err := c.ShouldBindJSON(&s); err != nil {
            c.JSON(http.StatusBadRequest, err)
            return
        }
        db.Create(&s)
        c.JSON(http.StatusCreated, s)
    })

    g.GET("", func(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

        var list []model.Step
        db.Scopes(service.Paginate(page, size)).Find(&list)
        c.JSON(http.StatusOK, list)
    })

    g.PUT("/:id", func(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

        var s model.Step
        if err := db.First(&s, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "step not found"})
            return
        }

        var input model.Step
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, err)
            return
        }

        db.Model(&s).Updates(input)
        c.JSON(http.StatusOK, s)
    })

    g.DELETE("/:id", func(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
        db.Delete(&model.Step{}, id)
        c.Status(http.StatusNoContent)
    })
}
