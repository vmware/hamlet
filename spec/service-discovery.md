# Service Discovery API

## Status of this Memo

This document specifies an experimental Service Discovery API standard for the
internet community, and requests, discussions, and suggestions for improvements.
Distribution of this document is unlimited.

An original version of this document is preserved [here](https://docs.google.com/document/d/1CVaKfvY_myM1-PNOl-cPXWDN_eVyAuPMSkTmSc-n9p4) for historical record.

## Table of Contents

TODO

## Introduction

There is established industry precedent for using standardized APIs for shared
activity across companies and organizational unit boundaries. But in an
enterprise-scale world demanding resilient, secure and high-performance
connections for shared workflows, APIs are not enough.

What if services meshes could be interconnected to deliver the associated
benefits of observability, control, security, etc. when those meshes are managed
by different organizations, and different vendors?

Much has been discussed about multi-cluster deployments of both Kubernetes and
Istio. Kubernetes main use case is configuration and resource replication across
clusters for either disaster recovery or high availability. With Istio you can
also expand a service mesh to include services running on VMs or bare metal
hosts, or combine the services from more than one cluster into a single
composite service mesh. While these use cases are sound and needed, both presume
that all of the clusters are under a shared and single administrative control.

Very little has been said about service mesh interoperation, where each mesh is
in a different and untrusted administrative domain (and hence workloads are
loosely coupled), and where each mesh can be of the same or different vendors,
can have the same or different control and data plane implementations, be single
or multi-cluster, and can provide the same or different functionality to its
customers.

In this document, we present a new service mesh interoperation OSS initiative to
discover, reach, and securely communicate services across meshes in different
administrative domains.

## Service Mesh Federation

For the purposes of this document, a Service Mesh (SM from now on) Federation
is comprised the problems of:

* Identity federation
* Service discovery federation
* Secure service to service communication across different administrative
  domains
* Policy federation

across multiple SM or peers. How each SM builds the infrastructure where the
services will run must be irrelevant to the SM Federation.

A SM Federation comprises the following features and constraints:

1. Several SM administrative domains, one for each member of the federation.
2. Each SM of the federation can and will continue operating as a standalone SM,
in addition to being a member of the federation.
3. Each SM of the federation is considered a black box to the others and only
   API interoperability can be guaranteed.
4. Each SM of the federation can be built using a different underlay technology
   (based or not on Istio, Kubernetes, etc.) and this must be transparent to the
   others.
5. Each SM of the federation can be running either on-prem or in the Cloud
   (either public or private) and this must be transparent to the others.
   Equally, how each SM of the federation builds the infrastructure where the
   services run must be transparent to the others (for example, single or multi
   cluster).
6. Workloads in each SM of the federation can run in any infrastructure
   (containers, VMs, physical servers, etc) and this must be transparent to the
   others.

The primary goal of this specification is to allow two different loosely coupled
services in two different administrative domains to discover and securely
communicate.

It is not a goal of this specification to load balance between the same two
(likely tightly coupled) services in two different administrative domains.

## Service

Even though there is no official definition of what microservices are, a
consensus view has evolved over time in the industry. For the purposes of this
document we will use the definition provided by Martin Fowler and other experts:

> Services in a microservice architecture are often processes that communicate
> over a network to fulfil a goal using technology-agnostic protocols such as
> HTTP.

## Federated Service

In a Federated SM, workloads will be placed in a different SM of the federation
depending on either the functionality required by the organization and provided
by each member of the federation, or the needs to interoperate different
services across organizations or within the same organization when services are
distributed across different locations.

A federated service describes the properties that a Federated Owner SM needs to
expose to a Federated Consumer SM in order for it to be able to discover, reach,
authenticate, and securely communicate with. A federated service adds additional
entries into the Federated Consumer SM internal service registry, so that
auto-discovered services in the Federated Consumer SM can access/route to these
federated services in the Federated Owner SM.

## Architecture

In order to simplify the explanation of the architecture, let’s assume that
there are two SMs in a federation, an owner of federated services, and a
consumer of federated services. In a real-world scenario there can be an
unlimited number of federated SMs, each of them behaving as owner and/or
consumer (dual behaviour is allowed and encouraged):

* **Server:** A Federated Service Owner SM that owns federated services which
  are exposed to another federated SM, which consumes them.
* **Client:** A Federated Service Consumer SM that consumes federated services
  which are exposed by another federated SM.

Since a single SM can both own and consume federated services, and the
implementers of the proposed protocol must have a server and a client endpoint
(Figure 1).

The Federated Service Discovery protocol is derived from Istio/MCP and
implemented as a gRPC bi-directional stream in order to quickly propagate
catalog updates when new federated services are registered or deregistered in a
Federated Service Owner SM. In a federation there is an unknown number of
federated service consumer SMs, each with a service catalog, which must be
updated as quickly as possible. The synchronization mechanism is asynchronous
since each federated service entry is independent from others and update
messages are placed on the wire as soon as federated services are added or
removed from the service owner catalog. Message passing protocol is synchronous
with only one message in flight.

Methods of the Federated Service Discovery API must be idempotent.

![Figure 1](figure-1.png "Figure 1. Client and server endpoints of the Federated
Service Discovery protocol")

*Figure 1. Client and server endpoints of the Federated Service Discovery
protocol*

### Federated service owner and consumer behaviour

The Federated Service Discovery API is a subscription-based catalog
synchronization / data replication API. The subscription is established during
the connection from a federated service consumer to a federated service owner.
Only federated service entries are synchronized between a federated service
owner and a consumer (see Federated Service Discovery API section in this
document). 

Upon subscription, the federated service consumer requests a download of the
full catalog of the federated service owner. The owner pushes all the federated
services in its catalog to the newly registered consumer. From there, the owner
sends catalog updates (add, removal, modifications of federated service entries)
as soon as services are registered or deregistered in the federated service
owner catalog. The consumer positively acknowledges the service addition or
removal from its own local catalog, if it was accepted, and negatively
acknowledges (with an error code) if it was rejected. Each federated service
entry in the catalog contains a unique identification to identify the federated
service entries across SMs. This ID is used when sending federated services
updates, and also in the ACK/NACK messages from the consumer.

Every message from a service owner containing a federated service must include
the full set of information. This avoids complexity associated with state
tracking on both client and server implementations, including the need for
anti-entropy mechanisms.

### Connection establishment and termination

Let’s assume two SMs that have been mutually authenticated and connected through
an API endpoint which offers the Federated Service Discovery API services.

Consumers can only send requests and notification messages to the owner, where
some of them are considered response-generating events. When the connection
between a consumer SM and an owner SM is first created and they are
authenticated to each other, the consumer registers itself in the owner and the
owner sends the full catalog of the federated services to the consumer. If the
owner is not federating any services yet, then it will send an empty catalog. As
soon as the consumer starts receiving federated services from the owner catalog,
it starts acknowledging after decoding, validating, and persisting the update to
its local service registry. See Figure 2.

![Figure 2](figure-2.png "Figure 2. Connection establishment of a Service Owner
and a Service Consumer")

*Figure 2. Connection establishment of a Service Owner and a Service Consumer*

Implementers of the Federated Service Discovery API should take into account
that any of the SMs in the federation (owner or consumer) can terminate the
connection at any time because of network unreliability, or intentionally in an
ordered way when the consumer disconnects from the owner.

If a consumer disconnects from an owner because errors are encountered (in
either server or client side), the consumer is responsible for re-initiating the
connection. Since the number of federated services updates during the time
between disconnection and reconnection is unknown, on reconnection the consumer
SM will download again the full catalog from the owner SM. How the consumer SM
catalog is reconciled with the newly downloaded owner catalog after unexpected
disconnection is out of the scope of this spec. Federated service DNS entries in
the consumer SM can remain for a period of time which is part of the SM
configuration, and it is recommended that clients cache the catalog of
previously downloaded services for some (unspecified) time to avoid cascading
failures. This time can range from zero (unlimited, services stay forever in the
DNS) to any arbitrary number expressed in multiples of seconds.

However, a consumer can intentionally disconnect from an owner, too. In this
case, the consumer must delete the local service catalog which was downloaded
and potentially updated from the owner. See Figure 3.

![Figure 3](figure-3.png "Figure 3. Connection termination of a Service Owner
and a Service Consumer")

*Figure 3. Connection termination of a Service Owner and a Service Consumer*

### Catalog updates

Once the initial handshake has finished and the owner catalog has been
downloaded to a consumer, the owner will send update messages to the consumer as
soon as possible each time a federated service is registered or deregistered in
the owner catalog.

The exact timing of the owner messages is directed by events. These are the
possible events 

* `CREATE` when the service owner has a new federated service.
* `DELETE` when the service owner has removed a federated service.
* `UPDATE` when the service owner has updated an existing federated service.

Upon receiving an update message, the consumer acknowledges after decoding,
validating, and persisting the update to its local service registry.

In order to provide clarity, authors are providing sample state diagrams of both
client and server implementations of the registration of a federated service of
the Federated Service Discovery API. Deregistration follows a similar process.

It should be noted that there are many ways in which the process can be
implemented while it conforms to the specification, and that this particular
implementation is only for reference.

This reference implementation can be found in Appendix A.

## Accessibility

Each SM of the federation will expose its Federated Service Discovery API
endpoint to the rest through a routable IP address or a FQDN and a port number.
Since this specification does not impose constraints in the addressing space
where the federation happens, the IP address can be in either a private or
public space. Alternatively, the FQDN can resolve to a private or a public IP
address. The API endpoint also contains a version string.

As an example, valid Federated Service Discovery API endpoints would be <br>
https://150.214.141.1:8008 <br>
https://acme.example.com:8008

The endpoint can be located either in the federated SMs or exposed through a
load balancer or a similar technology. As long as requests can reach the API
endpoint and traffic between services can be routed and secured, this should be
transparent to the federation mechanism.

### Versioning

To be able to eliminate fields or restructure resource representations, the
Federated Service Discovery API supports multiple API (interface) versions. 

The version is set at the API interface level rather than at the resource or
field level to

* Ensure that each API version presents a self-contained, clear and consistent
  view of system resources and behaviour.
* Enable access control to end-of-life and/or experimental APIs.

The Federated Service Discovery Endpoint API specification provided in this
specification is serialized using Protobufs.

### Transport

The Federated Service Discovery endpoint must be served over gRPC. In any case,
strong transport layer security is a requisite of the Federated Service
Discovery API. In the same way, service to service communication must be mTLS.

In future versions of the specification, HTTPS support may be provided, where
compliant clients must support at least one of the two transport protocols. gRPC
to JSON transcoding, including gRPC-HTTP error codes equivalencies, must follow
the rules described in https://cloud.google.com/endpoints/docs/grpc/transcoding.

### Authentication

Before any information is interchanged between two federated SMs, mutual gRPC
channel authentication is required (unsecured connections are not allowed). The
Federated Service Discovery API supported authentication mechanism will be X509
Digital certificates which are manually installed in each SM of the federation.
This mechanism authenticates and secures the communication channel between
federated SM. Request authentication is out of the scope of this specification
and may be considered in future versions.


## Error Codes

A number of error conditions may be encountered by the client when interacting
with the server endpoint. These errors range from generic problems with the gRPC
server to a malformed request (with incomplete or wrong parameters).

In the event that the Federated Service Discovery endpoint is running but
unavailable, for instance if it is still initializing, client implementations
will receive the gRPC status code `Unavailable`. Clients receiving this code or
clients which are unable to reach the Federated Service Discovery API endpoint
can retry with an exponential backoff with a minimum delay of 1 second.

Clients that are not authenticated will receive the gRPC status code
`Unauthenticated`. Clients encountering the `Unauthenticated` status code must
not retry, as this indicates that the meshes as possibly not federated, and it
is a non-recoverable error.

Client implementations which do not send at least the required mandatory
arguments for the messages, or which send malformed or incorrect arguments will
receive the gRPC status code `InvalidArgument`.

A summary of error conditions and codes can be found in Appendix B.

## Federated Service Discovery API Endpoint Specification

A federated service describes the properties that a service owner needs to
expose to a federated consumer platform in order for it to be able to discover,
reach, authenticate, and securely communicate. A federated service adds
additional entries into the consumer platform internal service registry, so that
federated services can be auto-discovered in the consumer platform (as if they
were local) and traffic routed to it.

The federated owner platform (server) will implement the following APIs.

### registerConsumer

Registers a federated consumer platform in a federated owner platform. The owner opens a server
stream and sends the full federated service catalog to the client. Before
registering, both client and server must have been authenticated to each other and a secure
TLS channel created.

**Request message format** <br>
`registerConsumer` does not provide a request payload.

**Response stream message format** <br>
`registerConsumer` does not provide a response payload.

**Errors**

| Code | Condition | Client behavior |
| --- | --- | --- |
| `Unavailable` | The Federated Service Discovery API endpoint is unable to handle the request from the client | Retry with a backoff |
| `InvalidArgument` | Request data doesn’t contain all the required mandatory fields. Request data contains all the required mandatory fields but data is malformed. | Report error and don’t retry |

### CREATE

Adds a federated service to the consumer platform Then, the federated service can be
discovered by the consumer platform as a local service. A `FederatedService` is uniquely idenfied in the owner platform by its `name`. 

A `FederatedService` is composed of multiple instances. In this way, we are creating a hierarchy in a way that a `FederatedService` can contain multiple backing services under the same FQDN. For flexibility, each `Instance` can be exposed in a different `Endpoint`, or on the same one, using a labeling system in both `Instance` and `Endpoint`. If there is no matching selector for an `Instance`, then consumer agent implementors must associate to that `Instance` all endpoints.

For each `FederatedService`, the consumer agent must create in the consumer platform an entry in the local service registry (for example, in the local DNS) with an FQDN for each of the `Instance` with the format `FederatedServiceInstance.spec.instanceID+FederatedService.spec.fqdn`. In addition, the consumer agent must create in the local service registry an entry with `FederatedService.spec.fqdn` and all the information in each `Instance` and corresponding `Endpoint`.

**Request message format**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | Yes | The unique name that identifies the service within an owner platform, and serves as its unique ID. Uniqueness in the service owner platform. The consumer platform must take into account possible name collistions that may happen with other `FederatedService` consumed from other owner platforms through, for example, aliasing. |
| `FQDN` | `string` | Yes | The FQDN (Fully Qualified Domain Name) the consumer platform uses as a local DNS name to access the imported `FederatedService`. The consumer platform DNS is programmed to reference the `Endpoint` of the imported `FederatedService`. It is unique within the service owner platform. The consumer platform must solve any possible naming collisions that may happen with other `FederatedService` consumed from other owner platforms through, for example, aliasing. |
| `SANS` | `Array<string>` | No | List of URI SANs A list of SANs (Subject Alternative Names) to enable end-to-end security for the federated service. |
| `Instances` | `Array<Instance>` | Yes | Represents an instance of the `FederatedService`. |
| `Endpoints` | `Array<Endpoint>` | Yes | The endpoints that provide access to a `FederatedService` in the owner platform. For example, this can be an ingress. |
| `Description` | `string` | No | Description of the `FederatedService`. |
| `Tags` | `Array<string>` | No | Informative values for filtering purposes. |
| `Labels` | `Map<string, string>` | No | Informative array of KV pairs for filtering purposes. |

**Response stream message format**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | Yes | The unique name that identifies the service within an owner platform, and serves as its unique ID. |

**Errors**

| Code | Condition | Client behavior |
| --- | --- | --- |
| `Unavailable` | The Federated Service Discovery API endpoint is unable to handle the request from the client | Retry with a backoff |

**Request message payload example**

```
FederatedService {
    Name: “example-service”
    FQDN: “db.mysql.com”
    SANS: [ “URI:spiffe://db1.mysql.com” ]
    Instances: [ DBInstance ]
    Endpoints: [ DBEndpoint ]
    Description: "This is an example federated database service"
    Tags: [ "database" ]
    Labels: { "version": "3.6" }
}

Instance {
    Name: “example-service-dbinstance1”
    Protocol: “https"
    Metadata: { "SNI": "outbound_.8080_.v1_.db.mysql.com" }
    Endpoint_selector: [ "dbendpoint1" ]
    Description: "This is the database endpoint."
}

Endpoint {
    Address: "84.15.190.249"
    Port: 443
    Labels: [ "dbendpoint1" ]
    Description: "This is the endpoint where the owner platform is exposing the database" 
}
```

### DELETE

Removes a federated service in the federated SM consumer.

**Request message parameters**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | Yes | The unique name that identifies the service within an owner platform, and serves as its unique ID. Uniqueness in the service owner platform. The consumer platform must take into account possible name collisions that may happen with other `FederatedService` consumed from other owner platforms through, for example, aliasing. |

**Response stream message format**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | Yes | The unique name that identifies the service within an owner platform, and serves as its unique ID. |

**Errors**

| Code | Condition | Client behavior |
| --- | --- | --- |
| `Unavailable` | The Federated Service Discovery API endpoint is unable to handle the request from the client | Retry with a backoff |
| `InvalidArgument` | Request data doesn’t contain all the required mandatory fields. Request data contains all the required mandatory fields but data is malformed. | Report error and don’t retry |

**Request message payload example**

```
{
     Name: “example-service”
}
```

### UPDATE

Modifies any number of the fields of a federated service in the consumer platform, except `Name` (it it changed, it would be then considered to be a different
federated service). If `Name` needs to be changed, the suggested approach would be to remove the federated service and create a new one with the new name. In an update message, the full object must be sent, meaning that the attributes that do not change, must be filled with their current values.

**Request message parameters**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | Yes | The unique name that identifies the service within an owner platform, and serves as its unique ID. Uniqueness in the service owner platform. The consumer platform must take into account possible name collisions that may happen with other `FederatedService` consumed from other owner platforms through, for example, aliasing. |
| `FQDN` | `string` | Yes | The FQDN (Fully Qualified Domain Name) the consumer platform uses as a local DNS name to access the imported `FederatedService`. The consumer platform DNS is programmed to reference the Endpoint of the imported `FederatedService`. It is unique within the service owner platform. The consumer platform must solve any possible naming collisions that may happen with other `FederatedService` consumed from other owner platforms through, for example, aliasing. |
| `SANS` | `Array<string>` | Yes | List of URI SANs A list of SANs (Subject Alternative Names) to enable end-to-end security for the federated service. |
| `Instances` | `Array<Instance>` | Yes | Represents an instance of the `FederatedService`. |
| `Endpoints` | `Array<Endpoint>` | Yes | The endpoints that provide access to a `FederatedService` in the owner platform. For example, this can be an ingress. |
| `Description` | `string` | Yes | Description of the `FederatedService`. |
| `Tags` | `Array<string>` | Yes | Informative values for filtering purposes. |
| `Labels` | `Map<string, string>` | Yes | Informative array of KV pairs for filtering purposes. |

**Response stream message format**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `Name` | `string` | Yes | The unique name that identifies the service within an owner platform, and serves as its unique ID. |

**Errors**

| Code | Condition | Client behavior |
| --- | --- | --- |
| `Unavailable` | The Federated Service Discovery API endpoint is unable to handle the request from the client | Retry with a backoff |
| `InvalidArgument` | Request data doesn’t contain all the required mandatory fields. Request data contains all the required mandatory fields but data is malformed. | Report error and don’t retry |

**Request message payload example**

```
FederatedService {
    Name: “example-service”
    FQDN: “db.mysql.com”
    SANS: [ “URI:spiffe://db1.mysql.com” ]
    Instances: [ DBInstance ]
    Endpoints: [ DBEndpoint ]
    Description: "This is an example federated database service"
    Tags: [ "database", "mysql" ]
    Labels: { "version": "3.6", "zone": "us-east-1" }
}

Instance {
    Name: “example-service-dbinstance1”
    Protocol: “https"
    Metadata: { "SNI": "outbound_.8080_.v1_.db.mysql.com" }
    Endpoint_selector: [ "dbendpoint1" ]
    Description: "This is the database endpoint."
}

Endpoint {
    Address: "84.15.190.249"
    Port: 443
    Labels: [ "dbendpoint1" ]
    Description: "This is the endpoint where the owner platform is exposing the database" 
}
```

### Complex Type Instance

An `Instance` represents one of the many possible backing microservices for a federated service.

`Instance` has mandatory labels. These instance labels are used to associate each `Instance` to one or more `Endpoint`. If there are no matches, consumer agent implementors must associate to the `Instance` all `Endpoint`.

The Metadata attribute is used to be able to transport the required information for each different use case, through reserved keys. Additional keys may also be used and may be incorporated as additional reserved keys in the protocol object model.

For example, for TLS connections with SNI routing at owner platform ingress, you may want to use SNI. It is worth noting that there is no set format for the value of the SNI in the context of a federated service, as different platforms may have different formats. As long as the service owner is able to facilitate communication with the federated service trough this value to its service mesh ingress, the owner is free to choose the format it prefers.

For example, for TCP cleartext connections with hostname based routing, you may want to use HOSTNAME and PORT.

For example, for TCP cleartext connections with NAT based routing, you may want to use only PORT.

**Instance**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `ID` | `string` | Yes | The unique identifier for the federated service instance within its group of instances for a FederatedService. |
| `Protocol` | `string` | Yes | The protocol corresponding to the federated service instance. MUST be one of `HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS|MTLS` |
| `Metadata` | `Map<string, string>` | Yes | Additional metadata for the purpose of establishing connectivity. RESERVED keys are `SNI|HOSTNAME|PORT` |
| `EndpointSelector` | `Array<string>` | Yes | The endpoint labels are used to associate the `Instance` to one or more `Endpoint`. If there are no matches, the `Instance` is associated to all `Endpoint`. |
| `Description` | `string` | No | Description of the Instance. |

### Complex Type Endpoint

An `Endpoint` represents the location of a service that provides access to a federated service in the owner platform. In this spec, this type has been used to represent a network entry point for an owner platform (typically an ingress gateway) that can be used to access one or more instances of the federated service.

`Endpoint` has mandatory labels. These endpoint labels are used to associate each `Instance` to one or more `Endpoint`. If there are no matches, consumer agent implementors must associate to the mismatching `Instance` all `Endpoint`.

**Endpoint**

| Parameter | Data type | Required | Description |
| --- | --- | --- | --- |
| `Address` | `string` | Yes | This is the address associated with the network endpoint. Valid values are host IP addresses and FQDN. |
| `Port` | `int` | Yes | Port associated with the network endpoint where the federated service is available . |
| `Labels` | `Array<string>` | Yes | The endpoint labels are used to associate each `Instance` to one or more `Endpoint`. If there are no matches, the `Instance` is associated to all `Endpoint`. |
| `Description` | `string` | No | Description of the Endpoint. |

## Appendix A. Sample State Machine Implementation of registerFederatedService

These state diagrams show a sample implementation of the
`registerFederatedService` method of the Federated Service Discovery API. The
client of the service owner initiates the connection to the server of the
service consumer. It is assumed that the Federated Service Discovery API
endpoints have already been mutually authenticated.

### Server State Machine

![Figure 4](figure-4.png "Figure 4. Example implementation of the Service Owner
(server)")

*Figure 4. Example implementation of the Service Owner (server)*

1. The Federated Service Discovery API endpoint is starting.
2. The gRPC server has started and is now accepting connections.
3. An incoming `registerConsumer` or `deregisterConsumer` request is being
   validated. This includes checking that the client has been authenticated, and
   that all required fields are provided and match the expected format.
4. The server is sending to the client a `registerFederatedService` or a
   `deregisterFederatedService` message.
5. The server is in waiting state. Transitioning out of the waiting state
   requires a cancellation (for example, because `deregisterConsumer` has been
   called) or an update because the server has a catalog update to share with
   the client. In the latter case, the update is sent.
6. The server is closing the stream and provides the client with an error code
   for the condition encountered.
7. The server has encountered a fatal condition and must stop. This can occur
   if there are data transmission errors.
8. The server has encountered a fatal condition and must stop. This can occur if
   the listener could not be created, or if the gRPC server encounters a fatal
   error.

### Client State Machine

![Figure 5](figure-5.png "Figure 5. Example implementation of the Service
Consumer (client)")

*Figure 5. Example implementation of the Service Consumer (client)*

1. The client is dialing the Federated Service Discovery API server endpoint.
2. The client is calling `registerConsumer` or `deregisterConsumer`.
3. The client is receiving data from the server stream, as a response to the
   possible calls to the server.
4. The client is validating and updating the local service catalog if the server
   has sent a `registerFederatedService`, `deregisterFederatedService` message,
   or answered a `registerConsumer`, or removing data from the local catalog if
   the server has answered a `deregisterConsumer` message sent from the client.
5. The client is in waiting state. Transitioning out of the waiting state
   requires a cancellation (for example, because `deregisterConsumer` has been
   called) or an update because the server has a catalog update to share with
   the client.
6. The client is performing an exponential backoff.
7. The client has encountered a fatal condition and must stop.
8. The client is performing an exponential backoff.
9. The client has encountered a fatal condition and must stop. This can occur if
   there are data transmission errors.

## Appendix B. List of Error Codes and Conditions

This section enumerates the various error codes that may be returned by a
Federated Service Discovery API endpoint implementation, the conditions under
which they may be returned, and how they should be handled. Please see the Error
Codes section and the [gRPC code package documentation](https://godoc.org/google.golang.org/grpc/codes)
for more information about these codes.

| Code | Condition | Client behavior |
| --- | --- | --- |
| `Unavailable` | The Federated Service Discovery API endpoint is unable to handle the request from the client | Retry with a backoff |
| `InvalidArgument` | Request data doesn’t contain all the required mandatory fields. Request data contains all the required mandatory fields but data is malformed. | Report error and don’t retry |
