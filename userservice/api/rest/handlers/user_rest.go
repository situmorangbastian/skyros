package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	customErrors "github.com/situmorangbastian/skyros/userservice/internal/errors"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

type userRestHandler struct {
	userUsecase    usecase.UserUsecase
	tokenSecretKey string
}

func NewUserHandler(e *echo.Echo, userUsecase usecase.UserUsecase, tokenSecretKey string) {
	if userUsecase == nil {
		panic("http: user usecase is nil")
	}

	handler := &userRestHandler{
		userUsecase:    userUsecase,
		tokenSecretKey: tokenSecretKey,
	}

	e.POST("/login", handler.login)
	e.POST("/register/buyer", handler.registerBuyer)
	e.POST("/register/seller", handler.registerSeller)
}

func (h *userRestHandler) login(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return customErrors.ConstraintError("invalid request body")
	}

	res, err := h.userUsecase.Login(c.Request().Context(), user.Email, user.Password)
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

func (h *userRestHandler) registerSeller(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return customErrors.ConstraintError("invalid request body")
	}

	user.Type = models.UserSellerType

	if err := c.Validate(&user); err != nil {
		return err
	}

	res, err := h.userUsecase.Register(c.Request().Context(), user)
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

func (h *userRestHandler) registerBuyer(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return customErrors.ConstraintError("invalid request body")
	}

	user.Type = models.UserBuyerType

	if err := c.Validate(&user); err != nil {
		return err
	}

	res, err := h.userUsecase.Register(c.Request().Context(), user)
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

func generateToken(user models.User, secretKey string) (string, error) {
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
