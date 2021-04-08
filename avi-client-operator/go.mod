module github.com/vmware/hamlet/avi-client-operator

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)

replace github.com/vmware/hamlet/avi-client-operator => /Users/sushils/work/vmware/allspark/hamlet/avi-client-operator

replace github.com/vmware/hamlet/avi-client-operator/api/v1alpha1 => /Users/sushils/work/vmware/allspark/hamlet/avi-client-operator/api/v1alpha1
