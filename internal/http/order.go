package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/situmorangbastian/skyros"
)

type orderHandler struct {
	service skyros.OrderService
}

// NewOrderHandler init the order handler
func NewOrderHandler(e *echo.Echo, service skyros.OrderService) {
	if service == nil {
		panic("http: nil product service")
	}

	handler := &orderHandler{service}

	e.POST("/order", handler.store)

	e.GET("/order/:id", handler.get)
	e.GET("/order", handler.fetch)

	e.PATCH("/order/:id", handler.patchStatus)
}

func (h orderHandler) store(c echo.Context) error {
	var order skyros.Order
	if err := c.Bind(&order); err != nil {
		return skyros.ConstraintError("invalid request body")
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
	filter := skyros.Filter{
		Cursor: c.QueryParam("cursor"),
		Search: c.QueryParam("search"),
		Num:    20,
	}

	if c.QueryParam("num") != "" {
		num, err := strconv.Atoi(c.QueryParam("num"))
		if err != nil {
			return skyros.ConstraintError("invalid num")
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
		return skyros.ConstraintError("invalid request body")
	}

	if err := c.Validate(&request); err != nil {
		return err
	}

	if request.Status != "accept" {
		return skyros.ConstraintError("unsupported status")
	}

	err := h.service.Accept(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
