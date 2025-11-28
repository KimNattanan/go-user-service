package entity

type Preference struct {
	UserID string `gorm:"type:uuid;primaryKey" json:"user_id"`
	Theme  string `gorm:"type:varchar(50)" json:"theme"`
}
