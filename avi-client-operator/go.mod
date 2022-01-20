module github.com/vmware/hamlet/avi-client-operator

go 1.13

require (
	github.com/avinetworks/sdk v0.0.0-20210401165128-e7d04ab96563
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.5
	github.com/vmware/hamlet v0.0.0-00010101000000-000000000000
	istio.io/pkg v0.0.0-20210322140956-5892a3b28d3e
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)

// replace github.com/vmware/hamlet => /Users/sushils/work/vmware/allspark/hamlet
replace github.com/vmware/hamlet => github.com/sushilks/hamlet v0.0.1-alpha.1
