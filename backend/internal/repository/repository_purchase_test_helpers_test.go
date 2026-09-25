package repository

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPurchaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func createPurchaseOffer(t *testing.T, db *gorm.DB, stockStatus, supplierStatus string, price float64) model.Offer {
	t.Helper()
	category := model.Category{Name: "瓷砖"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	product := model.Product{Name: "测试岩板", ModelNumber: "TEST-1", Unit: "片", CategoryID: category.ID}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	supplier := model.Supplier{Name: "测试供应商", Status: supplierStatus}
	if err := db.Create(&supplier).Error; err != nil {
		t.Fatal(err)
	}
	offer := model.Offer{ProductID: product.ID, SupplierID: supplier.ID, UnitPrice: price, MOQ: 2, StockStatus: stockStatus}
	if err := db.Create(&offer).Error; err != nil {
		t.Fatal(err)
	}
	return offer
}
