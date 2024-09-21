package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
	internalErr "github.com/situmorangbastian/skyros/orderservice/internal/errors"
	"github.com/situmorangbastian/skyros/orderservice/internal/usecase"
)

const (
	orderStatusAccept = iota + 1
)

type orderHandler struct {
	orderUsecase usecase.OrderUsecase
}

// NewOrderHandler init the order handler
func NewOrderHandler(g *echo.Group, orderUsecase usecase.OrderUsecase) {
	if orderUsecase == nil {
		panic("http: order usecase is nil")
	}

	handler := &orderHandler{orderUsecase}

	g.POST("/order", handler.store)

	g.GET("/order/:id", handler.get)
	g.GET("/order", handler.fetch)

	g.PATCH("/order/:id", handler.patchStatus)
}

func (h orderHandler) store(c echo.Context) error {
	var order models.Order
	if err := c.Bind(&order); err != nil {
		return internalErr.ConstraintError("invalid request body")
	}

	if err := c.Validate(&order); err != nil {
		return err
	}

	res, err := h.orderUsecase.Store(c.Request().Context(), order)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h orderHandler) get(c echo.Context) error {
	res, err := h.orderUsecase.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h orderHandler) fetch(c echo.Context) error {
	filter := models.Filter{
		Search:   c.QueryParam("search"),
		PageSize: 20,
	}

	if c.QueryParam("pagesize") != "" {
		pagesize, err := strconv.Atoi(c.QueryParam("pagesize"))
		if err != nil {
			return internalErr.ConstraintError("invalid pagesize")
		}

		filter.PageSize = pagesize
	}

	if c.QueryParam("page") != "" {
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			return internalErr.ConstraintError("invalid page")
		}

		filter.Page = page
	}

	res, err := h.orderUsecase.Fetch(c.Request().Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h orderHandler) patchStatus(c echo.Context) error {
	type patchRequest struct {
		Status string `json:"status" validate:"required"`
	}

	var request patchRequest
	if err := c.Bind(&request); err != nil {
		return internalErr.ConstraintError("invalid request body")
	}

	if err := c.Validate(&request); err != nil {
		return err
	}

	if request.Status != "accept" {
		return internalErr.ConstraintError("unsupported status")
	}

	err := h.orderUsecase.PatchStatus(c.Request().Context(), c.Param("id"), orderStatusAccept)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
