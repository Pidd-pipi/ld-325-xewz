package service

import (
	"strings"
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPurchaseOrderTestDB(t *testing.T) (*gorm.DB, model.Offer) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	supplier := model.Supplier{Name: "测试供应商", Status: constants.SupplierApproved}
	product := model.Product{Name: "测试瓷砖", ModelNumber: "TEST-1", Unit: "片"}
	if err := db.Create(&supplier).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	offer := model.Offer{
		ProductID: product.ID, SupplierID: supplier.ID, UnitPrice: 100,
		MOQ: 5, DeliveryDays: 3, StockStatus: constants.StatusInStock,
	}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatal(err)
	}
	return db, offer
}

func TestPurchaseOrderLocksPriceAndAccumulatesQuantity(t *testing.T) {
	db, offer := newPurchaseOrderTestDB(t)
	service := NewPurchaseOrderService(repository.NewPurchaseOrderRepository(db), repository.NewOfferRepository(db))

	if _, err := service.AddItem("buyer-1", dto.AddPurchaseItemRequest{OfferID: offer.ID, Quantity: 6}); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := db.Model(&model.Offer{}).Where("id = ?", offer.ID).Update("unit_price", 110).Error; err != nil {
		t.Fatal(err)
	}
	order, err := service.AddItem("buyer-1", dto.AddPurchaseItemRequest{OfferID: offer.ID, Quantity: 5})
	if err != nil {
		t.Fatalf("second add: %v", err)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected one accumulated item, got %d", len(order.Items))
	}
	line := order.Items[0]
	if line.Quantity != 11 || line.LockedUnitPrice != 100 || line.CurrentUnitPrice != 110 {
		t.Fatalf("unexpected locked item: %+v", line)
	}
	if line.PriceDifference != 10 || line.AmountDifference != 110 {
		t.Fatalf("unexpected difference: price=%f amount=%f", line.PriceDifference, line.AmountDifference)
	}
}

func TestPurchaseOrderRejectsUnqualifiedUnavailableAndBelowMOQ(t *testing.T) {
	db, offer := newPurchaseOrderTestDB(t)
	service := NewPurchaseOrderService(repository.NewPurchaseOrderRepository(db), repository.NewOfferRepository(db))

	_, err := service.AddItem("buyer-2", dto.AddPurchaseItemRequest{OfferID: offer.ID, Quantity: offer.MOQ - 1})
	if err == nil || !strings.Contains(err.Error(), constants.PurchaseBelowMOQMessage) {
		t.Fatalf("expected MOQ error, got %v", err)
	}

	if err := db.Model(&model.Offer{}).Where("id = ?", offer.ID).Update("stock_status", constants.StatusDiscontinued).Error; err != nil {
		t.Fatal(err)
	}
	_, err = service.AddItem("buyer-2", dto.AddPurchaseItemRequest{OfferID: offer.ID, Quantity: offer.MOQ})
	if err == nil || !strings.Contains(err.Error(), constants.PurchaseUnavailableStockMessage) {
		t.Fatalf("expected unavailable stock error, got %v", err)
	}
}

func TestPurchaseOrderStopsQuantityChangesWhenItemBecomesUnavailable(t *testing.T) {
	db, offer := newPurchaseOrderTestDB(t)
	service := NewPurchaseOrderService(repository.NewPurchaseOrderRepository(db), repository.NewOfferRepository(db))
	order, err := service.AddItem("buyer-3", dto.AddPurchaseItemRequest{OfferID: offer.ID, Quantity: offer.MOQ})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&model.Supplier{}).Where("id = ?", offer.SupplierID).Update("status", constants.SupplierRejected).Error; err != nil {
		t.Fatal(err)
	}
	_, err = service.UpdateItem("buyer-3", order.Items[0].ID, 8)
	if err == nil || !strings.Contains(err.Error(), constants.PurchaseBlockedAdjustmentMessage) {
		t.Fatalf("expected blocked adjustment error, got %v", err)
	}
	order, err = service.List("buyer-3")
	if err != nil {
		t.Fatal(err)
	}
	line := order.Items[0]
	if line.Purchasable || line.Quantity != offer.MOQ || line.LockedUnitPrice != 100 || line.BlockedReason != constants.PurchaseBlockedSupplier {
		t.Fatalf("locked item should remain unavailable: %+v", line)
	}
}
