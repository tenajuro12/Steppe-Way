package models

import (
	"gorm.io/gorm"
)

type Blog struct {
	gorm.Model
	Title    string `json:"title"`
	Content  string `json:"content"`
	UserID   uint   `json:"user_id"`
	Likes    int    `json:"likes"`
	Category string `json:"category"`
}

type Comment struct {
	gorm.Model
	Content string `json:"content"`
	BlogID  uint   `json:"blog_id"`
	UserID  uint   `json:"user_id"`
}

type BlogLike struct {
	gorm.Model
	UserID uint `json:"user_id"`
	BlogID uint `json:"blog_id"`
}
