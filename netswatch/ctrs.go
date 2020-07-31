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
	log "github.com/golang/glog"
)

func getCtrName(ctr *types.ContainerJSON) string {
	labels := ctr.Config.Labels

	// When container run by docker-compose or swarm,
	// labels map will not be empty
	if len(labels) > 0 {
		// docker-compose
		if v, ok := labels["com.docker.compose.service"]; ok {
			return v
		}
		//Docker Swarm
		stack, _ := labels["com.docker.stack.namespace"]
		svc, _ := labels["com.docker.swarm.service.name"]
		return strings.TrimPrefix(svc, stack+"_") // trim "_"
	}
	// container by "docker run"
	return strings.TrimPrefix(ctr.Name, "/") // trim prefix "/"
}

func listJoinedCtrs(ctx context.Context, name string) map[string]types.ContainerJSON {
	// List containers which joined Netswatch bridge network
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	nr, err := cli.NetworkInspect(ctx, name)
	if err != nil {
		panic(err)
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

func listContainers(ctx context.Context) {
	// listCtrInNetwork(ctx)

	// cli, err := client.NewEnvClient()
	// if err != nil {
	// 	panic(err)
	// }

	// containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	// if err != nil {
	// 	panic(err)
	// }

	// for _, container := range containers {
	// 	fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	// }

}

func extractCtrIDs(m map[string]types.ContainerJSON) []string {
	IDs := make([]string, len(m))

	i := 0
	for k := range m {
		IDs[i] = k
		i++
	}
	return IDs

}

func Debug(ctx context.Context, dns DNSRegistry, loop int) {
	for {

		svcIDs := dns.listSvcIDs()
		fmt.Println(svcIDs)
		// for _, id := range svcIDs {
		// 	fmt.Println(id)
		// }

		containers := listJoinedCtrs(ctx, dns.NetworkName)
		ctrIDs := extractCtrIDs(containers)
		// for _, ctrID := range ctrIDs {
		// 	fmt.Println(ctrID)
		// }

		remoteSet := NewSet()
		remoteSet.AddList(&svcIDs)
		fmt.Println(remoteSet.Size())

		localSet := NewSet()
		localSet.AddList(&ctrIDs)
		fmt.Println(localSet.Size())

		svc2Register := localSet.Difference(remoteSet)
		fmt.Println("No. of services to register: ", svc2Register.Size())
		for id := range svc2Register.content {
			ctrJSON := containers[id]
			log.Infof("Registering service: <%s>", ctrJSON.Name)
			dns.registerSvc(&ctrJSON)
		}
		svc2Deregister := remoteSet.Difference(localSet)
		fmt.Println("No. of services to deregister: ", svc2Deregister.Size())
		for id := range svc2Deregister.content {
			log.Infof("Deregistering service: <%s>", id)
			dns.deregisterSvc(id)
		}

		// for _, ctr := range containers {
		// }
		time.Sleep(5 * time.Second)
	}
}

func WatchCtrs(ctx context.Context, netName string, dns DNSRegistry, loop int) {
	// Main func for watching
	fmt.Println("ʕ•o•ʔ Containers' watch begins")

	filter := filters.NewArgs()
	// Watch Docker events with type: "container", "network"
	// filter.Add("type", "container")
	filter.Add("type", "network")
	// Only watch events below
	// filter.Add("event", "start")
	// filter.Add("event", "stop")
	// filter.Add("event", "restart")
	filter.Add("event", "connect")
	filter.Add("event", "disconnect")
	// filter.Add("event", "destroy")

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// Ignore error channel
	evtCh, _ := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})

	for evt := range evtCh {
		evtNetName := evt.Actor.Attributes["name"]
		if evtNetName == netName {
			fmt.Println("DETECT network connect/disconnect event")
			containers := listJoinedCtrs(ctx, netName)
			for _, ctr := range containers {
				fmt.Println(ctr.ContainerJSONBase.ID)
				fmt.Println(ctr.ContainerJSONBase.Name)
				fmt.Println(ctr.ContainerJSONBase.Image)
				fmt.Println("-----------------")
			}
		}

		// dns.listSvcs()
	}
}
