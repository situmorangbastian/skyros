package serviceutils

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// GatewayMarkerKey is metadata attached by the REST gateways to every request
// they forward, so services can tell external traffic (which must carry a
// bearer token) from trusted service-to-service gRPC calls.
const GatewayMarkerKey = "x-skyros-gateway"

// GatewayMetadataAnnotator always marks the forwarded request as coming from a
// gateway, regardless of which headers the HTTP client sent.
func GatewayMetadataAnnotator(_ context.Context, _ *http.Request) metadata.MD {
	return metadata.Pairs(GatewayMarkerKey, "true")
}
