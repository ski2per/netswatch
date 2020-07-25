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
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func listCtrInNetwork(ctx context.Context) {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	nr, err := cli.NetworkInspect(ctx, "sob")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", nr)
}

func ListContainers(ctx context.Context) {
	listCtrInNetwork(ctx)

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

func WatchCtrEvents(ctx context.Context, wg *sync.WaitGroup) {
	filter := filters.NewArgs()
	// Watch Docker events with type: "container", "network"
	filter.Add("type", "container")
	filter.Add("type", "network")
	// Only watch events below
	filter.Add("event", "start")
	filter.Add("event", "stop")
	filter.Add("event", "restart")
	filter.Add("event", "connect")
	filter.Add("event", "disconnect")
	filter.Add("event", "destroy")

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// Ignore error channel
	evtCh, _ := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})
	for evt := range evtCh {
		fmt.Println(evt.Type)
		fmt.Printf("%+v", evt)
	}
}

func WatchNetwork(ctx context.Context, wg *sync.WaitGroup) {
	for {
		fmt.Println("watching network")
		time.Sleep(time.Second * 2)
	}
}
