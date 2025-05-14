package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return errors.New("invalid scan source for StringArray")
	}

	return json.Unmarshal(bytes, s)
}

type Accommodation struct {
	gorm.Model
	Name        string      `json:"name" gorm:"not null"`
	Description string      `json:"description"`
	City        string      `json:"city" gorm:"index;not null"`
	Address     string      `json:"address" gorm:"not null"`
	Location    string      `json:"location" gorm:"index"`
	Type        string      `json:"type" gorm:"index;not null"`
	AdminID     uint        `json:"admin_id"`
	Website     string      `json:"website" gorm:"index;not null"`
	IsPublished bool        `json:"is_published" gorm:"default:false"`
	Amenities   StringArray `json:"amenities" gorm:"type:json"`

	Images    []AccommodationImage  `json:"images" gorm:"foreignKey:AccommodationID;constraint:OnDelete:CASCADE;"`
	RoomTypes []RoomType            `json:"room_types" gorm:"foreignKey:AccommodationID;constraint:OnDelete:CASCADE;"`
	Reviews   []AccommodationReview `json:"reviews" gorm:"foreignKey:AccommodationID;constraint:OnDelete:CASCADE;"`
}

type AccommodationImage struct {
	gorm.Model
	AccommodationID uint   `json:"accommodation_id" gorm:"index;not null"`
	URL             string `json:"url" gorm:"not null"`
}

type RoomType struct {
	gorm.Model
	AccommodationID uint        `json:"accommodation_id" gorm:"index;not null"`
	Name            string      `json:"name" gorm:"not null"`
	Description     string      `json:"description"`
	Price           float64     `json:"price" gorm:"not null"`
	MaxGuests       int         `json:"max_guests,MaxGuests" gorm:"not null"`
	BedType         string      `json:"bed_type,BedType"`
	Amenities       StringArray `json:"amenities,Amenities" gorm:"type:json"`

	Images []RoomImage `json:"images" gorm:"foreignKey:RoomTypeID;constraint:OnDelete:CASCADE;"`
}

type AccommodationReview struct {
	gorm.Model
	AccommodationID uint   `json:"accommodation_id" gorm:"index;not null"`
	UserID          uint   `json:"user_id" gorm:"index;not null"`
	Username        string `json:"username"`
	Rating          int    `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment         string `json:"comment"`

	Images []ReviewImage `json:"images" gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE;"`
}

type ReviewImage struct {
	gorm.Model
	ReviewID uint   `json:"review_id" gorm:"index;not null"`
	URL      string `json:"url" gorm:"not null"`
}

type RoomImage struct {
	gorm.Model
	RoomTypeID uint   `json:"room_type_id" gorm:"index;not null"`
	URL        string `json:"url" gorm:"not null"`
}
