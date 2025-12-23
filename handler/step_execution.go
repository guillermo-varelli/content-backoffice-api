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

func RegisterStepExecutionRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	g := r.Group("/step-executions")
	// Aplicar autenticación JWT a todas las rutas
	g.Use(middleware.AuthMiddleware(cfg))

	// GET: Listar todas las step executions con paginación y filtros
	g.GET("", middleware.RequireScopes("step-executions:read"), func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

		// Query params para filtros
		executionID := c.Query("executionId") // También acepta execution_id para compatibilidad
		if executionID == "" {
			executionID = c.Query("execution_id")
		}

		stepID := c.Query("stepId") // También acepta step_id para compatibilidad
		if stepID == "" {
			stepID = c.Query("step_id")
		}

		status := c.Query("status")

		query := db.Model(&model.StepExecution{})

		// Filtrar por executionId
		if executionID != "" {
			if execID, err := strconv.ParseUint(executionID, 10, 64); err == nil {
				query = query.Where("execution_id = ?", execID)
			}
		}

		// Filtrar por stepId
		if stepID != "" {
			if stID, err := strconv.ParseUint(stepID, 10, 64); err == nil {
				query = query.Where("step_id = ?", stID)
			}
		}

		// Filtrar por status
		if status != "" {
			query = query.Where("status = ?", status)
		}

		var list []model.StepExecution
		if err := query.Scopes(service.Paginate(page, size)).
			Preload("Step").
			Preload("Execution").
			Find(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch step executions"})
			return
		}

		c.JSON(http.StatusOK, list)
	})

	// GET: Obtener una step execution por ID
	g.GET("/:id", middleware.RequireScopes("step-executions:read"), func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step execution ID"})
			return
		}

		var se model.StepExecution
		if err := db.Preload("Step").
			Preload("Execution").
			First(&se, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, se)
	})

	// GET: Obtener step executions por execution_id
	g.GET("/execution/:execution_id", middleware.RequireScopes("step-executions:read"), func(c *gin.Context) {
		executionID, err := strconv.ParseUint(c.Param("execution_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

		var list []model.StepExecution
		if err := db.Where("execution_id = ?", executionID).
			Scopes(service.Paginate(page, size)).
			Preload("Step").
			Preload("Execution").
			Find(&list).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch step executions"})
			return
		}

		c.JSON(http.StatusOK, list)
	})

	// PUT: Actualizar step executions por execution_id
	g.PUT("/execution/:execution_id", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {
		executionID, err := strconv.ParseUint(c.Param("execution_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
			return
		}

		// Buscar la step execution por execution_id
		var se model.StepExecution
		if err := db.Where("execution_id = ?", executionID).First(&se).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found for this execution_id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Parsear el JSON como un mapa para extraer solo los campos que nos interesan
		var jsonData map[string]interface{}
		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Preparar campos para actualizar usando Updates() para asegurar que se persista
		updateFields := make(map[string]interface{})

		// Extraer status si existe en el JSON
		if statusValue, exists := jsonData["status"]; exists && statusValue != nil {
			if statusStr, ok := statusValue.(string); ok && statusStr != "" {
				updateFields["status"] = statusStr
			}
		}

		// Extraer output si existe en el JSON
		if outputValue, exists := jsonData["output"]; exists && outputValue != nil {
			if outputStr, ok := outputValue.(string); ok {
				updateFields["output"] = outputStr
			}
		}

		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update. Expected 'status' or 'output'"})
			return
		}

		// Usar Updates() en lugar de Save() para asegurar que los cambios se persistan
		// Updates() solo actualiza los campos especificados y no requiere campos zero values
		result := db.Model(&se).Updates(updateFields)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution", "details": result.Error.Error()})
			return
		}

		// Verificar que se actualizó al menos una fila
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found or no changes made"})
			return
		}

		// Cargar relaciones actualizadas
		db.Preload("Step").Preload("Execution").First(&se, se.ID)

		c.JSON(http.StatusOK, se)
	})

	// PUT: Actualizar step executions por execution_id (el parámetro :id se interpreta como execution_id)
	g.PUT("/:id", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {
		executionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
			return
		}

		// Buscar la step execution por execution_id
		var se model.StepExecution
		if err := db.Where("execution_id = ?", executionID).First(&se).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found for this execution_id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Parsear el JSON como un mapa para extraer solo los campos que nos interesan
		var jsonData map[string]interface{}
		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Preparar campos para actualizar usando Updates() para asegurar que se persista
		updateFields := make(map[string]interface{})
		
		// Extraer status si existe en el JSON
		if statusValue, exists := jsonData["status"]; exists && statusValue != nil {
			if statusStr, ok := statusValue.(string); ok && statusStr != "" {
				updateFields["status"] = statusStr
			}
		}
		
		// Extraer output si existe en el JSON
		if outputValue, exists := jsonData["output"]; exists && outputValue != nil {
			if outputStr, ok := outputValue.(string); ok {
				updateFields["output"] = outputStr
			}
		}

		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update. Expected 'status' or 'output'"})
			return
		}

		// Usar Updates() en lugar de Save() para asegurar que los cambios se persistan
		// Updates() solo actualiza los campos especificados y no requiere campos zero values
		result := db.Model(&se).Updates(updateFields)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution", "details": result.Error.Error()})
			return
		}
		
		// Verificar que se actualizó al menos una fila
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found or no changes made"})
			return
		}

		// Cargar relaciones actualizadas
		db.Preload("Step").Preload("Execution").First(&se, se.ID)

		c.JSON(http.StatusOK, se)
	})

	// PATCH: Actualización parcial de una step execution por execution_id
	g.PATCH("/execution/:execution_id", middleware.RequireScopes("step-executions:write"), func(c *gin.Context) {
		executionID, err := strconv.ParseUint(c.Param("execution_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
			return
		}

		var se model.StepExecution
		if err := db.Where("execution_id = ?", executionID).First(&se).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Step execution not found for this execution_id"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Permitir actualizar solo status y output
		allowedFields := map[string]bool{
			"status": true,
			"output": true,
		}

		updateFields := make(map[string]interface{})
		for key, value := range updateData {
			if allowedFields[key] {
				// Asegurar que no enviamos nil o valores inválidos
				if value != nil {
					updateFields[key] = value
				}
			}
		}

		if len(updateFields) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
			return
		}

		// Usar Updates() que siempre persiste los cambios
		if err := db.Model(&se).Updates(updateFields).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update step execution", "details": err.Error()})
			return
		}

		// Cargar relaciones actualizadas
		db.Preload("Step").Preload("Execution").First(&se, se.ID)

		c.JSON(http.StatusOK, se)
	})
}
