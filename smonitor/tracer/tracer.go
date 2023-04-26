package tracer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DeniesKresna/gobridge/serror"
	"github.com/DeniesKresna/gohelper/utstring"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func InitTracerProvider(url string, service string, environment string, version string) (tp *tracesdk.TracerProvider, err error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return
	}
	tp = tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.String("version", version),
		)),
	)

	otel.SetTracerProvider(tp)
	return
}

func Route(hf http.Handler) http.Handler {
	return otelhttp.NewHandler(hf, "route")
}

func SetKey(span trace.Span, key string, value interface{}) {
	valStr := fmt.Sprintf("%v", value)
	span.SetAttributes(attribute.Key(key).String(valStr))
}

func SetErr(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func SetErrx(span trace.Span, errx serror.SError) {
	if errx == nil {
		return
	}
	err := errx.GetError()
	if err != nil {
		span.RecordError(err)
	}
	span.SetStatus(codes.Error, errx.GetComment())
}

func Start(ctx context.Context, spanName string) (span trace.Span, ct context.Context) {
	spanName = utstring.CamelToSnake(spanName)
	tr := otel.Tracer(spanName)
	ct, span = tr.Start(ctx, spanName)
	return
}

func Stop(span trace.Span) {
	span.End()
}

func StopWithErr(span trace.Span, err error) {
	SetErr(span, err)
	span.End()
}

func StopWithErrx(span trace.Span, errx serror.SError) {
	SetErrx(span, errx)
	span.End()
}
