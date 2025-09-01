package telemetry

import (
	"context"
	"net/http"
	"time"

	"github.com/honeycombio/otel-config-go/otelconfig"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func ConfigureOpenTelemetry() (func(), error) {
	// TODO explicit config
	return otelconfig.ConfigureOpenTelemetry()
}

func NewHandler(handler http.Handler, operationName string) http.Handler {
	return otelhttp.NewHandler(handler, operationName)
}

func ForceFlushTraces() error {
	tp := otel.GetTracerProvider()
	if sdkTp, ok := tp.(*trace.TracerProvider); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return sdkTp.ForceFlush(ctx)
	}
	return nil
}
