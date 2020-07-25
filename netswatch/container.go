package netswatch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/filters"

	"github.com/docker/docker/api/types"
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
