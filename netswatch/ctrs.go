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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func listJoinedCtrs(ctx context.Context, name string) []types.ContainerJSON {
	// List containers which joined Netswatch bridge network
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	nr, err := cli.NetworkInspect(ctx, name)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(nr.Containers))

	containers := make([]types.ContainerJSON, len(nr.Containers))

	i := 0
	for cID := range nr.Containers {
		ctr, err := cli.ContainerInspect(ctx, cID)
		if err != nil {
			panic(nil)
		}
		containers[i] = ctr
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

func WatchCtrs(ctx context.Context, netName string, loop int) {
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
			// listJoinedCtrs(ctx, netName)
		}

		containers := listJoinedCtrs(ctx, netName)
		for _, ctr := range containers {
			fmt.Printf("%+v\n", ctr)
		}
	}
}
