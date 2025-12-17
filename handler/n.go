package handler

import (
    "net/http"
    "strconv"

    "example.com/workflowapi/model"
    "example.com/workflowapi/service"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func RegisterNRoutes(r *gin.Engine, db *gorm.DB) {
    g := r.Group("/n")

    g.POST("", func(c *gin.Context) {
        var n model.N
        if err := c.ShouldBindJSON(&n); err != nil {
            c.JSON(http.StatusBadRequest, err)
            return
        }
        db.Create(&n)
        c.JSON(http.StatusCreated, n)
    })

    g.GET("", func(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

        var list []model.N
        db.Scopes(service.Paginate(page, size)).Find(&list)
        c.JSON(http.StatusOK, list)
    })

    g.PUT("/:id", func(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

        var n model.N
        if err := db.First(&n, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
            return
        }

        var input model.N
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, err)
            return
        }

        db.Model(&n).Updates(input)
        c.JSON(http.StatusOK, n)
    })

    g.DELETE("/:id", func(c *gin.Context) {
        id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
        db.Delete(&model.N{}, id)
        c.Status(http.StatusNoContent)
    })
}
