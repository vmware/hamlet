# Hamlet

[![Slack](https://img.shields.io/badge/slack-join%20chat-e01563.svg?logo=slack)](https://code.vmware.com/web/code/join)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

Hamlet specifies a set of API standards for enabling service mesh federation.
The [API specification](spec/service-discovery.md) was realized as a
collaborative effort with service mesh vendors viz. Google, HashiCorp, Pivotal
and VMware.

The specification currently consists of the following APIs.

* [Federated Resource Discovery API](api/resourcediscovery/v1alpha1/resource_discovery_v1alpha1.proto)
  \- API to authenticate and securely distribute resources between federated
  service meshes.
* [Federated Service Discovery API](api/types/v1alpha1/federated_service_v1alpha1.proto)
  \- API to discover, reach, authenticate, and securely communicate with
  federated services.
## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) if you'd like to contribute.

## License

Hamlet is licensed under the Apache License, Version 2.0. See the
[LICENSE](LICENSE) file for the full license text.
