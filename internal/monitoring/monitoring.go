package monitoring

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.22.0"
)

type ShutdownFunc func()

func MustSetup(ctx context.Context, debug bool, version string) ShutdownFunc {
	var exp metric.Exporter
	var err error
	var res *resource.Resource
	// If there is no indicator that we want metrics to be submitted somewhere,
	// the metrics should be sent somewhere, the default meterprovider will be
	// used which does not do anything:
	wantMetrics := false
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "OTEL_EXPORTER_OTLP_ENDPOINT") {
			wantMetrics = true
			break
		}
	}
	if !wantMetrics {
		// If we are in debug mode, we still want to send metrics to stderr
		if debug {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")

			exp, err = stdoutmetric.New(
				stdoutmetric.WithEncoder(enc),
				stdoutmetric.WithoutTimestamps(),
			)
			if err != nil {
				panic(err)
			}
		} else {
			return func() {}
		}
	} else {
		exp, err = otlpmetrichttp.New(ctx)
		if err != nil {
			panic(err)
		}
	}

	res = resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceName("podcast-archiver"),
		semconv.ServiceVersion("0.1.0"),
	)
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exp)))
	otel.SetMeterProvider(meterProvider)
	return func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
}
