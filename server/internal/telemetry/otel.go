package telemetry

import (
	"net/http"

	"github.com/honeycombio/otel-config-go/otelconfig"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func ConfigureOpenTelemetry() (func(), error) {
	// TODO explicit config
	return otelconfig.ConfigureOpenTelemetry()
}

func NewHandler(handler http.Handler, operationName string) http.Handler {
	return otelhttp.NewHandler(handler, operationName)
}
