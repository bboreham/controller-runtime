package tracing

import (
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type tracingLogger struct {
	logr.Logger
	trace.Span
}

func (t tracingLogger) Enabled() bool {
	return t.Logger.Enabled()
}

func (t tracingLogger) Info(msg string, keysAndValues ...interface{}) {
	t.Logger.Info(msg, keysAndValues...)
	t.Span.AddEvent(msg, trace.WithAttributes(keyValues(keysAndValues...)...))
}

func (t tracingLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	t.Logger.Error(err, msg, keysAndValues...)
	kvs := append([]attribute.KeyValue{attribute.String("error", msg)}, keyValues(keysAndValues...)...)
	t.Span.AddEvent(msg, trace.WithAttributes(kvs...))
	t.Span.RecordError(err)
}

func (t tracingLogger) V(level int) logr.Logger {
	return tracingLogger{Logger: t.Logger.V(level), Span: t.Span}
}

func keyValues(keysAndValues ...interface{}) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(keysAndValues)/2)
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = "non-string"
		}
		attrs = append(attrs, attribute.Any(key, keysAndValues[i+1]))
	}
	return attrs
}

func (t tracingLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	t.Span.SetAttributes(keyValues(keysAndValues...)...)
	return tracingLogger{Logger: t.Logger.WithValues(keysAndValues...), Span: t.Span}
}

func (t tracingLogger) WithName(name string) logr.Logger {
	t.Span.SetAttributes(attribute.String("name", name))
	return tracingLogger{Logger: t.Logger.WithName(name), Span: t.Span}
}
