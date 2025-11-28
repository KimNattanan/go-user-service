package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           string `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Name         string `json:"name"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	ImageURL     string `json:"image_url"`

	Preference Preference `gorm:"foreignKey:UserID;constraint:onDelete:CASCADE"`
}

func (u *User) BeforeCreate(db *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}

func (u *User) AfterCreate(db *gorm.DB) (err error) {
	preference := &Preference{
		UserID: u.ID,
		Theme:  "light",
	}
	err = db.Create(preference).Error
	return
}
