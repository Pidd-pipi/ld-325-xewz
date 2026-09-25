package model

import "gorm.io/gorm"

type PurchaseOrder struct {
	gorm.Model
	UserID string `gorm:"uniqueIndex;not null"`
	Items  []PurchaseOrderItem
}

type PurchaseOrderItem struct {
	gorm.Model
	PurchaseOrderID uint `gorm:"uniqueIndex:idx_purchase_order_offer"`
	OfferID         uint `gorm:"uniqueIndex:idx_purchase_order_offer"`
	Quantity        int
	LockedUnitPrice float64
	Offer           Offer `gorm:"foreignKey:OfferID"`
}
