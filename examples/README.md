# Examples

The [server](server/) and [client](client/) examples showcase federated service
discovery using the service mesh federation - resource discovery protocol.

## Usage

1\. Generate root, server, and client certificates.

```console
$ make certs
```

2\. Start the server.

```console
$ make start-server
```

3\. Open another terminal session/window.

4\. Start the client.

```console
$ make start-client
```

5\. Hit ^c (ctrl + c) to terminate the server/client.