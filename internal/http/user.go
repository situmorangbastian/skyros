package http

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	"github.com/situmorangbastian/skyros"
)

type userHandler struct {
	service        skyros.UserService
	tokenSecretKey string
}

// NewUserHandler init the user handler
func NewUserHandler(e *echo.Echo, service skyros.UserService, tokenSecretKey string) {
	if service == nil {
		panic("http: nil user service")
	}

	handler := &userHandler{
		service:        service,
		tokenSecretKey: tokenSecretKey,
	}

	e.POST("/login", handler.login)
	e.POST("/register/buyer", handler.registerBuyer)
	e.POST("/register/seller", handler.registerSeller)
}

func (h userHandler) login(c echo.Context) error {
	var user skyros.User
	if err := c.Bind(&user); err != nil {
		return skyros.ConstraintError("invalid request body")
	}

	res, err := h.service.Login(c.Request().Context(), user.Email, user.Password)
	if err != nil {
		return err
	}

	accessToken, err := generateToken(res, h.tokenSecretKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func (h userHandler) registerSeller(c echo.Context) error {
	var user skyros.User
	if err := c.Bind(&user); err != nil {
		return skyros.ConstraintError("invalid request body")
	}

	user.Type = "seller"

	if err := c.Validate(&user); err != nil {
		return err
	}

	res, err := h.service.Register(c.Request().Context(), user)
	if err != nil {
		return err
	}

	accessToken, err := generateToken(res, h.tokenSecretKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"access_token": accessToken,
	})
}

func (h userHandler) registerBuyer(c echo.Context) error {
	var user skyros.User
	if err := c.Bind(&user); err != nil {
		return skyros.ConstraintError("invalid request body")
	}

	user.Type = "buyer"

	if err := c.Validate(&user); err != nil {
		return err
	}

	res, err := h.service.Register(c.Request().Context(), user)
	if err != nil {
		return err
	}

	accessToken, err := generateToken(res, h.tokenSecretKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"access_token": accessToken,
	})
}

func generateToken(user skyros.User, secretKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["name"] = user.Name
	claims["email"] = user.Email
	claims["address"] = user.Address
	claims["type"] = user.Type
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
