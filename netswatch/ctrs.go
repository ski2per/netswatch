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
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

func getCtrName(ctr *types.ContainerJSON) string {
	labels := ctr.Config.Labels

	if v, ok := labels["com.docker.compose.service"]; ok {
		// Container run by docker-compose
		return v
	} else if v, ok := labels["com.docker.stack.namespace"]; ok {
		// Container run by docker stack deploy (Docker Swarm)
		stack := v
		svc, _ := labels["com.docker.swarm.service.name"]
		return strings.TrimPrefix(svc, stack+"_") // trim "_"
	} else {
		// Container run by "docker run"
		return strings.TrimPrefix(ctr.Name, "/") // trim prefix "/"
	}
}

func listJoinedCtrs(ctx context.Context, name string) map[string]types.ContainerJSON {
	// List containers which joined Netswatch bridge network
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	nr, err := cli.NetworkInspect(ctx, name)
	if err != nil {
		log.Error(err)
	}

	containers := make(map[string]types.ContainerJSON)

	for cID := range nr.Containers {
		ctr, err := cli.ContainerInspect(ctx, cID)
		if err != nil {
			panic(nil)
		}
		containers[cID] = ctr
	}
	return containers
}

func listCtrIDs(m map[string]types.ContainerJSON, networkName string) []string {
	// Extract IDs of containers which joined in "networkName"
	// The returned ID string will be formatted as below:
	// cb1a99ed906023223399e64f98f1d136e8536d365141fc59abffac2152340785-10.68.128.2
	IDs := make([]string, len(m))

	i := 0
	for id, ctr := range m {
		addr := ctr.NetworkSettings.Networks[networkName].IPAddress
		IDs[i] = fmt.Sprintf("%s-%s", id, addr)
		i++
	}
	return IDs
}
func extractCtrID(fmtID string) string {
	// Input: cb1a99ed906023223399e64f98f1d136e8536d365141fc59abffac2152340785-10.68.128.2
	// Output: cb1a99ed906023223399e64f98f1d136e8536d365141fc59abffac2152340785
	return strings.Split(fmtID, "-")[0]
}

func syncContainers(ctx context.Context, dns DNSRegistry) {
	// Get service IDs in Consul
	svcIDs := dns.listSvcIDs()
	log.Debug("svcIDs length: ", len(svcIDs))

	containers := listJoinedCtrs(ctx, dns.NetworkName)
	ctrIDs := listCtrIDs(containers, dns.NetworkName)

	remoteSet := NewSet()
	remoteSet.AddList(&svcIDs)

	localSet := NewSet()
	localSet.AddList(&ctrIDs)

	svc2Register := localSet.Difference(remoteSet)
	log.Debug("No. of services to register: ", svc2Register.Size())
	for id := range svc2Register.content {
		realID := extractCtrID(id)
		ctrJSON := containers[realID]
		// Get service addr for logging
		addr := ctrJSON.NetworkSettings.Networks[dns.NetworkName].IPAddress
		log.Infof("Registering service: <%s>(%s: %s)", ctrJSON.Name, realID, addr)
		dns.registerSvc(&ctrJSON)
	}
	svc2Deregister := remoteSet.Difference(localSet)
	log.Debug("No. of services to deregister: ", svc2Deregister.Size())
	for id := range svc2Deregister.content {
		realID := extractCtrID(id)
		log.Infof("Deregistering service: <%s>", realID)
		dns.deregisterSvc(realID)
	}
}

// WatchCtrEvents is the function for watch container network releated event,
// and register/deregister containers as services in Consul.
func WatchCtrEvents(ctx context.Context, dns DNSRegistry) {
	// Main func for watching
	log.Info("c[_] CONTAINERS' EVENTS WATCH BEGINS")

	filter := filters.NewArgs()
	// Watch Docker events with type: "network"
	filter.Add("type", "network")
	// Only watch events below
	filter.Add("event", "connect")
	filter.Add("event", "disconnect")

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// Ignore error channel
	evtCh, _ := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})

	netName := dns.NetworkName

	for evt := range evtCh {
		evtNetName := evt.Actor.Attributes["name"]
		if evtNetName == netName {
			log.Info("c[_] GOT NETWORK connect/disconnect EVENT")
			syncContainers(ctx, dns)
		}
	}
}

// WatchCtrs is a function for shnchronizing containers periodically
func WatchCtrs(ctx context.Context, dns DNSRegistry, maxLoop int) {
	sleep := 1
	for {
		if sleep < maxLoop {
			sleep *= 2
		} else {
			sleep = 1
		}

		select {
		case <-ctx.Done():
			log.Info("c[_] CONTAINERS' WATCH IS ENDED")
			return
		default:
			log.Info("c[_] CONTAINERS' WATCH BEGINS")
			syncContainers(ctx, dns)
			time.Sleep(time.Duration(sleep) * time.Second)
		}

	}
}
