// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func Test_PeriodicReload(t *testing.T) {
	rootCAFile, err := ioutil.TempFile("", "rootCA.crt")
	if err != nil {
		t.Fatalf("error creating tmp root CA: %v", err)
	}
	defer os.Remove(rootCAFile.Name())
	if _, err := rootCAFile.Write([]byte(testRootCA)); err != nil {
		t.Fatalf("error writing to tmp root CA: %v", err)
	}

	serverCertFile, err := ioutil.TempFile("", "server.crt")
	if err != nil {
		t.Fatalf("error creating tmp server cert: %v", err)
	}
	defer os.Remove(serverCertFile.Name())
	if _, err := serverCertFile.Write([]byte(testServerCert)); err != nil {
		t.Fatalf("error writing to tmp server cert: %v", err)
	}

	serverKeyFile, err := ioutil.TempFile("", "server.key")
	if err != nil {
		t.Fatalf("error creating tmp server key: %v", err)
	}
	defer os.Remove(serverKeyFile.Name())
	if _, err := serverKeyFile.Write([]byte(testServerKey)); err != nil {
		t.Fatalf("error writing to tmp server key: %v", err)
	}

	certLoader, err := NewPeriodicCertLoader(serverCertFile.Name(), serverKeyFile.Name(), 500*time.Millisecond)
	if err != nil {
		t.Fatalf("error starting periodic cert reloader: %v", err)
	}

	serverConfig := &tls.Config{}
	serverConfig.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return certLoader.Current(), nil
	}
	clientConfig := &tls.Config{}
	clientConfig.RootCAs = prepareCertPool(rootCAFile.Name())

	// first check client -> server w/ TLS to validate initial reload worked
	testClientServerHello(t, clientConfig, serverConfig)

	// nil out the current certificate, should expect an error
	certLoader.current = nil
	testClientServerHelloError(t, clientConfig, serverConfig)

	// start certificate loader and sleep for next reload
	go certLoader.Start()
	time.Sleep(1 * time.Second)

	testClientServerHello(t, clientConfig, serverConfig)
}
