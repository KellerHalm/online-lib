package models

import "time"

type Book struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:200;not null"`
	Author      string    `json:"author" gorm:"size:100;not null"`
	Year        int       `json:"year" gorm:"not null"`
	ISBN        string    `json:"isbn" gorm:"size:13;not null;uniqueIndex"`
	IsAvailable bool      `json:"isAvailable" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"createdAt"`
}
