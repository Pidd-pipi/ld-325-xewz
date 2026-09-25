package handler

import (
	"strconv"

	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PurchaseOrderHandler struct {
	service  *service.PurchaseOrderService
	validate *validator.Validate
}

func NewPurchaseOrderHandler(s *service.PurchaseOrderService, v *validator.Validate) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{service: s, validate: v}
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	data, err := h.service.List(c.GetString("user_id"))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}

func (h *PurchaseOrderHandler) AddItem(c *gin.Context) {
	var req dto.AddPurchaseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	data, err := h.service.AddItem(c.GetString("user_id"), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}

func (h *PurchaseOrderHandler) UpdateItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || itemID == 0 {
		c.Error(apperrors.ErrInvalidInput)
		return
	}
	var req dto.UpdatePurchaseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	data, err := h.service.UpdateItem(c.GetString("user_id"), uint(itemID), req.Quantity)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}
