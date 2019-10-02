# Copyright 2019 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Constants
SHELL := /bin/bash

# Prints the help message
# Usage: make help
.PHONY: help
help:
	@echo "Usage: make <TARGET> [<ARGUMENTS>]"
	@echo ""
	@echo "Available targets are:"
	@echo ""
	@echo "    certs           Generate certificates"
	@echo "    clean           Remove all certificates"
	@echo "    start-client    Start the client"
	@echo "    start-server    Start the server"
	@echo ""

# Generate root certificate
.PHONY: _root-cert
_root-cert:
	openssl req \
		-newkey rsa:2048 \
		-new -nodes -x509 \
		-days 3650 \
		-out rootCA.crt \
		-keyout rootCA.key \
		-subj "/CN=localhost"

# Generate peer certificate
.PHONY: _peer-cert
_peer-cert:
	openssl req \
		-newkey rsa:2048 \
		-new -nodes -sha256 \
		-out $(PREFIX).csr \
		-keyout $(PREFIX).key \
		-subj "/CN=localhost"
	openssl x509 -sha256\
		-req -in $(PREFIX).csr \
		-CA rootCA.crt \
		-CAkey rootCA.key \
		-CAcreateserial \
		-out $(PREFIX).crt \
		-days 3650
	@rm $(PREFIX).csr
	@rm rootCA.srl

# Generate certificates
# Usage: make certs
.PHONY: certs
certs:
	$(MAKE) _root-cert
	$(MAKE) PREFIX=server _peer-cert
	$(MAKE) PREFIX=client _peer-cert
	@echo "Certificates generated successfully!"

# Remove all certificates
# Usage: make clean
.PHONY: clean
clean:
	rm -rf rootCA.key rootCA.crt
	rm -rf server.key server.crt
	rm -rf client.key client.crt

# Start the server
# Usage: make start-server
.PHONY: start-server
start-server:
	go run -v server/main.go start \
		--root-ca-cert rootCA.crt \
		--peer-cert server.crt \
		--peer-key server.key \
		--port 8001

# Start the client
# Usage: make start-client
.PHONY: start-client
start-client:
	go run -v client/main.go start \
		--root-ca-cert rootCA.crt \
		--peer-cert client.crt \
		--peer-key client.key \
		--insecure-skip-verify \
		--server-addr ":8001"
