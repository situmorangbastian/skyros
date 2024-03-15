package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/situmorangbastian/skyros/orderservice"
)

const (
	orderStatusAccept = iota + 1
)

type orderHandler struct {
	service orderservice.OrderService
}

// NewOrderHandler init the order handler
func NewOrderHandler(g *echo.Group, service orderservice.OrderService) {
	if service == nil {
		panic("http: nil product service")
	}

	handler := &orderHandler{service}

	g.POST("/order", handler.store)

	g.GET("/order/:id", handler.get)
	g.GET("/order", handler.fetch)

	g.PATCH("/order/:id", handler.patchStatus)
}

func (h orderHandler) store(c echo.Context) error {
	var order orderservice.Order
	if err := c.Bind(&order); err != nil {
		return orderservice.ConstraintError("invalid request body")
	}

	if err := c.Validate(&order); err != nil {
		return err
	}

	res, err := h.service.Store(c.Request().Context(), order)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h orderHandler) get(c echo.Context) error {
	res, err := h.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h orderHandler) fetch(c echo.Context) error {
	filter := orderservice.Filter{
		Cursor: c.QueryParam("cursor"),
		Search: c.QueryParam("search"),
		Num:    20,
	}

	if c.QueryParam("num") != "" {
		num, err := strconv.Atoi(c.QueryParam("num"))
		if err != nil {
			return orderservice.ConstraintError("invalid num")
		}

		filter.Num = num
	}

	res, cursor, err := h.service.Fetch(c.Request().Context(), filter)
	if err != nil {
		return err
	}

	c.Response().Header().Set(`X-Cursor`, cursor)
	return c.JSON(http.StatusOK, res)
}

func (h orderHandler) patchStatus(c echo.Context) error {
	type patchRequest struct {
		Status string `json:"status" validate:"required"`
	}

	var request patchRequest
	if err := c.Bind(&request); err != nil {
		return orderservice.ConstraintError("invalid request body")
	}

	if err := c.Validate(&request); err != nil {
		return err
	}

	if request.Status != "accept" {
		return orderservice.ConstraintError("unsupported status")
	}

	err := h.service.PatchStatus(c.Request().Context(), c.Param("id"), orderStatusAccept)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
