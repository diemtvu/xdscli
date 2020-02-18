module xdscli

go 1.14

replace istio.io/istio => github.com/istio/istio v0.0.0-20200218044045-88b0085faa96

require (
	github.com/envoyproxy/go-control-plane v0.9.4
	github.com/golang/protobuf v1.3.3
	github.com/spf13/cobra v0.0.5
	google.golang.org/grpc v1.27.1
	istio.io/istio v0.0.0-20200218044045-88b0085faa96
	istio.io/pkg v0.0.0-20200214155848-e5ca416a8c07
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
)
