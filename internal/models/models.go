package models

import (
	"time"
)

type BaseModel struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Network struct {
	BaseModel
	Title                string `json:"title" gorm:"size:255;not null"`
	AnalyticsSubgraphURL string `json:"analytics_subgraph_url" gorm:"not null"`
	FarmingSubgraphURL   string `json:"farming_subgraph_url" gorm:"not null"`
	APIKey               string `json:"api_key" gorm:"size:255"`
}

type Pool struct {
	BaseModel
	Title     string   `json:"title" gorm:"size:256;not null"`
	Address   string   `json:"address" gorm:"size:42;not null"`
	LastAPR   *float64 `json:"last_apr"`
	MaxAPR    *float64 `json:"max_apr"`
	NetworkID uint     `json:"network_id"`
	Network   Network  `json:"network" gorm:"foreignKey:NetworkID"`
}

type Farming struct {
	BaseModel
	Hash      string   `json:"hash" gorm:"size:66;uniqueIndex;not null"`
	TVL       *float64 `json:"tvl"`
	LastAPR   *float64 `json:"last_apr"`
	MaxAPR    *float64 `json:"max_apr"`
	NetworkID uint     `json:"network_id"`
	Network   Network  `json:"network" gorm:"foreignKey:NetworkID"`
}

func (Pool) TableName() string {
	return "pools"
}

func (Farming) TableName() string {
	return "farmings"
}

func (Network) TableName() string {
	return "networks"
}
