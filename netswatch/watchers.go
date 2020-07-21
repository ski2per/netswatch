package netswatch

import (
	"context"
	"fmt"
	"time"
)

func Hello() {
	fmt.Println("一哭二闹三上悠亚")
}

func WatchNets(ctx context.Context) {
	for {
		fmt.Println("watching nets")
		time.Sleep(2 * time.Second)

		select {
		case <-ctx.Done():
			fmt.Println("done netswatch")
			return
		default:
			// case <-time.After(2 * time.Second):
			// 	fmt.Println("1024")
		}

	}
}
