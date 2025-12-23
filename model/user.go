package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`  // No exponer en JSON
	Scopes       string    `gorm:"type:text" json:"-"` // Scopes como JSON string
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SetPassword hashea y establece la contraseña del usuario
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifica si la contraseña proporcionada coincide
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// BeforeCreate es un hook de GORM que se ejecuta antes de crear el usuario
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Scopes == "" {
		// Scopes por defecto si no se especifican
		u.Scopes = `["agents:read","agents:write","workflows:read","workflows:write","steps:read","steps:write","step-executions:read","step-executions:write","n:read","n:write"]`
	}
	return nil
}
