package middleware

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cstmErrs "github.com/situmorangbastian/skyros/userservice/internal/errors"
)

func ErrorHandlingInterceptor(log *logrus.Entry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		err = errors.Cause(err)
		switch err.(type) {
		case cstmErrs.ConstraintError:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case cstmErrs.NotFoundError:
			return nil, status.Error(codes.NotFound, err.Error())
		case cstmErrs.ConflictError:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			log.Error(err)
			return nil, status.Error(codes.Internal, "Internal Server Error")
		}
	}
}
