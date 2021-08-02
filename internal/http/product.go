package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/situmorangbastian/skyros"
)

type productHandler struct {
	service skyros.ProductService
}

// NewProductHandler init the product handler
func NewProductHandler(e *echo.Echo, g *echo.Group, service skyros.ProductService) {
	if service == nil {
		panic("http: nil product service")
	}

	handler := &productHandler{service}

	g.POST("/product", handler.store)

	e.GET("/product/:id", handler.get)
	e.GET("/product", handler.fetch)
}

func (h productHandler) store(c echo.Context) error {
	var product skyros.Product
	if err := c.Bind(&product); err != nil {
		return skyros.ConstraintError("invalid request body")
	}

	if err := c.Validate(&product); err != nil {
		return err
	}

	res, err := h.service.Store(c.Request().Context(), product)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h productHandler) get(c echo.Context) error {
	res, err := h.service.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h productHandler) fetch(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")

	if tokenString != "" {
		tokenString = strings.Replace(tokenString, "Bearer ", "", -1)
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return "skyros-secret", nil
		})

		claims := token.Claims.(jwt.MapClaims)
		id := claims["id"].(string)
		address := claims["address"].(string)
		name := claims["name"].(string)
		email := claims["email"].(string)
		type_ := claims["type"].(string)

		validUser := skyros.User{
			ID:      id,
			Address: address,
			Name:    name,
			Email:   email,
			Type:    type_,
		}

		c.SetRequest(c.Request().WithContext(skyros.NewCustomContext(c.Request().Context(), validUser)))
	}

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
