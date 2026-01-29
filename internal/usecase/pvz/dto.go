package pvz

import "time"

type PVZInfo struct {
	ID         string
	City       string
	CreatedAt  time.Time
	Receptions []ReceptionInfo
}

type ReceptionInfo struct {
	ID        string
	StartedAt time.Time
	Status    string
	Products  []ProductInfo
}

type ProductInfo struct {
	ID      string
	AddedAt time.Time
	Type    string
}
