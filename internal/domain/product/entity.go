package product

import "time"

type Product struct {
	ID          string
	AddedAt     time.Time
	Type        string
	ReceptionID string
}

var AllowedTypes = []string{"electronics", "clothes", "shoes"}
