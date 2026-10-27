package model

type Product struct {
	ID    int64   `gorm:"primaryKey" json:"id"`
	Name  string  `gorm:"size:100;not null" json:"name"`
	Price float64 `gorm:"not null" json:"price"`
}
