package main

import (
	"fmt"

	"github.com/JayantSeth/pfms/utils"
)


func main() {
	l1 := utils.Linux{Node: utils.Node{IpAddress: "localhost", Port: "22", Username: "", Password: ""}}
	a1 := utils.Arista{Node: utils.Node{IpAddress: "172.18.0.101", Port: "22", Username: "", Password: ""}}
	destIps := []string{"8.8.8.8", "4.4.2.2", "172.18.0.1"}
	a1.DoPing(destIps)
	fmt.Printf("A1 Result: %v\n", a1.Result)
	destIps2 := []string{"8.8.8.8", "4.4.2.2", "172.18.0.101"}
	l1.DoPing(destIps2)
	fmt.Printf("L1 Result: %v\n", l1.Result)
}