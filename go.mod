module sigs.k8s.io/controller-runtime

go 1.16

require (
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v0.18.0
	go.opentelemetry.io/otel/exporters/otlp v0.18.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.18.0
	go.opentelemetry.io/otel/sdk v0.18.0
	go.opentelemetry.io/otel/trace v0.18.0
	go.uber.org/goleak v1.1.10
	go.uber.org/zap v1.16.0
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324
	gomodules.xyz/jsonpatch/v2 v2.1.0
	k8s.io/api v0.21.0-beta.0
	k8s.io/apiextensions-apiserver v0.21.0-beta.0
	k8s.io/apimachinery v0.21.0-beta.0
	k8s.io/client-go v0.21.0-beta.0
	k8s.io/component-base v0.21.0-beta.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/yaml v1.2.0
)
