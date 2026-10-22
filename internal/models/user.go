package models

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"size:200;not null;uniqueIndex"`
	Password  string    `json:"-" gorm:"size:100;not null"`
	CreatedAt time.Time `json:"createdAt"`
}
