package main

import (
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

func main() {

	for {
		// Sample configuration for testing. Use constant sampling to sample every trace
		// and enable LogSpan to log every span via configured Logger.
		cfg := jaegercfg.Configuration{
			ServiceName: "sdurc",
			Sampler: &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
			Reporter: &jaegercfg.ReporterConfig{
				LogSpans: true,
			},
		}

		// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
		// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
		// frameworks.
		jLogger := jaegerlog.StdLogger
		jMetricsFactory := metrics.NullFactory

		// Initialize tracer with a logger and a metrics factory
		tracer, _, _ := cfg.NewTracer(
			jaegercfg.Logger(jLogger),
			jaegercfg.Metrics(jMetricsFactory),
		)
		// Set the singleton opentracing.Tracer with the Jaeger tracer.
		opentracing.SetGlobalTracer(tracer)

		clientSpan := tracer.StartSpan("call cruds")
		//defer clientSpan.Finish()

		url := "http://tracing.loc/cruds/1"
		req, _ := http.NewRequest("GET", url, nil)

		// Set some tags on the clientSpan to annotate that it's the client span. The additional HTTP tags are useful for debugging purposes.
		ext.SpanKindRPCClient.Set(clientSpan)
		ext.HTTPUrl.Set(clientSpan, url)
		ext.HTTPMethod.Set(clientSpan, "GET")

		// Inject the client span context into the headers
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		resp, _ := http.DefaultClient.Do(req)
		ext.HTTPStatusCode.Set(clientSpan, uint16(resp.StatusCode))
		clientSpan.Finish()
		time.Sleep(time.Second * 5)
	}
}
