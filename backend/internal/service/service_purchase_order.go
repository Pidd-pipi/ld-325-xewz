package service

import (
	"errors"
	"fmt"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

type PurchaseOrderService struct {
	repo repository.PurchaseOrderRepository
}

func NewPurchaseOrderService(repo repository.PurchaseOrderRepository) *PurchaseOrderService {
	return &PurchaseOrderService{repo: repo}
}

func (s *PurchaseOrderService) AddOffer(userID string, input dto.AddPurchaseOfferRequest) (dto.PurchaseOrderItemResponse, error) {
	offer, err := s.repo.GetOfferDetails(input.OfferID)
	if err != nil {
		return dto.PurchaseOrderItemResponse{}, fmt.Errorf("get offer for purchase: %w", err)
	}
	if reason, available := purchaseAvailability(offer); !available {
		return dto.PurchaseOrderItemResponse{}, purchaseValidationError(reason)
	}
	if input.Quantity < offer.MOQ {
		return dto.PurchaseOrderItemResponse{}, apperrors.NewValidationError(constants.PurchaseErrorBelowMOQ)
	}

	item, err := s.repo.FindItem(userID, offer.ID)
	if errors.Is(err, apperrors.ErrNotFound) {
		item = model.PurchaseOrderItem{
			UserID:          userID,
			OfferID:         offer.ID,
			Quantity:        input.Quantity,
			LockedUnitPrice: offer.UnitPrice,
		}
		item, err = s.repo.CreateItem(item)
		if err != nil {
			return dto.PurchaseOrderItemResponse{}, fmt.Errorf("create purchase item: %w", err)
		}
	} else if err != nil {
		return dto.PurchaseOrderItemResponse{}, fmt.Errorf("find existing purchase item: %w", err)
	} else {
		if input.Quantity < offer.MOQ || item.Quantity+input.Quantity < offer.MOQ {
			return dto.PurchaseOrderItemResponse{}, apperrors.NewValidationError(constants.PurchaseErrorBelowMOQ)
		}
		item.Quantity += input.Quantity
		item, err = s.repo.SaveItem(item)
		if err != nil {
			return dto.PurchaseOrderItemResponse{}, fmt.Errorf("accumulate purchase quantity: %w", err)
		}
	}
	saved, err := s.repo.GetItem(userID, item.ID)
	if err != nil {
		return dto.PurchaseOrderItemResponse{}, fmt.Errorf("reload purchase item: %w", err)
	}
	return mapPurchaseOrderItem(saved, offer), nil
}

func (s *PurchaseOrderService) UpdateQuantity(userID string, itemID uint, quantity int) (dto.PurchaseOrderItemResponse, error) {
	item, err := s.repo.GetItem(userID, itemID)
	if err != nil {
		return dto.PurchaseOrderItemResponse{}, fmt.Errorf("get purchase item: %w", err)
	}
	if reason, available := purchaseAvailability(item.Offer); !available {
		return dto.PurchaseOrderItemResponse{}, purchaseValidationError(reason)
	}
	if quantity < item.Offer.MOQ {
		return dto.PurchaseOrderItemResponse{}, apperrors.NewValidationError(constants.PurchaseErrorBelowMOQ)
	}
	item.Quantity = quantity
	item, err = s.repo.SaveItem(item)
	if err != nil {
		return dto.PurchaseOrderItemResponse{}, fmt.Errorf("update purchase quantity: %w", err)
	}
	return mapPurchaseOrderItem(item, item.Offer), nil
}

func (s *PurchaseOrderService) List(userID string) (dto.PurchaseOrderResponse, error) {
	items, err := s.repo.ListItems(userID)
	if err != nil {
		return dto.PurchaseOrderResponse{}, fmt.Errorf("list purchase order: %w", err)
	}
	rows := make([]dto.PurchaseOrderItemResponse, 0, len(items))
	response := dto.PurchaseOrderResponse{Items: rows}
	for _, item := range items {
		row := mapPurchaseOrderItem(item, item.Offer)
		response.Items = append(response.Items, row)
		response.TotalLockedAmount += row.LockedUnitPrice * float64(row.Quantity)
		response.TotalCurrentAmount += row.CurrentPrice * float64(row.Quantity)
		if row.Purchasable {
			response.PurchasableCount++
		}
	}
	response.TotalDifference = response.TotalCurrentAmount - response.TotalLockedAmount
	return response, nil
}

func purchaseAvailability(offer model.Offer) (string, bool) {
	if offer.StockStatus != constants.StatusInStock {
		return constants.PurchaseReasonStock, false
	}
	if offer.Supplier.Status != constants.SupplierApproved {
		return constants.PurchaseReasonSupplier, false
	}
	return constants.PurchaseReasonAvailable, true
}

func purchaseValidationError(reason string) error {
	if reason == constants.PurchaseReasonSupplier {
		return apperrors.NewValidationError(constants.PurchaseErrorSupplierBlocked)
	}
	return apperrors.NewValidationError(constants.PurchaseErrorOfferUnavailable)
}

func mapPurchaseOrderItem(item model.PurchaseOrderItem, offer model.Offer) dto.PurchaseOrderItemResponse {
	reason, purchasable := purchaseAvailability(offer)
	lockedPrice := item.LockedUnitPrice
	if lockedPrice == 0 {
		lockedPrice = offer.UnitPrice
	}
	return dto.PurchaseOrderItemResponse{
		ID:              item.ID,
		OfferID:         offer.ID,
		ProductID:       offer.ProductID,
		ProductName:     offer.Product.Name,
		ProductModel:    offer.Product.ModelNumber,
		Unit:            offer.Product.Unit,
		SupplierID:      offer.SupplierID,
		SupplierName:    offer.Supplier.Name,
		Quantity:        item.Quantity,
		MOQ:             offer.MOQ,
		LockedUnitPrice: lockedPrice,
		CurrentPrice:    offer.UnitPrice,
		PriceDifference: offer.UnitPrice - lockedPrice,
		StockStatus:     offer.StockStatus,
		SupplierStatus:  offer.Supplier.Status,
		Purchasable:     purchasable,
		Reason:          reason,
	}
}
