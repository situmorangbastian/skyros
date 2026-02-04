package serviceutils

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

func NewRestErrorHandler() runtime.ErrorHandlerFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler,
		w http.ResponseWriter, r *http.Request, err error) {

		st, ok := status.FromError(err)
		if !ok {
			st = status.New(codes.Internal, "Internal Server Error")
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, st.Err())
			return
		}

		code := st.Code()
		message := st.Message()

		needsSanitization := false
		switch st.Code() {
		case codes.InvalidArgument,
			codes.AlreadyExists,
			codes.NotFound,
			codes.Unauthenticated,
			codes.PermissionDenied:
		case codes.Unavailable:
			message = "Service Unavailable"
			needsSanitization = true
		default:
			message = "Internal Server Error"
			code = codes.Internal
			needsSanitization = true
		}

		if !needsSanitization {
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			return
		}

		newSt := status.New(code, message)

		if len(st.Details()) > 0 {
			details := make([]protoadapt.MessageV1, 0, len(st.Details()))
			for _, detail := range st.Details() {
				if protoMsg, ok := detail.(protoadapt.MessageV1); ok {
					details = append(details, protoMsg)
				}
			}

			if len(details) > 0 {
				newSt, _ = newSt.WithDetails(details...)
			}
		}

		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, newSt.Err())
	}
}
