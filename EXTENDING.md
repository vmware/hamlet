# Extending

This document describes constructs which you can use to enable federation over
your service mesh.

## Components

**Server** represents the component owned by a federated service mesh owner
  to service federated service mesh resource discovery requests. This component
  is responsible to identify and distribute federated resources to consumers.

**Client** represents the component owned by a federated service mesh consumer
  to request and consume federated resources using the federated service mesh
  resource discovery protocol. This component is responsible for subscribing to
  events related to federated resources and use those to configure the service
  mesh to enable federation.

You can find concrete extension samples in the [examples](examples/) directory.

## The Server

### Creating a Server Instance

You can create an instance of the server, like so:

```go
tlsConfig := tls.PrepareServerConfig(rootCACerts, peerCert, peerKey)
srv, err := server.NewServer(port, tlsConfig, provider)
```

We shall look at the `provider` parameter further down below.

### Starting the Server

You can start the server by executing the following blocking call.

```go
err := srv.Start()
```

### Publishing Federated Resources

You can publish resource events like so:

```go
err := srv.Resources().Create(res) // Publish a create event
err := srv.Resources().Update(res) // Publish a update event
err := srv.Resources().Delete(res) // Publish a delete event
```

Note that the resource to be published must be a Protobuf type. For example, you
can publish `FederatedService` resources like so:

```go
err := srv.Resources().Create(&sd.FederatedService{
        Name: "Service Name",
	...
})
```

The server will publish the resource event to all consumers that have subscribed
to the resource. In the case of the `FederatedService`, the event would have 
been published to all consumers that might have subscribed to
"type.googleapis.com/federation.types.v1alpha1.FederatedService" prior to the
call.

### Providing the State of Resources

Once a resource event is published, it is forgotten by the owner. However, there
may be federated service mesh consumers that connect with the owner at a later
point in time. At a minimum, such a consumer would need to know the status of
the system at that point in time. In order to accomodate this use-case, the
server module provides a `StateProvider` interface.

You can create a provider implementation, like so:

```go
type myProvider struct {
	state.StateProvider
}

func (p *myProvider) GetState(resourceUrl string) ([]proto.Message, error) {
	// Query the service mesh and return an array of federated resources of
	// the type resourceUrl.

	// For example, if the resourceUrl said
	// "type.googleapis.com/federation.types.v1alpha1.FederatedService",
	// an array of available FederatedService instances should be returned.

	return resources, nil
}
```

This state provider needs to be supplied to the `server.NewServer` call as shown
above. To do that, first create the `provider` instance, like so:

```go
provider := &myProvider{}
```

### Stopping the Server

You can stop the server, like so:

```go
err := srv.Stop()
```

## The Client

### Creating a Client Instance

You can create a client instance, like so:

```go
tlsConfig := tls.PrepareClientConfig(rootCACert, peerCert, peerKey, insecureSkipVerify)
cl, err := client.NewClient("server-address:port", tlsConfig)
```

### Watching FederatedService Events

You can watch `FederatedService` events by first implementing the
`FederatedServiceObserver` interface, like so:

```go
type myObserver struct {
        client.FederatedServiceObserver
}

func (o *myObserver) OnCreate(fs *sd.FederatedService) error {
	// Notify the service mesh that a new FederatedService is available.
}

func (o *myObserver) OnUpdate(fs *sd.FederatedService) error {
	// Notify the service mesh that a FederatedService has been updated.
}

func (o *myObserver) OnDelete(fs *sd.FederatedService) error {
	// Notify the service mesh that a FederatedService is no longer
	// available.
}
```

This implementation can then be used for subscribing to events related to
federated services using the `WatchFederatedServices` blocking call, like so:

```go
err := cl.WatchFederatedServices(ctx, &myObserver{})
```

### Watching Generic Resource Events

You can also watch events related to other federated resources using the generic
observer API, like so:

```go
type genericObserver struct {
	client.ResourceObserver
}

func (o *genericObserver) OnCreate(res *any.Any) error {
	// Notify the service mesh that a new resource is available.
}

func (o *genericObserver) OnUpdate(res *any.Any) error {
	// Notify the service mesh that a resource has been updated.
}

func (o *genericObserver) OnDelete(res *any.Any) error {
	// Notify the service mesh that a resource is no longer available.
}
```

This implementation can then be used for subscribing to generic resouce events
using the `WatchResources` blocking call. You can subscribe to events related to
the "type.googleapis.com/some.package.Resource" resource type, like so:

```go
err := cl.WatchResources(ctx, "type.googleapis.com/some.package.Resource", &genericObserver{})
```
