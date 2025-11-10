package models

import "time"

// TrustedOrigin represents an allowed domain for dashboard access
// Enables CNAME-based custom domains for multi-tenant dashboard access
type TrustedOrigin struct {
	ID          int       `json:"id"`
	Domain      string    `json:"domain"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
