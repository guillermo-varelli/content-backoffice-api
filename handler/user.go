package handler

import (
	"encoding/json"
	"net/http"

	"example.com/workflowapi/config"
	"example.com/workflowapi/middleware"
	"example.com/workflowapi/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Username string   `json:"username" binding:"required"`
	Password string   `json:"password" binding:"required,min=6"`
	Scopes   []string `json:"scopes,omitempty"`
}

type UpdateUserRequest struct {
	Password *string  `json:"password,omitempty"`
	Scopes   []string `json:"scopes,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// RegisterUserRoutes registra las rutas de gestión de usuarios
func RegisterUserRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	users := r.Group("/users")

	// Ruta pública para crear el primer usuario administrador (solo si no hay usuarios)
	users.POST("/bootstrap", bootstrapAdminHandler(db))

	// Rutas protegidas requieren autenticación y permisos de administrador
	protected := users.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))
	protected.Use(middleware.RequireScopes("users:admin"))

	protected.POST("", createUserHandler(db))
	protected.GET("", listUsersHandler(db))
	protected.GET("/:id", getUserHandler(db))
	protected.PUT("/:id", updateUserHandler(db))
	protected.DELETE("/:id", deleteUserHandler(db))
}

// bootstrapAdminHandler permite crear el primer usuario admin si no existe ningún usuario
func bootstrapAdminHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar si ya existen usuarios
		var count int64
		if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bootstrap only available when no users exist"})
			return
		}

		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Crear usuario admin con todos los permisos
		user := model.User{
			Username: req.Username,
			IsActive: true,
		}

		// Hashear password
		if err := user.SetPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Scopes completos incluyendo users:admin
		adminScopes := []string{
			"agents:read", "agents:write",
			"workflows:read", "workflows:write",
			"steps:read", "steps:write",
			"step-executions:read", "step-executions:write",
			"n:read", "n:write",
			"users:admin",
		}

		scopesJSON, _ := json.Marshal(adminScopes)
		user.Scopes = string(scopesJSON)

		// Guardar en la base de datos
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// No exponer el password hash en la respuesta
		user.PasswordHash = ""
		c.JSON(http.StatusCreated, gin.H{
			"message": "Admin user created successfully",
			"user":    user,
		})
	}
}

func createUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verificar si el usuario ya existe
		var existingUser model.User
		if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}

		// Crear nuevo usuario
		user := model.User{
			Username: req.Username,
			IsActive: true,
		}

		// Hashear password
		if err := user.SetPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Procesar scopes
		scopesJSON, err := json.Marshal(req.Scopes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scopes format"})
			return
		}
		if len(req.Scopes) == 0 {
			// Scopes por defecto
			defaultScopes := []string{
				"agents:read", "agents:write",
				"workflows:read", "workflows:write",
				"steps:read", "steps:write",
				"step-executions:read", "step-executions:write",
				"n:read", "n:write",
			}
			scopesJSON, _ = json.Marshal(defaultScopes)
		}
		user.Scopes = string(scopesJSON)

		// Guardar en la base de datos
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// No exponer el password hash en la respuesta
		user.PasswordHash = ""
		c.JSON(http.StatusCreated, user)
	}
}

func listUsersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []model.User
		if err := db.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		// Limpiar password hashes de la respuesta
		for i := range users {
			users[i].PasswordHash = ""
		}

		c.JSON(http.StatusOK, users)
	}
}

func getUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var user model.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		user.PasswordHash = ""
		c.JSON(http.StatusOK, user)
	}
}

func updateUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var user model.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Actualizar password si se proporciona
		if req.Password != nil {
			if err := user.SetPassword(*req.Password); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
				return
			}
		}

		// Actualizar scopes si se proporcionan
		if req.Scopes != nil {
			scopesJSON, err := json.Marshal(req.Scopes)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scopes format"})
				return
			}
			user.Scopes = string(scopesJSON)
		}

		// Actualizar estado activo si se proporciona
		if req.IsActive != nil {
			user.IsActive = *req.IsActive
		}

		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		user.PasswordHash = ""
		c.JSON(http.StatusOK, user)
	}
}

func deleteUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := db.Delete(&model.User{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
