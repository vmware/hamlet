// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	aviClients "github.com/avinetworks/sdk/go/clients"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	hamletClient "github.com/vmware/hamlet/pkg/v1alpha2/client"
	"istio.io/pkg/log"
)

func (sh *aviSyncHandler) createAviSync(ctx context.Context, ctxCancel context.CancelFunc, hClient hamletClient.Client, aClient *aviClients.AviClient) <-chan error {
	// connect to avi services
	// discover all virtual services in avi
	// pass the virtual service as entry to hamlet
	retCh := make(chan error)
	go func() {
		loopCount := 0
		done := false
		for {
			cv, err := aClient.AviSession.GetControllerVersion()

			if err != nil {
				sh.log.Info("AVI Controller ", "iteration", loopCount, "Version", cv, "Error", err.Error())
			} else {
				sh.log.Info("AVI Controller ", "iteration", loopCount, "Version", cv)
			}
			vs, err := aClient.VirtualService.GetAll()
			if err != nil {
				sh.log.Error(err, "Virtual service get from avi resulted in error")
			} else {
				svcListAdd := make(map[string]bool)
				for _, vsItm := range vs {
					sh.log.Info("Got Virtual Service",
						"Name", vsItm.Name,
						"UUID", vsItm.UUID,
						"TrafficEnabled", vsItm.TrafficEnabled,
						"PoolRef", vsItm.PoolRef,
					)

					// get virtual service
					// form vs get /vsvip/{uuid} it has vip public/private ip address
					// select public and if not present use private
					// get /applicationprofile/{uuid} it has application type
					//  https = APPLICATION_PROFILE_TYPE_SSL , http = APPLICATION_PROFILE_TYPE_HTTP
					// get /pool/{uuid}
					//   it has "default_server_port" and list of servers with port
					// get /sslkeyandcertificate/{uuid}
					//   it has certificate associated with the service
					EPPortSSL := int32(0)
					EPPort := int32(0)
					for _, svcItm := range vsItm.Services {
						if *svcItm.EnableSsl {
							EPPortSSL = *svcItm.Port
						} else {
							EPPort = *svcItm.Port
						}
					}

					// get the endpoints for the service
					publicEP := []string{}
					privateEP := []string{}
					svcDNS := ""
					if vsItm.VsvipRef != nil {
						uuid_ := strings.Split(*vsItm.VsvipRef, "/")
						uuid := uuid_[len(uuid_)-1]
						vsvip, err := aClient.VsVip.Get(uuid)
						if err != nil {
							sh.log.Error(err, "Virtual service VsVip get from avi resulted in error")
						} else {
							for _, vipInst := range vsvip.Vip {
								if vipInst.IPAddress != nil {
									sh.log.Info("             ",
										"vipId", *vipInst.VipID,
										"ip_address", *vipInst.IPAddress.Addr,
										"ip_address_type", *vipInst.IPAddress.Type)
								}
								if vipInst.FloatingIP != nil {
									sh.log.Info("             ",
										"vipId", *vipInst.VipID,
										"floating_ip", *vipInst.FloatingIP.Addr,
										"floating_ip_type", *vipInst.FloatingIP.Type)
								}
								if vipInst.IPAddress != nil && *vipInst.IPAddress.Type == "V4" {
									privateEP = append(privateEP, *vipInst.IPAddress.Addr)
								}
								if vipInst.FloatingIP != nil && *vipInst.FloatingIP.Type == "V4" {
									publicEP = append(publicEP, *vipInst.FloatingIP.Addr)
								}
							}
							for _, dnsInst := range vsvip.DNSInfo {
								if dnsInst.Fqdn != nil {
									svcDNS = *dnsInst.Fqdn
								}
							}
						}
					}
					// get the protocol
					protocol := ""
					if vsItm.ApplicationProfileRef != nil {
						uuid_ := strings.Split(*vsItm.ApplicationProfileRef, "/")
						uuid := uuid_[len(uuid_)-1]
						vsapp, err := aClient.ApplicationProfile.Get(uuid)
						if err != nil {
							sh.log.Error(err, "Virtual service ApplicationProfile get from avi resulted in error")
						} else {
							if *vsapp.Type == "APPLICATION_PROFILE_TYPE_HTTP" {
								protocol = "HTTP"
							} else if *vsapp.Type == "APPLICATION_PROFILE_TYPE_SSL" {
								protocol = "HTTPS"
							}
							sh.log.Info("             ",
								"Type", *vsapp.Type,
								"Inferred Protocol", protocol)
						}
					}
					// get the port for the application
					svcPort := int32(0)
					if vsItm.PoolRef != nil {
						uuid_ := strings.Split(*vsItm.PoolRef, "/")
						uuid := uuid_[len(uuid_)-1]
						vspool, err := aClient.Pool.Get(uuid)
						if err != nil {
							sh.log.Error(err, "Virtual service ApplicationProfile get from avi resulted in error")
						} else {
							svcPort = *vspool.DefaultServerPort
							if svcPort == 0 && len(vspool.Servers) > 0 && vspool.Servers[0].Port != nil {
								svcPort = *vspool.Servers[0].Port
							}
							sh.log.Info("             ",
								"DefaultServerPort", *vspool.DefaultServerPort,
								"Inferred Port", svcPort)
						}
					}
					// get the certificate
					cert := ""
					for _, sslKey := range vsItm.SslKeyAndCertificateRefs {
						uuid_ := strings.Split(sslKey, "/")
						uuid := uuid_[len(uuid_)-1]
						vsSSL, err := aClient.SSLKeyAndCertificate.Get(uuid)
						if err != nil {
							sh.log.Error(err, "Virtual service SSLKeyAndCertificate get from avi resulted in error")
						} else {
							if vsSSL.Certificate != nil {
								cert = *vsSSL.Certificate.Certificate
							}
							sh.log.Info("             ",
								"name", *vsSSL.Name,
								"type", *vsSSL.Type,
								"Certificate ", cert,
							)
						}
					}

					// update to the local registry
					if *vsItm.TrafficEnabled {
						svcFqdn := *vsItm.Name + ".external"
						if svcDNS != "" {
							svcFqdn = svcDNS
						}
						gwPort := EPPort
						if protocol == "HTTPS" {
							gwPort = EPPortSSL
						}
						sh.log.Info("             ",
							"gwPort", gwPort,
							"svcPort", svcPort,
						)
						var ep []*types2.FederatedService_Endpoint
						if len(publicEP) > 0 {
							// use public endpoints
							for _, pubEP := range publicEP {
								ep = append(ep, &types2.FederatedService_Endpoint{
									Address: pubEP,
									Port:    uint32(gwPort),
								})
							}
						} else {
							// user private endpoint only if public is  not available
							for _, privEP := range privateEP {
								ep = append(ep, &types2.FederatedService_Endpoint{
									Address: privEP,
									Port:    uint32(gwPort),
								})
							}
						}
						var ins []*types2.FederatedService_Instance
						ins = append(ins, &types2.FederatedService_Instance{
							Id:       "inst-pool-1",
							Protocol: protocol,
							Metadata: map[string]string{"port": fmt.Sprint(svcPort)},
						})
						if cert != "" {
							ins[0].Metadata["cert"] = cert
						}
						svc := &types2.FederatedService{
							Name:      svcFqdn,
							Fqdn:      svcFqdn,
							Endpoints: ep,
							Instances: ins,
						}
						if len(svc.Endpoints) > 0 && svcPort != 0 && gwPort != 0 {
							svcListAdd[svcFqdn] = true
							if err := hClient.Upsert(svc.Fqdn, svc); err != nil {
								log.Error(err, "Error occurred while creating service", "svc", svc)
							}
							log.Info("HamletClient: Created a resource", "resource", svc.Fqdn)
						}
					}
				}
				// remove the items that did not get updated from the registry
				rlist := hClient.GetAllLocalResourceIDs("type.googleapis.com/federation.types.v1alpha2.FederatedService")
				for _, ritm := range rlist {
					if _, ok := svcListAdd[ritm]; !ok {
						if err := hClient.Delete(ritm); err != nil {
							log.Error(err, "Error occurred while Deleteing resource", "id", ritm)
						}
					}
				}
			}

			// for any error we can call ctxCancel()
			select {
			case <-ctx.Done():
				done = true
			case <-time.After(sh.retryTimeInterval):
				break
			}
			if done {
				break
			}
			loopCount++
		}
		retCh <- nil
	}()
	return retCh
}
