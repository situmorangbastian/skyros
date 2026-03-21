package serviceutils

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewRestErrorHandler() runtime.ErrorHandlerFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler,
		w http.ResponseWriter, r *http.Request, err error) {

		st, ok := status.FromError(err)
		if !ok {
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r,
				status.New(codes.Internal, "Internal Server Error").Err(),
			)
			return
		}

		switch st.Code() {
		case codes.InvalidArgument,
			codes.AlreadyExists,
			codes.NotFound,
			codes.Unauthenticated,
			codes.PermissionDenied:
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			return
		case codes.Unavailable:
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r,
				status.New(codes.Unavailable, "Service Unavailable").Err(),
			)
			return
		default:
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r,
				status.New(codes.Internal, "Internal Server Error").Err(),
			)
		}
	}
}
