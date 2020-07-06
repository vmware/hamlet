// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// defaultSyncPeriod is the default period for reloading TLS certificates from disk
	defaultSyncPeriod = 10 * time.Minute
)

// PeriodicCertLoader periodically reloads the TLS keypair from disk
type PeriodicCertLoader struct {
	sync.RWMutex

	certPath string
	keyPath  string

	syncPeriod time.Duration

	current *tls.Certificate
}

// NewPeriodicCertLoader returns an instance of PeriodicCertLoader given the cert/key file path
func NewPeriodicCertLoader(certPath, keyPath string, syncPeriod time.Duration) (*PeriodicCertLoader, error) {
	if syncPeriod == 0 {
		syncPeriod = defaultSyncPeriod
	}

	c := &PeriodicCertLoader{
		certPath:   certPath,
		keyPath:    keyPath,
		syncPeriod: syncPeriod,
	}

	// initial load of certificates
	if err := c.Reload(); err != nil {
		return nil, err
	}

	return c, nil
}

// Start starts the period sync loop for reloading certificates
func (c *PeriodicCertLoader) Start() {
	ticker := time.NewTicker(c.syncPeriod)
	for range ticker.C {
		if err := c.Reload(); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Errorf("error reloading certificate from disk")
		}
	}
}

// Reload loads the TLS key pair based on the certificate paths
func (c *PeriodicCertLoader) Reload() error {
	keyPair, err := tls.LoadX509KeyPair(c.certPath, c.keyPath)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.current = &keyPair

	return nil
}

// GetCertificate returns the current TLS certificate last read from disk.
// Call this function from tls.Config.GetCertificate or tls.Config.GetClientCertificate
func (c *PeriodicCertLoader) Current() *tls.Certificate {
	c.RLock()
	defer c.RUnlock()

	return c.current
}
