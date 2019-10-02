// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

// TODO: Add tests for the tls package.

// prepareCertPool prepares a TLS cert pool given a slice of CA certificates.
func prepareCertPool(caCerts ...string) *x509.CertPool {
	certPool := x509.NewCertPool()
	for _, caCert := range caCerts {
		certFile, err := ioutil.ReadFile(caCert)
		if err != nil {
			log.WithFields(log.Fields{
				"caCert": caCert,
				"err":    err,
			}).Fatalln("Error loading CA cert file")
		}
		certPool.AppendCertsFromPEM(certFile)
	}
	return certPool
}

// prepareCommonConfig prepares a TLS config instance which is common across
// server and client implementations that want mTLS enabled.
func prepareCommonConfig(peerCert string, peerKey string) *tls.Config {
	// Load the peer cert and key files.
	peerPair, err := tls.LoadX509KeyPair(peerCert, peerKey)
	if err != nil {
		log.WithField("err", err).Fatalln("Error loading peer cert/key")
	}

	// Build the TLS credentials.
	config := &tls.Config{
		Certificates: []tls.Certificate{peerPair},
	}
	return config
}

// PrepareServerConfig prepares a TLS config instance for a server that wants
// mTLS enabled.
func PrepareServerConfig(rootCACerts []string, peerCert string, peerKey string) *tls.Config {
	config := prepareCommonConfig(peerCert, peerKey)
	config.ClientCAs = prepareCertPool(rootCACerts...)
	config.ClientAuth = tls.RequireAndVerifyClientCert
	return config
}

// PrepareClientConfig prepares a TLS config instance for a client that wants
// mTLS enabled.
func PrepareClientConfig(rootCACert string, peerCert string, peerKey string, insecureSkipVerify bool) *tls.Config {
	config := prepareCommonConfig(peerCert, peerKey)
	config.RootCAs = prepareCertPool(rootCACert)
	config.InsecureSkipVerify = insecureSkipVerify
	return config
}
