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
			st := status.New(codes.Internal, "Internal Server Error")
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, st.Err())
			return
		}

		code := st.Code()
		message := st.Message()

		switch st.Code() {
		case codes.InvalidArgument, codes.AlreadyExists, codes.NotFound, codes.Unauthenticated:
		case codes.Unavailable:
			message = "Service Unavailable"
		default:
			message = "Internal Server Error"
			code = codes.Internal
		}

		st = status.New(code, message)
		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, st.Err())
	}
}
