package utils

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func SetUp() Node {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Failed to load .env file, make sure all environment variables are set\n")
	}
	node := Node{Name: "localhost", Type: "linux", IpAddress: "127.0.0.1", Username: os.Getenv("LOCAL_USER"), Password: os.Getenv("LOCAL_PASS"), Port: "22", DestIps: []string{"8.8.8.8", "4.4.2.2"}}
	return node
}

func TestGetSourceIp(t *testing.T) {
	node := SetUp()
	sourceIp := node.GetSourceIp()
	if sourceIp != "127.0.0.1" {
		t.Errorf("Get source IP did not return correct Source IP")
	}
}

func TestLinuxPing(t *testing.T) {
	node := SetUp()
	l := Linux{Node: node}
	result, err := l.DoPing()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if result["8.8.8.8"] != true {
		t.Errorf("8.8.8.8 should be reachable, but it's not")
	}
}
