module github.com/vmware/hamlet/avi-client-operator

go 1.13

require (
	github.com/avinetworks/sdk v0.0.0-20210401165128-e7d04ab96563
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.5
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/vmware/hamlet v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.36.1
	istio.io/pkg v0.0.0-20210322140956-5892a3b28d3e
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)

replace github.com/vmware/hamlet/avi-client-operator => /Users/sushils/work/vmware/allspark/hamlet/avi-client-operator

replace github.com/vmware/hamlet => /Users/sushils/work/vmware/allspark/hamlet
