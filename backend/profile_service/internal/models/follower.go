package models

import "time"

type Follow struct {
	FollowerID uint      `gorm:"primaryKey" json:"follower_id"`
	FolloweeID uint      `gorm:"primaryKey" json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}
