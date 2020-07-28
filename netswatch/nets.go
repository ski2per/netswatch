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

package netswatch

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	log "github.com/golang/glog"

	"github.com/coreos/flannel/pkg/ip"
	"github.com/coreos/flannel/subnet"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type IP uint32

type NodeMeta struct {
	OrgName  string
	NodeType string
	NodeName string
	HostIP   net.IP
}

func Hello() {
	fmt.Println("一哭二闹三上悠亚")
}

func createBridge(ctx context.Context, brName string, sn ip.IP4Net) {
	// 10.66.66.0 -> 10.66.66.1
	sn.IP++
	subnet := sn.String()

	ipamConfig := network.IPAMConfig{
		Subnet:  subnet,
		Gateway: sn.IP.String(),
	}

	ipam := network.IPAM{
		Driver: "default",
		Config: []network.IPAMConfig{ipamConfig},
	}

	nc := types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
		EnableIPv6:     false,
		Internal:       false,
		Attachable:     true,
		IPAM:           &ipam,
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	ncResp, err := cli.NetworkCreate(ctx, brName, nc)
	// Network with "brName" exists, check its IPAM
	if err != nil {
		log.Info(err)
		nr, _ := cli.NetworkInspect(ctx, brName)
		runningSubnet := nr.IPAM.Config[0].Subnet

		if runningSubnet == subnet {
			log.Infof("Bridge network <%v> synchronized with Flannel", brName)
		} else {
			log.Error("!!! Bridge is not synchronized")
			// Add a "god mod" later,
			// force remove bridge network even with containers in it.
		}
	} else {
		log.Infof("Bridge network <%v> created with ID: <%v>", brName, ncResp.ID)
	}
}

func WatchNets(ctx context.Context, sm subnet.Manager, sn ip.IP4Net, netName string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("done netswatch")
			return
		default:
			fmt.Println("watching nets")
			createBridge(ctx, netName, sn)

			// leases, err := sm.GetSubnets(ctx)
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Println(len(leases))
			// for _, lease := range leases {
			// 	fmt.Printf("%v\n", lease.Attrs.PublicIP)
			// 	fmt.Printf("%v\n", lease.Attrs.BackendType)

			// 	mac := struct{ VtepMac string }{}

			// 	if err := json.Unmarshal(lease.Attrs.BackendData, &mac); err != nil {
			// 		panic(err)
			// 	}
			// 	fmt.Println(mac.VtepMac)

			// 	var m NodeMeta
			// 	if err := json.Unmarshal(lease.Attrs.Meta, &m); err != nil {
			// 		panic(err)
			// 	}
			// 	fmt.Printf("%v\n", m.HostIP)
			// 	fmt.Printf("%v\n", m.NodeName)
			// 	fmt.Println("-----------------------------")

			// }
			time.Sleep(30 * time.Second)
		}

	}
}

func ExtendNodeMeta(meta *NodeMeta) *NodeMeta {
	// If meta.NodeName is not set, then use hostname for node name.
	if len(meta.NodeName) == 0 {
		name, err := os.Hostname()
		if err != nil {
			fmt.Println("get hostname error")
			fmt.Printf("%v", err)
			name = "default-node"
		}
		meta.NodeName = name
	}

	return meta
}
