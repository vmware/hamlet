/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

/*
	{LastModified:0xc00029bbb0
		ActiveStandbySeTag:0xc00029bbc0
		AdvertiseDownVs:0xc00062740c
		AllowInvalidClientCert:0xc00062740d
		AnalyticsPolicy:0xc0003e6840
		AnalyticsProfileRef:0xc00029bbf0
		ApicContractGraph:<nil>
		ApplicationProfileRef:0xc00029bc00
		AutoAllocateFloatingIP:<nil>
		AutoAllocateIP:<nil>
		AvailabilityZone:<nil>
		AviAllocatedFip:<nil> AviAllocatedVip:<nil> AzureAvailabilitySet:<nil> BotPolicyRef:<nil>
		BulkSyncKvcache:0xc0006274c2
		ClientAuth:<nil>
		CloseClientConnOnConfigUpdate:0xc0006274c3
		CloudConfigCksum:<nil>
		CloudRef:0xc00029bc10
		CloudType:0xc00029bc20
		ConnectionsRateLimit:<nil>
		ContentRewrite:0xc0006105f0
		CreatedBy:<nil>
		DelayFairness:0xc0006274cf
		Description:<nil>
		DiscoveredNetworkRef:[] DiscoveredNetworks:[] DiscoveredSubnet:[] DNSInfo:[] DNSPolicies:[]
		EastWestPlacement:0xc0006274d0
		EnableAutogw:0xc0006274d1 EnableRhi:<nil> EnableRhiSnat:<nil>
		Enabled:0xc0006274d2 ErrorPageProfileRef:<nil> FloatingIP:<nil> FloatingSubnetUUID:<nil>
		FlowDist:0xc00029bc40
		FlowLabelType:0xc00029bc50 Fqdn:<nil> HostNameXlate:<nil>
		HTTPPolicies:[0xc00029bc60] IcapRequestProfileRefs:[]
		IgnPoolNetReach:0xc0006274ee IPAddress:<nil> IPAMNetworkSubnet:<nil> JwtConfig:<nil> L4Policies:[] Labels:[]
		LimitDoser:0xc0006274ef
		MaxCpsPerClient:0xc0006274f0 MicroserviceRef:<nil> MinPoolsUp:<nil>
		Name:0xc00029bc80
		NetworkProfileRef:0xc00029bc90 NetworkRef:<nil>
		NetworkSecurityPolicyRef:0xc00029bca0 NsxSecuritygroup:[] PerformanceLimits:<nil> PoolGroupRef:<nil>
		PoolRef:0xc00029bcb0 PortUUID:<nil> RemoveListeningPortOnVsDown:0xc0006274fb RequestsRateLimit:<nil> SamlSpConfig:<nil>
		ScaleoutEcmp:0xc0006274fc
		SeGroupRef:0xc00029bcc0 SecurityPolicyRef:<nil> ServerNetworkProfileRef:<nil> ServiceMetadata:<nil> ServicePoolSelect:[]
		Services:[0xc00069a7e0]
		SidebandProfile:0xc000567ae0 SnatIP:[] SpPoolRefs:[] SslKeyAndCertificateRefs:[] SslProfileRef:<nil> SslProfileSelectors:[]
		SslSessCacheAvgSize:0xc000627518 SsoPolicy:<nil> SsoPolicyRef:<nil> StaticDNSRecords:[] Subnet:<nil> SubnetUUID:<nil> TenantRef:0xc00029bcd0 TestSeDatastoreLevel1Ref:<nil> TopologyPolicies:[] TrafficCloneProfileRef:<nil>
		TrafficEnabled:0xc000627520
		Type:0xc00029bce0
		URL:0xc00029bcf0
		UseBridgeIPAsVip:0xc000627530
		UseVipAsSnat:0xc000627531
		UUID:0xc00029bd00 VhDomainName:[] VhMatches:[] VhParentVsUUID:<nil> VhType:0xc00029bd10
		Vip:[] VrfContextRef:0xc00029bd20 VsDatascripts:[] VsvipCloudConfigCksum:<nil>
		VsvipRef:0xc00029bd30 WafPolicyRef:<nil>
		Weight:0xc000627540}
*/
