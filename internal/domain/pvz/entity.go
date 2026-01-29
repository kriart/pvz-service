package pvz

import "time"

type PVZ struct {
	ID        string
	CreatedAt time.Time
	City      string
}

var AllowedCities = []string{"Москва", "Санкт-Петербург", "Казань"}
