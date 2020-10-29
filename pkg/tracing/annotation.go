package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// TraceAnnotationPrefix is where we store span contexts in Kubernetes annotations
const TraceAnnotationPrefix string = "trace.kubernetes.io/"

// Store tracing propagation inside Kubernetes annotations
type annotationsCarrier map[string]string

var _ propagation.TextMapCarrier = annotationsCarrier{}

// Get implements propagation.TextMapCarrier
func (a annotationsCarrier) Get(key string) string {
	return a[TraceAnnotationPrefix+key]
}

// Set implements propagation.TextMapCarrier
func (a annotationsCarrier) Set(key string, value string) {
	a[TraceAnnotationPrefix+key] = value
}

// Keys lists the keys stored in this carrier (implements propagation.TextMapCarrier).
func (a annotationsCarrier) Keys() []string {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	return keys
}

// SpanFromAnnotations takes a map as found in Kubernetes objects and
// makes a new Span parented on the context found there, or nil if not found.
func SpanFromAnnotations(ctx context.Context, name string, annotations map[string]string) (context.Context, trace.Span) {
	innerCtx := spanContextFromAnnotations(ctx, annotations)
	if innerCtx == ctx {
		return ctx, nil
	}
	return otel.Tracer(libName).Start(innerCtx, name)
}

func spanContextFromAnnotations(ctx context.Context, annotations map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, annotationsCarrier(annotations))
}

// AddTraceAnnotation adds an annotation encoding current span ID
func AddTraceAnnotation(ctx context.Context, annotations map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, annotationsCarrier(annotations))
}

// AddTraceAnnotationToUnstructured adds an annotation encoding current span ID to all objects
// Objects are modified in-place.
func AddTraceAnnotationToUnstructured(ctx context.Context, objs []unstructured.Unstructured) error {
	for _, o := range objs {
		a := o.GetAnnotations()
		if a == nil {
			a = make(map[string]string)
		}
		AddTraceAnnotation(ctx, a)
		o.SetAnnotations(a)
	}

	return nil
}

// AddTraceAnnotationToObject - if there is a span for the current context, and
// the object doesn't already have one set, adds it as an annotation
func AddTraceAnnotationToObject(ctx context.Context, obj runtime.Object) error {
	m, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	annotations := m.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	} else {
		// Check if the object already has some context set.
		for _, key := range otel.GetTextMapPropagator().Fields() {
			if annotationsCarrier(annotations).Get(key) != "" {
				return nil // Don't override
			}
		}
	}
	AddTraceAnnotation(ctx, annotations)
	m.SetAnnotations(annotations)
	return nil
}
