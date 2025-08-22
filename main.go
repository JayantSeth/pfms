package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/JayantSeth/pfms/report"
	"github.com/JayantSeth/pfms/utils"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
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
	srcIpToNodeMap := map[string]utils.Node{}
	for _, n := range data.Nodes {
		srcIpToNodeMap[n.IpAddress] = n
		switch n.Type {
		case "linux":
			linux := utils.Linux{Node: n}
			ops = append(ops, linux)
		case "arista_eos":
			arista := utils.Arista{Node: n}
			ops = append(ops, arista)
		case "cisco_ios":
			cisco := utils.Cisco{Node: n}
			ops = append(ops, cisco)
		default:
			linux := utils.Linux{Node: n}
			ops = append(ops, linux)
		}
	}
	result := utils.DoMultiplePing(ops)

	basic_html := report.GenBasicStructure()
	tables := ""
	for SourceIp,Result := range result {
		node := srcIpToNodeMap[SourceIp]
		table := report.GenTable(node)
		rows := ""
		for DstIp, Reachable := range Result {
			row := report.GenRow(DstIp, Reachable)
			if (Reachable) {
				fmt.Printf("\tDestination IP: %s is %sreachable%s\n",DstIp, Green, Reset)
			} else {
				fmt.Printf("\tDestination IP: %s is %snot reachable%s\n",DstIp, Red, Reset)
			}
			rows = fmt.Sprintf(`%s%s`, rows, row)
		}
		table = strings.Replace(table, "ROWS_PLACEHOLDER", rows, 1)
		tables = fmt.Sprintf("%s%s", tables, table)
		fmt.Printf("=================================================\n")
	}
	complete_html := strings.Replace(basic_html, "PLACEHOLDER", tables, 1)
	file, err := os.Create("index.html")
	if err != nil {
		log.Fatalf("Unable to open file: %s\n", err.Error())
	}
	file.WriteString(complete_html)
	fmt.Printf("Time taken: %v\n", time.Since(start_time))
}
