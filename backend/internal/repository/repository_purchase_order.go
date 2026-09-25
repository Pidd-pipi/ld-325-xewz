package repository

import (
	"errors"
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type PurchaseOrderRepository interface {
	AddItem(userID string, offerID uint, lockedUnitPrice float64, quantity int) (model.PurchaseOrderItem, error)
	UpdateItemQuantity(userID string, itemID uint, quantity int) (model.PurchaseOrderItem, error)
	ListItems(userID string) ([]model.PurchaseOrderItem, error)
}

type purchaseOrderRepository struct{ db *gorm.DB }

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db}
}

func (r *purchaseOrderRepository) AddItem(userID string, offerID uint, lockedUnitPrice float64, quantity int) (model.PurchaseOrderItem, error) {
	var item model.PurchaseOrderItem
	err := r.db.Transaction(func(tx *gorm.DB) error {
		order, err := findOrCreatePurchaseOrder(tx, userID)
		if err != nil {
			return err
		}
		if err := tx.Where("purchase_order_id = ? AND offer_id = ?", order.ID, offerID).First(&item).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("find purchase item: %w", err)
			}
			item = model.PurchaseOrderItem{
				PurchaseOrderID: order.ID,
				OfferID:         offerID,
				Quantity:        quantity,
				LockedUnitPrice: lockedUnitPrice,
			}
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("create purchase item: %w", err)
			}
			return nil
		}
		item.Quantity += quantity
		if err := tx.Save(&item).Error; err != nil {
			return fmt.Errorf("accumulate purchase quantity: %w", err)
		}
		return nil
	})
	return item, err
}

func (r *purchaseOrderRepository) UpdateItemQuantity(userID string, itemID uint, quantity int) (model.PurchaseOrderItem, error) {
	var item model.PurchaseOrderItem
	err := r.db.Transaction(func(tx *gorm.DB) error {
		order, err := findPurchaseOrder(tx, userID)
		if err != nil {
			return err
		}
		if err := tx.Preload("Offer.Supplier").Preload("Offer.Product").
			Where("id = ? AND purchase_order_id = ?", itemID, order.ID).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("purchase item: %w", gorm.ErrRecordNotFound)
			}
			return fmt.Errorf("find purchase item: %w", err)
		}
		item.Quantity = quantity
		if err := tx.Save(&item).Error; err != nil {
			return fmt.Errorf("update purchase quantity: %w", err)
		}
		return nil
	})
	return item, err
}

func (r *purchaseOrderRepository) ListItems(userID string) ([]model.PurchaseOrderItem, error) {
	order, err := findPurchaseOrder(r.db, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []model.PurchaseOrderItem{}, nil
	}
	if err != nil {
		return nil, err
	}
	var items []model.PurchaseOrderItem
	err = r.db.Preload("Offer.Supplier").Preload("Offer.Product").
		Where("purchase_order_id = ?", order.ID).Order("created_at ASC").Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list purchase items: %w", err)
	}
	return items, nil
}

func findOrCreatePurchaseOrder(tx *gorm.DB, userID string) (model.PurchaseOrder, error) {
	order, err := findPurchaseOrder(tx, userID)
	if err == nil {
		return order, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return order, err
	}
	order = model.PurchaseOrder{UserID: userID}
	if err := tx.Create(&order).Error; err != nil {
		return order, fmt.Errorf("create purchase order: %w", err)
	}
	return order, nil
}

func findPurchaseOrder(tx *gorm.DB, userID string) (model.PurchaseOrder, error) {
	var order model.PurchaseOrder
	if err := tx.Where("user_id = ?", userID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return order, fmt.Errorf("find purchase order: %w", gorm.ErrRecordNotFound)
		}
		return order, fmt.Errorf("find purchase order: %w", err)
	}
	return order, nil
}
