package dto

type AddPurchaseItemRequest struct {
	OfferID  uint `json:"offer_id" validate:"required,gt=0"`
	Quantity int  `json:"quantity" validate:"required,gt=0"`
}

type UpdatePurchaseItemRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

type PurchaseOrderItemView struct {
	ID               uint    `json:"id"`
	OfferID          uint    `json:"offer_id"`
	ProductID        uint    `json:"product_id"`
	ProductName      string  `json:"product_name"`
	Brand            string  `json:"brand"`
	Model            string  `json:"model"`
	Unit             string  `json:"unit"`
	SupplierID       uint    `json:"supplier_id"`
	SupplierName     string  `json:"supplier_name"`
	SupplierStatus   string  `json:"supplier_status"`
	StockStatus      string  `json:"stock_status"`
	MOQ              int     `json:"moq"`
	Quantity         int     `json:"quantity"`
	LockedUnitPrice  float64 `json:"locked_unit_price"`
	CurrentUnitPrice float64 `json:"current_unit_price"`
	PriceDifference  float64 `json:"price_difference"`
	LockedAmount     float64 `json:"locked_amount"`
	CurrentAmount    float64 `json:"current_amount"`
	AmountDifference float64 `json:"amount_difference"`
	DeliveryDays     int     `json:"delivery_days"`
	Freight          string  `json:"freight"`
	Purchasable      bool    `json:"purchasable"`
	BlockedReason    string  `json:"blocked_reason,omitempty"`
}

type PurchaseOrderView struct {
	ID               uint                    `json:"id"`
	Items            []PurchaseOrderItemView `json:"items"`
	LockedTotal      float64                 `json:"locked_total"`
	CurrentTotal     float64                 `json:"current_total"`
	TotalDifference  float64                 `json:"total_difference"`
	PurchasableCount int                     `json:"purchasable_count"`
	UnavailableCount int                     `json:"unavailable_count"`
}
