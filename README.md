# Hamlet

[![CircleCI](https://img.shields.io/circleci/project/github/vmware/hamlet/master.svg?logo=circleci)](https://circleci.com/gh/vmware/hamlet)
[![Slack](https://img.shields.io/badge/slack-join%20chat-e01563.svg?logo=slack)](https://code.vmware.com/web/code/join)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

Hamlet specifies a set of API standards for enabling service mesh federation.
The [API specification](spec/service-discovery.md) was realized as a
collaborative effort with service mesh vendors viz. VMware, HashiCorp, Pivotal
and Google.

The specification currently consists of the following APIs.

* [Federated Resource Discovery API](api/resourcediscovery/v1alpha1/resource_discovery.proto)
  \- API to authenticate and securely distribute resources between federated
  service meshes.
* [Federated Service Discovery API](api/types/v1alpha1/federated_service.proto)
  \- API to discover, reach, authenticate, and securely communicate with
  federated services.

## Extending

Please see [EXTENDING.md](EXTENDING.md) if you'd like to extend this project's
core to implement functionality for an owning or consuming federated service
mesh. You can find concrete extension samples in the [examples](examples/)
directory.

## Building

To compile a Protobuf file into Go, you can use the following command.

```console
$ protoc -I api/types/v1alpha1/ federated_service.proto --go_out=api/
```

To download the external dependencies, use the following commands.

```console
$ mkdir -p external/google/rpc/
$ curl -L -o external/google/rpc/status.proto https://github.com/grpc/grpc/raw/master/src/proto/grpc/status/status.proto
```

To compile a Protobuf into Go along with the necessary gRPC server and client
stubs, you can use the following command.

```console
$ protoc -I external/ -I api/resourcediscovery/v1alpha1/ resource_discovery.proto --go_out=plugins=grpc:api/
```

## Testing

The project relies on the [mockgen](https://github.com/golang/mock#installation)
tool for generating gRPC mocks for unit tests. Please make sure that you have it
installed before proceeding.

1\. Generate mocks.

```console
$ go generate ./...
```

2\. Run tests using [Ginkgo](https://onsi.github.io/ginkgo/) or `go test`.

```console
$ ginkgo -v ./...
```

OR

```console
$ go test -v ./...
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) if you'd like to contribute.

## License

Hamlet is licensed under the Apache License, Version 2.0. See the
[LICENSE](LICENSE) file for the full license text.
