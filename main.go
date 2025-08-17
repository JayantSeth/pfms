package main

import (
	"fmt"
	"log"
	"os"

	"github.com/JayantSeth/pfms/utils"
	"gopkg.in/yaml.v3"
)

type Data struct {
	Nodes []utils.Node `yaml:"nodes"`
}


func main() {
	yamlFile, err := os.ReadFile("data.yaml")
	if err != nil {
		log.Fatalf("Failed to open file: %s\n", err.Error())
	}
	expandedContent := os.ExpandEnv(string(yamlFile))
	var data Data
	err = yaml.Unmarshal([]byte(expandedContent), &data)
	if err != nil {
		log.Fatalf("Failed to unmarshal: %s", err.Error())
	}

	fmt.Printf("Data: %v\n", data)
	fmt.Printf("Nodes Length: %d\n", len(data.Nodes))
}


func main2() {
	l1 := utils.Linux{Node: utils.Node{IpAddress: "localhost", Port: "22", Username: "", Password: ""}}
	a1 := utils.Arista{Node: utils.Node{IpAddress: "172.18.0.101", Port: "22", Username: "", Password: ""}}
	c1 := utils.Cisco{Node: utils.Node{IpAddress: "172.18.0.102", Port: "22", Username: "", Password: ""}}
	destIps := []string{"8.8.8.8", "4.4.2.2", "172.18.0.1", "172.18.0.102"}
	a1.DoPing(destIps)
	destIps2 := []string{"8.8.8.8", "4.4.2.2", "172.18.0.101", "172.18.0.102"}
	l1.DoPing(destIps2)
	destIps3 := []string{"172.18.0.1", "8.8.8.8", "4.4.2.2", "172.18.0.101"}
	c1.DoPing(destIps3)
	fmt.Printf("L1 Result: %v\n", l1.Result)
	fmt.Printf("A1 Result: %v\n", a1.Result)
	fmt.Printf("C1 Result: %v\n", c1.Result)
}