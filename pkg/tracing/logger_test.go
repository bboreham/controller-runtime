package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func intermediate(keysAndValues ...interface{}) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	attrs = append(attrs, keyValues(keysAndValues...)...)
	return attrs
}

func TestKeyValues(t *testing.T) {
	got := intermediate("foo", "foo-value", "bar", 42)
	expected := []attribute.KeyValue{attribute.String("foo", "foo-value"), attribute.Int("bar", 42)}
	assert.Equal(t, expected, got)
}
