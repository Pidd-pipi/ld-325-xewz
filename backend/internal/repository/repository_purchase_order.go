package repository

import (
	"errors"
	"fmt"

	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type PurchaseOrderRepository interface {
	GetOfferDetails(id uint) (model.Offer, error)
	ListItems(userID string) ([]model.PurchaseOrderItem, error)
	FindItem(userID string, offerID uint) (model.PurchaseOrderItem, error)
	GetItem(userID string, itemID uint) (model.PurchaseOrderItem, error)
	CreateItem(item model.PurchaseOrderItem) (model.PurchaseOrderItem, error)
	SaveItem(item model.PurchaseOrderItem) (model.PurchaseOrderItem, error)
}

type purchaseOrderRepository struct{ db *gorm.DB }

func NewPurchaseOrderRepository(db *gorm.DB) PurchaseOrderRepository {
	return &purchaseOrderRepository{db: db}
}

func (r *purchaseOrderRepository) GetOfferDetails(id uint) (model.Offer, error) {
	var offer model.Offer
	err := r.db.Preload("Product").Preload("Supplier").First(&offer, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return offer, apperrors.ErrNotFound
	}
	if err != nil {
		return offer, fmt.Errorf("get purchase offer: %w", err)
	}
	return offer, nil
}

func (r *purchaseOrderRepository) ListItems(userID string) ([]model.PurchaseOrderItem, error) {
	var items []model.PurchaseOrderItem
	err := r.db.
		Preload("Offer.Product").
		Preload("Offer.Supplier").
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("list purchase items: %w", err)
	}
	return items, nil
}

func (r *purchaseOrderRepository) FindItem(userID string, offerID uint) (model.PurchaseOrderItem, error) {
	var item model.PurchaseOrderItem
	err := r.db.Where("user_id = ? AND offer_id = ?", userID, offerID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, apperrors.ErrNotFound
	}
	if err != nil {
		return item, fmt.Errorf("find purchase item: %w", err)
	}
	return item, nil
}

func (r *purchaseOrderRepository) GetItem(userID string, itemID uint) (model.PurchaseOrderItem, error) {
	var item model.PurchaseOrderItem
	err := r.db.
		Preload("Offer.Product").
		Preload("Offer.Supplier").
		Where("user_id = ?", userID).
		First(&item, itemID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, apperrors.ErrNotFound
	}
	if err != nil {
		return item, fmt.Errorf("get purchase item: %w", err)
	}
	return item, nil
}

func (r *purchaseOrderRepository) CreateItem(item model.PurchaseOrderItem) (model.PurchaseOrderItem, error) {
	if err := r.db.Create(&item).Error; err != nil {
		return item, fmt.Errorf("create purchase item: %w", err)
	}
	return item, nil
}

func (r *purchaseOrderRepository) SaveItem(item model.PurchaseOrderItem) (model.PurchaseOrderItem, error) {
	if err := r.db.Save(&item).Error; err != nil {
		return item, fmt.Errorf("save purchase item: %w", err)
	}
	return item, nil
}
