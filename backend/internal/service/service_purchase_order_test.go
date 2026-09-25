package service

import (
	"errors"
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

type mockPurchaseRepository struct {
	offers       map[uint]model.Offer
	items        map[string]model.PurchaseOrderItem
	savedItem    *model.PurchaseOrderItem
	nextItemID   uint
	findNotFound bool
}

func newMockPurchaseRepository() *mockPurchaseRepository {
	return &mockPurchaseRepository{
		offers: map[uint]model.Offer{
			1: {Model: gorm.Model{ID: 1}, ProductID: 10, SupplierID: 20, UnitPrice: 100, MOQ: 2, StockStatus: constants.StatusInStock,
				Product:  model.Product{Name: "岩板", ModelNumber: "YB", Unit: "片"},
				Supplier: model.Supplier{Model: gorm.Model{ID: 20}, Name: "筑家", Status: constants.SupplierApproved}},
			2: {Model: gorm.Model{ID: 2}, ProductID: 10, SupplierID: 20, UnitPrice: 100, MOQ: 5, StockStatus: constants.StatusDiscontinued,
				Product:  model.Product{Name: "岩板"},
				Supplier: model.Supplier{Model: gorm.Model{ID: 20}, Status: constants.SupplierApproved}},
			3: {Model: gorm.Model{ID: 3}, ProductID: 10, SupplierID: 21, UnitPrice: 100, MOQ: 2, StockStatus: constants.StatusInStock,
				Product:  model.Product{Name: "岩板"},
				Supplier: model.Supplier{Model: gorm.Model{ID: 21}, Status: constants.SupplierRejected}},
		},
		items:      map[string]model.PurchaseOrderItem{},
		nextItemID: 1,
	}
}

func purchaseKey(user string, offerID uint) string { return user + ":" + string(rune(offerID)) }

func (m *mockPurchaseRepository) GetOfferDetails(id uint) (model.Offer, error) {
	offer, ok := m.offers[id]
	if !ok {
		return model.Offer{}, apperrors.ErrNotFound
	}
	return offer, nil
}
func (m *mockPurchaseRepository) ListItems(userID string) ([]model.PurchaseOrderItem, error) {
	rows := []model.PurchaseOrderItem{}
	for _, item := range m.items {
		if item.UserID == userID {
			rows = append(rows, m.attachOffer(item))
		}
	}
	return rows, nil
}
func (m *mockPurchaseRepository) FindItem(userID string, offerID uint) (model.PurchaseOrderItem, error) {
	if m.findNotFound {
		return model.PurchaseOrderItem{}, apperrors.ErrNotFound
	}
	item, ok := m.items[purchaseKey(userID, offerID)]
	if !ok {
		return model.PurchaseOrderItem{}, apperrors.ErrNotFound
	}
	return item, nil
}
func (m *mockPurchaseRepository) GetItem(userID string, itemID uint) (model.PurchaseOrderItem, error) {
	for _, item := range m.items {
		if item.UserID == userID && item.ID == itemID {
			return m.attachOffer(item), nil
		}
	}
	return model.PurchaseOrderItem{}, apperrors.ErrNotFound
}
func (m *mockPurchaseRepository) CreateItem(item model.PurchaseOrderItem) (model.PurchaseOrderItem, error) {
	item.ID = m.nextItemID
	m.nextItemID++
	m.items[purchaseKey(item.UserID, item.OfferID)] = item
	return item, nil
}
func (m *mockPurchaseRepository) SaveItem(item model.PurchaseOrderItem) (model.PurchaseOrderItem, error) {
	m.items[purchaseKey(item.UserID, item.OfferID)] = item
	m.savedItem = &item
	return item, nil
}
func (m *mockPurchaseRepository) attachOffer(item model.PurchaseOrderItem) model.PurchaseOrderItem {
	item.Offer = m.offers[item.OfferID]
	return item
}

func TestPurchaseOrderAddLocksFirstPriceAndAccumulates(t *testing.T) {
	repo := newMockPurchaseRepository()
	svc := NewPurchaseOrderService(repo)

	first, err := svc.AddOffer("buyer", dto.AddPurchaseOfferRequest{OfferID: 1, Quantity: 2})
	if err != nil {
		t.Fatalf("first add: %v", err)
	}
	if first.LockedUnitPrice != 100 || first.Quantity != 2 || !first.Purchasable {
		t.Fatalf("unexpected first item: %+v", first)
	}

	offer := repo.offers[1]
	offer.UnitPrice = 120
	repo.offers[1] = offer
	second, err := svc.AddOffer("buyer", dto.AddPurchaseOfferRequest{OfferID: 1, Quantity: 3})
	if err != nil {
		t.Fatalf("second add: %v", err)
	}
	if second.Quantity != 5 || second.LockedUnitPrice != 100 || second.CurrentPrice != 120 || second.PriceDifference != 20 {
		t.Fatalf("lock was not preserved: %+v", second)
	}
}

func TestPurchaseOrderRejectsUnavailableOffersAndMOQ(t *testing.T) {
	svc := NewPurchaseOrderService(newMockPurchaseRepository())
	cases := []struct {
		name    string
		offerID uint
		qty     int
	}{
		{name: "discontinued stock", offerID: 2, qty: 5},
		{name: "rejected supplier", offerID: 3, qty: 2},
		{name: "below moq", offerID: 1, qty: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.AddOffer("buyer", dto.AddPurchaseOfferRequest{OfferID: tc.offerID, Quantity: tc.qty})
			if !errors.Is(err, apperrors.ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
		})
	}
}

func TestPurchaseOrderUpdateStopsForUnavailableItem(t *testing.T) {
	repo := newMockPurchaseRepository()
	svc := NewPurchaseOrderService(repo)
	item := model.PurchaseOrderItem{Model: gorm.Model{ID: 1}, UserID: "buyer", OfferID: 1, Quantity: 2, LockedUnitPrice: 95}
	repo.items[purchaseKey("buyer", 1)] = item
	offer := repo.offers[1]
	offer.StockStatus = constants.StatusDiscontinued
	repo.offers[1] = offer

	_, err := svc.UpdateQuantity("buyer", item.ID, 5)
	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected quantity update to stop, got %v", err)
	}
	if repo.savedItem != nil {
		t.Fatal("unavailable item quantity was changed")
	}
}

func TestPurchaseOrderListMarksExistingItemUnpurchasable(t *testing.T) {
	repo := newMockPurchaseRepository()
	item := model.PurchaseOrderItem{Model: gorm.Model{ID: 9}, UserID: "buyer", OfferID: 1, Quantity: 4, LockedUnitPrice: 95}
	repo.items[purchaseKey("buyer", 1)] = item
	offer := repo.offers[1]
	offer.UnitPrice = 88
	offer.StockStatus = constants.StatusOutOfStock
	repo.offers[1] = offer

	order, err := NewPurchaseOrderService(repo).List("buyer")
	if err != nil {
		t.Fatalf("list order: %v", err)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(order.Items))
	}
	row := order.Items[0]
	if row.Purchasable || row.Reason != constants.PurchaseReasonStock || row.Quantity != 4 || row.LockedUnitPrice != 95 || row.CurrentPrice != 88 || row.PriceDifference != -7 {
		t.Fatalf("unexpected unavailable row: %+v", row)
	}
	if order.PurchasableCount != 0 {
		t.Fatalf("expected no purchasable items, got %d", order.PurchasableCount)
	}
}
