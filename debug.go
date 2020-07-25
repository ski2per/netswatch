package main

import (
	"fmt"

	nw "github.com/coreos/flannel/netswatch"
)

func main() {
	fmt.Println("hello")
	fmt.Println(nw.GenerateNodeMeta())
}
