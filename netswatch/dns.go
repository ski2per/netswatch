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
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	consul "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

type DNSRegistry struct {
	Agent          consul.Agent
	Endpoint       string
	Token          string
	NetworkName    string
	OrgName        string
	NodeName       string
	NetdataEnabled bool
	NetdataPort    int
	HostIP         net.IP
}

func formatServiceString(s string) string {
	/*
		Input string and return:
		abc         -> abc
		abc.com     -> abc-com
		abc..com    -> abc-com
		abc_com     -> abc-com
		c.com_cn  	-> abc-com-cn
	*/
	replaced := regexp.MustCompile(`\.+|_+`)
	return replaced.ReplaceAllString(s, "-")
}

func (dnsr *DNSRegistry) InitAgent() {
	cli, err := consul.NewClient(&consul.Config{
		Address: dnsr.Endpoint,
		Token:   dnsr.Token,
	})
	if err != nil {
		panic(err)
	}
	agent := cli.Agent()
	dnsr.Agent = *agent
}

func (dnsr *DNSRegistry) listSvcIDs() []string {
	// Get service IDs from Consul
	filter := fmt.Sprintf("Tags contains \"%s\" and Tags contains \"%s\"", dnsr.OrgName, dnsr.NodeName)
	svcs, err := dnsr.Agent.ServicesWithFilter(filter)
	if err != nil {
		log.Error(err)
	}

	svcIDs := make([]string, len(svcs))

	i := 0
	for id := range svcs {
		svcIDs[i] = id
		i++
	}

	return svcIDs
}

func (dnsr *DNSRegistry) registerSvc(ctr *types.ContainerJSON) {
	var svc consul.AgentServiceRegistration
	svc.ID = ctr.ID
	svc.Name = fmt.Sprintf("%s-%s-%s", formatServiceString(dnsr.OrgName), dnsr.NodeName, getCtrName(ctr))
	svc.Address = ctr.NetworkSettings.Networks[dnsr.NetworkName].IPAddress
	svc.Tags = []string{dnsr.OrgName, dnsr.NodeName}

	// Extend service with Netdata data when NW_NETDATA_ENABLED is true
	if dnsr.NetdataEnabled && strings.Contains(svc.Name, "netdata") {
		svc.Tags = append(svc.Tags, "netdata")
		svc.Port = dnsr.NetdataPort
		svcMeta := map[string]string{
			"host":    dnsr.NodeName,
			"host_ip": dnsr.HostIP.String(),
		}
		svc.Meta = svcMeta
	}

	regErr := dnsr.Agent.ServiceRegister(&svc)
	if regErr != nil {
		log.Error(regErr)
	}

}

func (dnsr *DNSRegistry) deregisterSvc(id string) {
	deregErr := dnsr.Agent.ServiceDeregister(id)
	if deregErr != nil {
		log.Errorf("!!! Error deregistering: <%s>", id)
	}

}
