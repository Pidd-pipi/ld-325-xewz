package handler

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
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
	data, err := h.service.List(c.GetString(constants.UserIDContextKey))
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}

func (h *PurchaseOrderHandler) AddOffer(c *gin.Context) {
	var req dto.AddPurchaseOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	data, err := h.service.AddOffer(c.GetString(constants.UserIDContextKey), req)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}

func (h *PurchaseOrderHandler) UpdateQuantity(c *gin.Context) {
	var path struct {
		ID uint `uri:"id" binding:"required"`
	}
	var req dto.UpdatePurchaseQuantityRequest
	if err := c.ShouldBindUri(&path); err != nil {
		c.Error(err)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.Error(err)
		return
	}
	data, err := h.service.UpdateQuantity(c.GetString(constants.UserIDContextKey), path.ID, req.Quantity)
	if err != nil {
		c.Error(err)
		return
	}
	success(c, data)
}
