// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

type AviSyncSecret struct {
	AviUsername      string
	AviPassword      string
	AviTenant        string
	HamletToken      string
	HamletServerCert []byte
}

func (a *AviSyncSecret) CompareEqual(b *AviSyncSecret) bool {
	return a.AviUsername == b.AviUsername &&
		a.AviPassword == b.AviPassword &&
		a.AviTenant == b.AviTenant &&
		string(a.HamletServerCert) == string(b.HamletServerCert) &&
		a.HamletToken == b.HamletToken
}
