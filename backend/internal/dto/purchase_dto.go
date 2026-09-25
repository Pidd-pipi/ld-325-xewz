package dto

type AddPurchaseOfferRequest struct {
	OfferID  uint `json:"offer_id" validate:"required,gt=0"`
	Quantity int  `json:"quantity" validate:"required,gt=0"`
}

type UpdatePurchaseQuantityRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

type PurchaseOrderItemResponse struct {
	ID              uint    `json:"id"`
	OfferID         uint    `json:"offer_id"`
	ProductID       uint    `json:"product_id"`
	ProductName     string  `json:"product_name"`
	ProductModel    string  `json:"product_model"`
	Unit            string  `json:"unit"`
	SupplierID      uint    `json:"supplier_id"`
	SupplierName    string  `json:"supplier_name"`
	Quantity        int     `json:"quantity"`
	MOQ             int     `json:"moq"`
	LockedUnitPrice float64 `json:"locked_unit_price"`
	CurrentPrice    float64 `json:"current_price"`
	PriceDifference float64 `json:"price_difference"`
	StockStatus     string  `json:"stock_status"`
	SupplierStatus  string  `json:"supplier_status"`
	Purchasable     bool    `json:"purchasable"`
	Reason          string  `json:"reason"`
}

type PurchaseOrderResponse struct {
	Items              []PurchaseOrderItemResponse `json:"items"`
	TotalLockedAmount  float64                     `json:"total_locked_amount"`
	TotalCurrentAmount float64                     `json:"total_current_amount"`
	TotalDifference    float64                     `json:"total_difference"`
	PurchasableCount   int                         `json:"purchasable_count"`
}
