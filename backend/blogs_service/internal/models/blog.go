package models

import (
	"gorm.io/gorm"
)

type Blog struct {
	gorm.Model
	Title    string      `json:"title"`
	Content  string      `json:"content"`
	UserID   uint        `json:"user_id"`
	Username string      `json:"username"`
	Likes    int         `json:"likes"`
	Category string      `json:"category"`
	Comments []Comment   `gorm:"foreignKey:BlogID" json:"comments"`
	Images   []BlogImage `json:"images" gorm:"foreignKey:BlogID"`
}

type BlogImage struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	BlogID uint   `json:"blog_id"`
	URL    string `json:"url"`
}

type Comment struct {
	gorm.Model
	Content  string         `json:"content"`
	BlogID   uint           `json:"blog_id"`
	UserID   uint           `json:"user_id"`
	Username string         `json:"username"`
	Images   []CommentImage `json:"images" gorm:"foreignKey:CommentID"`
}

type CommentImage struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	CommentID uint   `json:"comment_id"`
	URL       string `json:"url"`
}

type BlogLike struct {
	gorm.Model
	UserID   uint   `json:"user_id"`
	BlogID   uint   `json:"blog_id"`
	Username string `json:"username"`
}
