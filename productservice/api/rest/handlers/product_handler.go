package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	restCtx "github.com/situmorangbastian/skyros/productservice/api/rest/context"
	internalErr "github.com/situmorangbastian/skyros/productservice/internal/errors"
	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
)

type productHandler struct {
	productUsecase usecase.ProductUsecase
}

// NewProductHandler init the product handler
func NewProductHandler(e *echo.Echo, g *echo.Group, productUsecase usecase.ProductUsecase) {
	if productUsecase == nil {
		panic("http: product usecase is nil")
	}

	handler := &productHandler{productUsecase}

	g.POST("/product", handler.store)

	e.GET("/product/:id", handler.get)
	e.GET("/product", handler.fetch)
}

func (h productHandler) store(c echo.Context) error {
	var product models.Product
	if err := c.Bind(&product); err != nil {
		return internalErr.ConstraintError("invalid request body")
	}

	if err := c.Validate(&product); err != nil {
		return err
	}

	res, err := h.productUsecase.Store(c.Request().Context(), product)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h productHandler) get(c echo.Context) error {
	res, err := h.productUsecase.Get(c.Request().Context(), c.Param("id"))
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

		c.SetRequest(c.Request().WithContext(restCtx.NewCustomContext(c.Request().Context(), claims)))
	}

	filter := models.ProductFilter{
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

	res, err := h.productUsecase.Fetch(c.Request().Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
