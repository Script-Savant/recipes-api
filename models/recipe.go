package models

import "time"

type Recipe struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags" gorm:"serializer:json"`
	Ingredients  []string  `json:"ingredients" gorm:"serializer:json"`
	Instructions []string  `json:"instructions" gorm:"serializer:json"`
	PublishedAt  time.Time `json:"publishedAt"`
}