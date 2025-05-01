package grpc

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
	"github.com/situmorangbastian/skyros/userservice/api/validators"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

type userGrpcHandler struct {
	userUsecase    usecase.UserUsecase
	tokenSecretKey string
	validators     validators.CustomValidator
}

func NewUserGrpcServer(userUsecase usecase.UserUsecase, tokenSecretKey string, validators validators.CustomValidator) grpcService.UserServiceServer {
	return &userGrpcHandler{
		userUsecase:    userUsecase,
		tokenSecretKey: tokenSecretKey,
		validators:     validators,
	}
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
