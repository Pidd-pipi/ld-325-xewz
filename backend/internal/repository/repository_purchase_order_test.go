package repository

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

func TestPurchaseOrderRepositoryAccumulatesUniqueUserOfferItem(t *testing.T) {
	db := newPurchaseTestDB(t)
	offer := createPurchaseOffer(t, db, constants.StatusInStock, constants.SupplierApproved, 100)
	repo := NewPurchaseOrderRepository(db)

	first := model.PurchaseOrderItem{UserID: "buyer-a", OfferID: offer.ID, Quantity: 2, LockedUnitPrice: 100}
	if _, err := repo.CreateItem(first); err != nil {
		t.Fatalf("create first item: %v", err)
	}
	second := model.PurchaseOrderItem{UserID: "buyer-b", OfferID: offer.ID, Quantity: 3, LockedUnitPrice: 100}
	if _, err := repo.CreateItem(second); err != nil {
		t.Fatalf("create second user item: %v", err)
	}
	if err := db.Create(&model.PurchaseOrderItem{UserID: "buyer-a", OfferID: offer.ID, Quantity: 1, LockedUnitPrice: 99}).Error; err == nil {
		t.Fatal("expected duplicate user/offer item to be rejected")
	}

	rows, err := repo.ListItems("buyer-a")
	if err != nil || len(rows) != 1 {
		t.Fatalf("list items: rows=%d err=%v", len(rows), err)
	}
	if rows[0].Offer.Product.Name != "测试岩板" || rows[0].Offer.Supplier.Name != "测试供应商" {
		t.Fatalf("expected product and supplier preload, got %+v", rows[0])
	}
}
