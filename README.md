# Hamlet

Hamlet specifies a set of API standards for enabling service mesh federation.
The API specification was realized as a collaborative effort with service mesh
vendors viz. Google, HashiCorp, Pivotal and VMware.

The specification currently consists of the following APIs.

* [Federated Resource Discovery API](api/resourcediscovery/v1alpha1/resource_discovery.proto)
  \- API to authenticate and securely distribute resources between federated
  service meshes.
* [Federated Service Discovery API](api/types/v1alpha1/federated_service.proto)
  \- API to discover, reach, authenticate, and securely communicate with
  federated services.

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

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) if you'd like to contribute.

## License

Hamlet is licensed under the Apache License, Version 2.0. See the
[LICENSE](LICENSE) file for the full license text.
