package models

import (
	"gorm.io/gorm"
	"time"
)

type Plan struct {
	gorm.Model
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	UserID      uint      `json:"user_id"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	City        string    `json:"city"`
}

type PlanItem struct {
	gorm.Model
	PlanID       uint      `json:"plan_id"`
	ItemType     string    `json:"item_type"`
	ItemID       uint      `json:"item_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	Address      string    `json:"address"`
	ScheduledFor time.Time `json:"scheduled_for"`
	Duration     int       `json:"duration"`
	OrderIndex   int       `json:"order_index"`
	Notes        string    `json:"notes"`
}

type PlanTemplate struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Duration    int    `json:"duration"`
	Category    string `json:"category"`
	IsPublic    bool   `json:"is_public" gorm:"default:true"`
}

type TemplateItem struct {
	gorm.Model
	TemplateID  uint   `json:"template_id"`
	ItemType    string `json:"item_type"`
	ItemID      uint   `json:"item_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	DayNumber   int    `json:"day_number"`
	OrderInDay  int    `json:"order_in_day"`
	Duration    int    `json:"duration"`
	Recommended bool   `json:"recommended"`
}
