// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

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

// prepareServerConfigWithPeriodicReload prepares a server TLS config instance that periodically
// reloads the certificates from disk
func prepareServerConfigWithPeriodicReload(peerCert string, peerKey string, syncPeriod time.Duration) *tls.Config {
	certLoader, err := NewPeriodicCertLoader(peerCert, peerKey, syncPeriod)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatalln("Error starting TLS certificate reloader")
	}
	go certLoader.Start()

	config := &tls.Config{}
	config.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return certLoader.Current(), nil
	}
	return config
}

// prepareClientConfigWithPeriodicReload prepares a client TLS config instance that periodically
// reloads the certificates from disk
func prepareClientConfigWithPeriodicReload(peerCert string, peerKey string, syncPeriod time.Duration) *tls.Config {
	certLoader, err := NewPeriodicCertLoader(peerCert, peerKey, syncPeriod)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatalln("Error starting TLS certificate reloader")
	}
	go certLoader.Start()

	config := &tls.Config{}
	config.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return certLoader.Current(), nil
	}
	return config
}

// PrepareServerConfig prepares a TLS config instance for a server that wants mTLS enabled.
func PrepareServerConfig(rootCACerts []string, peerCert string, peerKey string) *tls.Config {
	config := prepareCommonConfig(peerCert, peerKey)
	config.ClientCAs = prepareCertPool(rootCACerts...)
	config.ClientAuth = tls.RequireAndVerifyClientCert
	return config
}

// PrepareServerConfigPeriodicWithReload prepares a TLS config instance for a server that wants mTLS enabled
// and dynamic reloading of server certificates.
func PrepareServerConfigWithPeriodicReload(rootCACerts []string, peerCert string, peerKey string, syncPeriod time.Duration) *tls.Config {
	config := prepareServerConfigWithPeriodicReload(peerCert, peerKey, syncPeriod)
	config.ClientCAs = prepareCertPool(rootCACerts...)
	config.ClientAuth = tls.RequireAndVerifyClientCert
	return config
}

// PrepareClientConfig prepares a TLS config instance for a client that wants mTLS enabled.
func PrepareClientConfig(rootCACert string, peerCert string, peerKey string, insecureSkipVerify bool) *tls.Config {
	config := prepareCommonConfig(peerCert, peerKey)
	config.RootCAs = prepareCertPool(rootCACert)
	config.InsecureSkipVerify = insecureSkipVerify
	return config
}

// PrepareClientConfigWithPeriodicReload prepares a TLS config instance for a client that wants mTLS enabled
// and dynamic reloading of client certificates.
func PrepareClientConfigWithPeriodicReload(rootCACert string, peerCert string, peerKey string, insecureSkipVerify bool, syncPeriod time.Duration) *tls.Config {
	config := prepareClientConfigWithPeriodicReload(peerCert, peerKey, syncPeriod)
	config.RootCAs = prepareCertPool(rootCACert)
	config.InsecureSkipVerify = insecureSkipVerify
	return config
}
