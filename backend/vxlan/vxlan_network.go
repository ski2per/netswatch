// Copyright 2015 flannel authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// +build !windows

package vxlan

import (
	"encoding/json"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/context"

	"syscall"

	"github.com/coreos/flannel/backend"
	"github.com/coreos/flannel/pkg/ip"
	"github.com/coreos/flannel/subnet"
)

type network struct {
	backend.SimpleNetwork
	dev       *vxlanDevice
	subnetMgr subnet.Manager
}

const (
	// Guess: 	8bytes(vxlan header) +
	// 			8bytes(udp header) +
	// 			20bytes(ip header) +
	// 			14bytes(ethernet header, no inner FCS)
	// 			= 50 bytes
	encapOverhead = 50
)

func newNetwork(subnetMgr subnet.Manager, extIface *backend.ExternalInterface, dev *vxlanDevice, _ ip.IP4Net, lease *subnet.Lease) (*network, error) {
	nw := &network{
		SimpleNetwork: backend.SimpleNetwork{
			SubnetLease: lease,
			ExtIface:    extIface,
		},
		subnetMgr: subnetMgr,
		dev:       dev,
	}

	return nw, nil
}

func (nw *network) Run(ctx context.Context) {
	wg := sync.WaitGroup{}

	log.Info("watching for new subnet leases")
	events := make(chan []subnet.Event)
	wg.Add(1)
	go func() {
		subnet.WatchLeases(ctx, nw.subnetMgr, nw.SubnetLease, events)
		log.Info("WatchLeases exited")
		wg.Done()
	}()

	defer wg.Wait()

	for {
		select {
		case evtBatch := <-events:
			log.Info("c[_] GOT LEASE EVENT")
			nw.handleSubnetEvents(ctx, evtBatch)

		case <-ctx.Done():
			return
		}
	}
}

func (nw *network) MTU() int {
	return nw.ExtIface.Iface.MTU - encapOverhead
}

type vxlanLeaseAttrs struct {
	VtepMAC hardwareAddr
}

func (nw *network) handleSubnetEvents(ctx context.Context, batch []subnet.Event) {
	// Get router and their IP4Net
	routers := nw.subnetMgr.GetRouters(ctx)

	// Unmarshal Lease Meta
	meta := struct {
		OrgName  string
		NodeType string
	}{}

	//  Get current node's meta
	if err := json.Unmarshal(nw.SubnetLease.Attrs.Meta, &meta); err != nil {
		log.Error("error decoding current subnet lease Meta: ", err)
	}
	currentOrgName := meta.OrgName
	currentNodeType := meta.NodeType

	for _, event := range batch {
		sn := event.Lease.Subnet
		attrs := event.Lease.Attrs
		if attrs.BackendType != "vxlan" {
			log.Warningf("ignoring non-vxlan subnet(%s): type=%v", sn, attrs.BackendType)
			continue
		}

		var vxlanAttrs vxlanLeaseAttrs
		if err := json.Unmarshal(attrs.BackendData, &vxlanAttrs); err != nil {
			log.Error("error decoding subnet lease JSON: ", err)
			continue
		}

		// ====================================
		//              Netswatch
		// Add routing adjustment algorithm
		// (From Netswatch.py)
		// ------------------------------------

		if err := json.Unmarshal(event.Lease.Attrs.Meta, &meta); err != nil {
			log.Error("error decoding subnet lease Meta: ", err)
		}

		// ------------------------------------
		// ====================================

		vxlanRoute := netlink.Route{
			LinkIndex: nw.dev.link.Attrs().Index,
			Scope:     netlink.SCOPE_UNIVERSE,
			Dst:       sn.ToIPNet(),
		}

		if len(routers) <= 0 || currentNodeType == "internal" || meta.NodeType == "internal" || meta.OrgName == currentOrgName {
			log.Debug("++++++++++++++++ default route ++++++++++++++++")
			log.Debugf("Current node type: %s", currentNodeType)
			log.Debugf("Meta node type: %s", meta.NodeType)
			log.Debugf("Current org name: %s", currentOrgName)
			log.Debugf("Meta org name: %s", meta.OrgName)
			log.Debug("-----------------------------------------------")
			// No need to adjust route, use default Flannel route, i.e:
			// n0 via n0 dev nw.100 onlink
			// n1 via n1 dev nw.100 onlink
			// r0 via r0 dev nw.100 onlink
			// r1 via r1 dev nw.100 onlink
			vxlanRoute.Gw = sn.IP.ToIP()

		} else {
			// Need to adjust route, use target org's router as gateway, i.e:
			// n0 via r0 dev nw.100 onlink
			// n1 via r1 dev nw.100 onlink
			// r0 via r0 dev nw.100 onlink
			// r1 via r1 dev nw.100 onlink
			log.Debug("++++++++++++++++ adjust route ++++++++++++++++")
			log.Debugf("Current node type: %s", currentNodeType)
			log.Debugf("Meta node type: %s", meta.NodeType)
			log.Debugf("Current org name: %s", currentOrgName)
			log.Debugf("Meta org name: %s", meta.OrgName)
			log.Debug("------------------------------ ----------------")

			if currentNodeType == "node" {
				// Use current org's router as gateway
				if value, exists := routers[currentOrgName]; exists {
					vxlanRoute.Gw = value.IP.ToIP()
				} else {
					log.Errorf("!!! Find no router in org <%s>", meta.OrgName)
					vxlanRoute.Gw = sn.IP.ToIP()
				}
			} else {
				if value, exists := routers[meta.OrgName]; exists {
					vxlanRoute.Gw = value.IP.ToIP()
				} else {
					log.Errorf("!!! Find no router in org <%s>", meta.OrgName)
					vxlanRoute.Gw = sn.IP.ToIP()
				}
			}

		}

		// This route is used when traffic should be vxlan encapsulated
		vxlanRoute.SetFlag(syscall.RTNH_F_ONLINK)

		// directRouting is where the remote host is on the same subnet so vxlan isn't required.
		directRoute := netlink.Route{
			Dst: sn.ToIPNet(),
			Gw:  attrs.PublicIP.ToIP(),
		}
		var directRoutingOK = false
		if nw.dev.directRouting {
			if dr, err := ip.DirectRouting(attrs.PublicIP.ToIP()); err != nil {
				log.Error(err)
			} else {
				directRoutingOK = dr
			}
		}

		switch event.Type {
		case subnet.EventAdded:
			if directRoutingOK {
				log.Infof("Adding direct route to subnet: %s PublicIP: %s", sn, attrs.PublicIP)

				if err := netlink.RouteReplace(&directRoute); err != nil {
					log.Errorf("Error adding route to %v via %v: %v", sn, attrs.PublicIP, err)
					continue
				}
			} else {
				log.Infof("adding subnet: %s PublicIP: %s VtepMAC: %s", sn, attrs.PublicIP, net.HardwareAddr(vxlanAttrs.VtepMAC))
				if err := nw.dev.AddARP(neighbor{IP: sn.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
					log.Error("AddARP failed: ", err)
					continue
				}

				if err := nw.dev.AddFDB(neighbor{IP: attrs.PublicIP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
					log.Error("AddFDB failed: ", err)

					// Try to clean up the ARP entry then continue
					if err := nw.dev.DelARP(neighbor{IP: event.Lease.Subnet.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
						log.Error("DelARP failed: ", err)
					}

					continue
				}

				// Set the route - the kernel would ARP for the Gw IP address if it hadn't already been set above so make sure
				// this is done last.
				if err := netlink.RouteReplace(&vxlanRoute); err != nil {
					log.Errorf("failed to add vxlanRoute (%s -> %s): %v", vxlanRoute.Dst, vxlanRoute.Gw, err)

					// Try to clean up both the ARP and FDB entries then continue
					if err := nw.dev.DelARP(neighbor{IP: event.Lease.Subnet.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
						log.Error("DelARP failed: ", err)
					}

					if err := nw.dev.DelFDB(neighbor{IP: event.Lease.Attrs.PublicIP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
						log.Error("DelFDB failed: ", err)
					}

					continue
				}
			}
		case subnet.EventRemoved:
			if directRoutingOK {
				log.Infof("Removing direct route to subnet: %s PublicIP: %s", sn, attrs.PublicIP)
				if err := netlink.RouteDel(&directRoute); err != nil {
					log.Errorf("Error deleting route to %v via %v: %v", sn, attrs.PublicIP, err)
				}
			} else {
				log.Infof("removing subnet: %s PublicIP: %s VtepMAC: %s", sn, attrs.PublicIP, net.HardwareAddr(vxlanAttrs.VtepMAC))

				// Try to remove all entries - don't bail out if one of them fails.
				if err := nw.dev.DelARP(neighbor{IP: sn.IP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
					log.Error("DelARP failed: ", err)
				}

				if err := nw.dev.DelFDB(neighbor{IP: attrs.PublicIP, MAC: net.HardwareAddr(vxlanAttrs.VtepMAC)}); err != nil {
					log.Error("DelFDB failed: ", err)
				}

				if err := netlink.RouteDel(&vxlanRoute); err != nil {
					log.Errorf("failed to delete vxlanRoute (%s -> %s): %v", vxlanRoute.Dst, vxlanRoute.Gw, err)
				}
			}
		default:
			log.Error("internal error: unknown event type: ", int(event.Type))
		}
	}
}
