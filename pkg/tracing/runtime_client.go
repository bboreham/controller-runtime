package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

const libName = "controller-runtime" // shows up in traces

type tracingClientBuilder struct {
	upstream cluster.ClientBuilder
}

func (b *tracingClientBuilder) WithUncached(objs ...client.Object) cluster.ClientBuilder {
	return b.upstream.WithUncached(objs...)
}

func (b *tracingClientBuilder) Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	client, err := b.upstream.Build(cache, config, options)
	if err != nil {
		return nil, err
	}
	return &tracingClient{Client: client, scheme: options.Scheme}, nil
}

// WrapClientBuilder wraps a ClientBuilder function with one that does tracing
func WrapClientBuilder(upstream cluster.ClientBuilder) cluster.ClientBuilder {
	return &tracingClientBuilder{upstream: upstream}
}

func objectAttrs(obj runtime.Object) (attrs []attribute.KeyValue) {
	if gvk := obj.GetObjectKind().GroupVersionKind(); !gvk.Empty() {
		attrs = append(attrs, attribute.String("objectKind", gvk.String()))
	}
	if m, err := meta.Accessor(obj); err == nil {
		attrs = append(attrs, attribute.String("objectKey", m.GetNamespace()+"/"+m.GetName()))
	}
	return
}

func logStart(ctx context.Context, op string, attrs ...attribute.KeyValue) trace.Span {
	sp := trace.SpanFromContext(ctx)
	if sp != nil {
		sp.AddEvent(op, trace.WithAttributes(attrs...))
	}
	return sp
}

func logError(ctx context.Context, sp trace.Span, err error) error {
	if sp != nil && err != nil {
		sp.RecordError(err)
	}
	return err
}

// wrapper for Client which emits spans on each call
type tracingClient struct {
	client.Client
	scheme *runtime.Scheme
}

func (c *tracingClient) blankObjectAttrs(obj runtime.Object) (attrs []attribute.KeyValue) {
	if c.scheme != nil {
		gvks, _, _ := c.scheme.ObjectKinds(obj)
		for _, gvk := range gvks {
			attrs = append(attrs, attribute.String("objectKind", gvk.String()))
		}
	}
	return
}

func (c *tracingClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	sp := logStart(ctx, "k8s.Get", append([]attribute.KeyValue{attribute.String("objectKey", key.String())}, c.blankObjectAttrs(obj)...)...)
	return logError(ctx, sp, c.Client.Get(ctx, key, obj))
}

func (c *tracingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	sp := logStart(ctx, "k8s.List", c.blankObjectAttrs(list)...)
	return logError(ctx, sp, c.Client.List(ctx, list, opts...))
}

func (c *tracingClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	AddTraceAnnotationToObject(ctx, obj)
	sp := logStart(ctx, "k8s.Create", c.blankObjectAttrs(obj)...)
	return logError(ctx, sp, c.Client.Create(ctx, obj, opts...))
}

func (c *tracingClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	sp := logStart(ctx, "k8s.Delete", objectAttrs(obj)...)
	return logError(ctx, sp, c.Client.Delete(ctx, obj, opts...))
}

func (c *tracingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	sp := logStart(ctx, "k8s.Update", objectAttrs(obj)...)
	return logError(ctx, sp, c.Client.Update(ctx, obj, opts...))
}

func (c *tracingClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	attrs := objectAttrs(obj)
	if data, err := patch.Data(obj); err == nil {
		attrs = append(attrs, attribute.String("patch", string(data)))
	}
	sp := logStart(ctx, "k8s.Patch", attrs...)
	return logError(ctx, sp, c.Client.Patch(ctx, obj, patch, opts...))
}

func (c *tracingClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	sp := logStart(ctx, "k8s.DeleteAllOf", c.blankObjectAttrs(obj)...)
	return logError(ctx, sp, c.Client.DeleteAllOf(ctx, obj, opts...))
}

func (c *tracingClient) Status() client.StatusWriter {
	return &tracingStatusWriter{StatusWriter: c.Client.Status()}
}

type tracingStatusWriter struct {
	client.StatusWriter
}

func (s *tracingStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	sp := logStart(ctx, "k8s.Status.Update", objectAttrs(obj)...)
	return logError(ctx, sp, s.StatusWriter.Update(ctx, obj, opts...))
}

func (s *tracingStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	attrs := objectAttrs(obj)
	if data, err := patch.Data(obj); err == nil {
		attrs = append(attrs, attribute.String("patch", string(data)))
	}
	sp := logStart(ctx, "k8s.Status.Patch", attrs...)
	return logError(ctx, sp, s.StatusWriter.Patch(ctx, obj, patch, opts...))
}
