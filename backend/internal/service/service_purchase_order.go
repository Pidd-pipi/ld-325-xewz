package service

import (
	"errors"
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/gorm"
)

type PurchaseOrderService struct {
	orderRepo repository.PurchaseOrderRepository
	offerRepo repository.OfferRepository
}

func NewPurchaseOrderService(orderRepo repository.PurchaseOrderRepository, offerRepo repository.OfferRepository) *PurchaseOrderService {
	return &PurchaseOrderService{orderRepo: orderRepo, offerRepo: offerRepo}
}

func (s *PurchaseOrderService) AddItem(userID string, input dto.AddPurchaseItemRequest) (dto.PurchaseOrderView, error) {
	offer, err := s.offerRepo.GetWithDetails(input.OfferID)
	if err != nil {
		return dto.PurchaseOrderView{}, purchaseNotFoundError(constants.PurchaseOfferNotFoundMessage, err)
	}
	if err := validateOfferSelectable(offer, input.Quantity); err != nil {
		return dto.PurchaseOrderView{}, err
	}
	if _, err := s.orderRepo.AddItem(userID, offer.ID, offer.UnitPrice, input.Quantity); err != nil {
		return dto.PurchaseOrderView{}, fmt.Errorf("add purchase item: %w", err)
	}
	return s.List(userID)
}

func (s *PurchaseOrderService) UpdateItem(userID string, itemID uint, quantity int) (dto.PurchaseOrderView, error) {
	items, err := s.orderRepo.ListItems(userID)
	if err != nil {
		return dto.PurchaseOrderView{}, fmt.Errorf("load purchase items: %w", err)
	}
	var current *model.PurchaseOrderItem
	for index := range items {
		if items[index].ID == itemID {
			current = &items[index]
			break
		}
	}
	if current == nil {
		return dto.PurchaseOrderView{}, purchaseNotFoundError(constants.PurchaseItemNotFoundMessage, gorm.ErrRecordNotFound)
	}
	offer, err := s.offerRepo.GetWithDetails(current.OfferID)
	if err != nil {
		return dto.PurchaseOrderView{}, purchaseNotFoundError(constants.PurchaseOfferNotFoundMessage, err)
	}
	if reason, blocked := purchaseBlockedReason(offer); blocked {
		return dto.PurchaseOrderView{}, apperrors.NewBusinessError(constants.ErrorBusiness, constants.PurchaseBlockedAdjustmentMessage, errors.New(reason))
	}
	if quantity < offer.MOQ {
		return dto.PurchaseOrderView{}, apperrors.NewBusinessError(constants.ErrorBusiness, constants.PurchaseBelowMOQMessage, apperrors.ErrInvalidInput)
	}
	if _, err := s.orderRepo.UpdateItemQuantity(userID, itemID, quantity); err != nil {
		return dto.PurchaseOrderView{}, fmt.Errorf("update purchase item: %w", err)
	}
	return s.List(userID)
}

func (s *PurchaseOrderService) List(userID string) (dto.PurchaseOrderView, error) {
	items, err := s.orderRepo.ListItems(userID)
	if err != nil {
		return dto.PurchaseOrderView{}, fmt.Errorf("list purchase order: %w", err)
	}
	view := dto.PurchaseOrderView{Items: []dto.PurchaseOrderItemView{}}
	if len(items) > 0 {
		view.ID = items[0].PurchaseOrderID
	}
	for _, item := range items {
		line := buildPurchaseItemView(item)
		view.Items = append(view.Items, line)
		view.LockedTotal += line.LockedAmount
		view.CurrentTotal += line.CurrentAmount
		if line.Purchasable {
			view.PurchasableCount++
		} else {
			view.UnavailableCount++
		}
	}
	view.TotalDifference = view.CurrentTotal - view.LockedTotal
	return view, nil
}

func validateOfferSelectable(offer model.Offer, quantity int) error {
	if reason, blocked := purchaseBlockedReason(offer); blocked {
		message := constants.PurchaseSupplierRejectedMessage
		if reason == constants.PurchaseBlockedStock {
			message = constants.PurchaseUnavailableStockMessage
		}
		return apperrors.NewBusinessError(constants.ErrorBusiness, message, errors.New(reason))
	}
	if quantity < offer.MOQ {
		return apperrors.NewBusinessError(constants.ErrorBusiness, constants.PurchaseBelowMOQMessage, apperrors.ErrInvalidInput)
	}
	return nil
}

func purchaseBlockedReason(offer model.Offer) (string, bool) {
	if offer.Supplier.Status != constants.SupplierApproved {
		return constants.PurchaseBlockedSupplier, true
	}
	if offer.StockStatus != constants.StatusInStock {
		return constants.PurchaseBlockedStock, true
	}
	return "", false
}

func buildPurchaseItemView(item model.PurchaseOrderItem) dto.PurchaseOrderItemView {
	offer := item.Offer
	product := offer.Product
	supplier := offer.Supplier
	currentAmount := offer.UnitPrice * float64(item.Quantity)
	lockedAmount := item.LockedUnitPrice * float64(item.Quantity)
	reason, blocked := purchaseBlockedReason(offer)
	return dto.PurchaseOrderItemView{
		ID: item.ID, OfferID: offer.ID, ProductID: product.ID, ProductName: product.Name,
		Brand: product.Brand, Model: product.ModelNumber, Unit: product.Unit,
		SupplierID: supplier.ID, SupplierName: supplier.Name, SupplierStatus: supplier.Status,
		StockStatus: offer.StockStatus, MOQ: offer.MOQ, Quantity: item.Quantity,
		LockedUnitPrice: item.LockedUnitPrice, CurrentUnitPrice: offer.UnitPrice,
		PriceDifference: offer.UnitPrice - item.LockedUnitPrice,
		LockedAmount:    lockedAmount, CurrentAmount: currentAmount,
		AmountDifference: currentAmount - lockedAmount, DeliveryDays: offer.DeliveryDays,
		Freight: offer.Freight, Purchasable: !blocked, BlockedReason: reason,
	}
}

func purchaseNotFoundError(message string, cause error) error {
	return apperrors.NewBusinessError(constants.ErrorNotFound, message, cause)
}
