package models

import "time"

type Profile struct {
	UserID     uint      `gorm:"primaryKey" json:"user_id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Bio        string    `json:"bio"`
	ProfileImg string    `json:"profile_img"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
