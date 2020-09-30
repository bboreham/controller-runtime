package tracing

import (
	"fmt"
	"io"

	ot "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// SetupJaeger sets up Jaeger with some defaults and config taken from the environment
func SetupJaeger(serviceName string) (io.Closer, error) {
	jcfg := &jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LocalAgentHostPort: "jaeger:6831", // can be overridden by env JAEGER_AGENT_HOST
		},
	}
	jcfg, err := jcfg.FromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed read config from environment: %w", err)
	}

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := jcfg.NewTracer()
	if err != nil {
		return nil, err
	}
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	ot.SetGlobalTracer(tracer)
	return closer, nil
}
