package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"example.com/workflowapi/client"
	"example.com/workflowapi/model"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run scripts/create_admin.go <username> <password>")
		fmt.Println("Example: go run scripts/create_admin.go admin admin123456")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	db := client.InitDB()

	// Verificar si ya existe un usuario con ese nombre
	var existingUser model.User
	result := db.Where("username = ?", username).First(&existingUser)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Fatalf("Database error checking user: %v", result.Error)
	}
	if result.Error == nil {
		log.Fatalf("User '%s' already exists in database (ID: %d)!", username, existingUser.ID)
	}

	// Crear usuario admin
	user := model.User{
		Username: username,
		IsActive: true,
	}

	// Hashear password
	if err := user.SetPassword(password); err != nil {
		log.Fatalf("Failed to hash password: %v", err)
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
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✅ Admin user '%s' created successfully!\n", username)
	fmt.Printf("   User ID: %d\n", user.ID)
	fmt.Printf("   Scopes: %s\n", string(scopesJSON))
}
