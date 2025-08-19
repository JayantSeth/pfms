package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JayantSeth/pfms/utils"
	"gopkg.in/yaml.v3"
	"github.com/joho/godotenv"
)

type Data struct {
	Nodes []utils.Node `yaml:"nodes"`
}

const (
	Red = "\033[31m"
	Green = "\033[32m"
	Reset = "\033[0m"
)


func main() {
	start_time := time.Now()
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Failed to load .env file, make sure all environment variables are set\n")
	}
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

	var ops []utils.Operations
	for _, n := range data.Nodes {
		switch n.Type {
		case "linux":
			linux := &utils.Linux{Node: &n}
			ops = append(ops, linux)
		case "arista_eos":
			arista := &utils.Arista{Node: &n}
			ops = append(ops, arista)
		case "cisco_ios":
			cisco := &utils.Cisco{Node: &n}
			ops = append(ops, cisco)
		default:
			linux := &utils.Linux{Node: &n}
			ops = append(ops, linux)
		}
	}
	result := utils.DoMultiplePing(ops)
	
	for SourceIp,Result := range result {
		fmt.Printf("From Source IP: %s\n", SourceIp)
		for DstIp, Reachable := range Result {
			if (Reachable) {
				fmt.Printf("\tDestination IP: %s is %sreachable%s\n",DstIp, Green, Reset)
			} else {
				fmt.Printf("\tDestination IP: %s is %snot reachable%s\n",DstIp, Red, Reset)
			}
		}
		fmt.Printf("=================================================\n")
	}
	fmt.Printf("Time taken: %v\n", time.Since(start_time))
}

// OOPs, concurrency & parallelism
