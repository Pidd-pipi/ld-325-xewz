package model

import "gorm.io/gorm"

type PurchaseOrderItem struct {
	gorm.Model
	UserID          string `gorm:"uniqueIndex:idx_purchase_user_offer;not null"`
	OfferID         uint   `gorm:"uniqueIndex:idx_purchase_user_offer;not null"`
	Offer           Offer
	Quantity        int     `gorm:"not null"`
	LockedUnitPrice float64 `gorm:"not null"`
}
