# Operator for creating Hamlet based client for sync to AVI LoadBalancer

## Project creation steps
cd hamlet-avi-client
go mod init github.com/vmware/hamlet/avi-client-operator
kubebuilder init --domain tanzu.vmware.com
kubebuilder  create api --group hamlet --version v1alpha1 --kind AVISync --resource --controller
