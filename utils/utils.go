package utils

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Node struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	IpAddress string `yaml:"ip"`
	Username string	`yaml:"username"`
	Password string	`yaml:"password"`
	Port string	`yaml:"ssh_port"`
	DestIps []string `yaml:"dest_ips"`
	Result map[string]bool
}

func (n Node) ExecuteCommands(commands []string) string {
	Ciphers := ssh.InsecureAlgorithms().Ciphers
	Ciphers = append(Ciphers, ssh.SupportedAlgorithms().Ciphers...)
	KeyExchanges := ssh.InsecureAlgorithms().KeyExchanges
	KeyExchanges = append(KeyExchanges, ssh.SupportedAlgorithms().KeyExchanges...)
	Macs := ssh.InsecureAlgorithms().MACs
	Macs = append(Macs, ssh.SupportedAlgorithms().MACs...)
	config := &ssh.ClientConfig{
		User: n.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(n.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echon []bool) ([]string, error) {
				// The server is prompting for a password
				if len(questions) == 1 && strings.Contains(strings.TrimSpace(strings.ToLower(questions[0])), "password:") {
					return []string{n.Password}, nil
				}
				return nil, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Config: ssh.Config{
			Ciphers: Ciphers,
			KeyExchanges: KeyExchanges,
			MACs: Macs,
		},
	}

	client, err := ssh.Dial("tcp", n.IpAddress + ":" + n.Port, config)
	if err != nil {
		msg := fmt.Sprintf("Failed to connect to host: %v on port 22, error: %v, Username: %v, Password: %v", n.IpAddress, err, n.Username, n.Password)
		return msg
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		msg := fmt.Sprintf("Failed to create a session with client: %v", err.Error())
		return msg
	}
	defer session.Close()
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatalf("Unable to setup stdin for session: %v", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Unable to setup stdout for session: %v", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		log.Fatalf("Unable to setup stderr for session: %v", err)
	}

	output := ""

	// Start the remote shell
	if err := session.Shell(); err != nil {
		log.Fatalf("Failed to start shell: %v", err)
	}

	// Goroutine to read stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			output_now := scanner.Text()
			output += output_now
		}
	}()

	// Goroutine to read stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			output += scanner.Text()
		}
	}()

	// Send commands
	writer := bufio.NewWriter(stdin)

	for _, cmd := range commands {
		_, err := writer.WriteString(cmd + "\n")
		if err != nil {
			log.Printf("Error writing command: %v", err)
			break
		}
		writer.Flush()
		time.Sleep(6 * time.Second)
	}

	// Close stdin to signal end of input
	stdin.Close()

	// Wait for the session to finish (optional, depending on your needs)
	session.Wait()	
	return output
}


type Operations interface {
	DoPing() string
}


type Linux struct {
	Node
}

func (n *Node) ExtractPingResult(dstIps []string, output string) {
	data := strings.Split(output, "PING")
	ipRegex := regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])`)
	pingResultRegex := regexp.MustCompile(`(\d{1,3})%\spacket\sloss,`)
	resultMap := make(map[string]bool)
	for _, ip := range dstIps {
		resultMap[ip] = false
	}
	for _, data := range data {
		ip := ipRegex.FindString(data)
		pingResult := pingResultRegex.FindStringSubmatch(data)
		if len(pingResult) == 0 {
			continue
		}
		// fmt.Printf("|%s|\n%s", pingResult, ip)
		if pingResult[1] == "0" {
			resultMap[ip] = true
		} else {
			resultMap[ip] = false
		}
	}
	n.Result = resultMap
}


func (l *Linux) DoPing(dstIps []string) {
	commands := []string{}
	for _, ip := range dstIps {
		commands = append(commands, fmt.Sprintf("ping %s -c 3", ip))
	}
	output := l.ExecuteCommands(commands)
	l.ExtractPingResult(dstIps, output)

}

type Arista struct {
	Node
}


func (a *Arista) DoPing(dstIps []string) {
	commands := []string{}
	for _, ip := range dstIps {
		commands = append(commands, fmt.Sprintf("ping %s", ip))
	}
	output := a.ExecuteCommands(commands)
	a.ExtractPingResult(dstIps, output)
}

type Cisco struct {
	Node
}

func (c *Cisco) DoPing(dstIps []string) {
	commands := []string{}
	for _, ip := range dstIps {
		commands = append(commands, fmt.Sprintf("ping %s", ip))
	}
	output := c.ExecuteCommands(commands)
	c.ExtractPingResult(dstIps, output)
}


func (c *Cisco) ExtractPingResult(dstIps []string, output string) {
	data := strings.Split(output, "Echos")
	ipRegex := regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9]),`)
	pingResultRegex := regexp.MustCompile(`(\d{1,3})\spercent\s\(`)
	resultMap := make(map[string]bool)
	for _, ip := range dstIps {
		resultMap[ip] = false
	}
	for _, data := range data {
		ip := ipRegex.FindString(data)
		if len(ip) == 0 {
			continue
		}
		ip = ip[:len(ip)-1]
		pingResult := pingResultRegex.FindStringSubmatch(data)
		if len(pingResult) == 0 {
			continue
		}
		// fmt.Printf("|%s|\n%s", pingResult, ip)
		if pingResult[1] == "0" {
			resultMap[ip] = false
		} else {
			resultMap[ip] = true
		}
	}
	c.Result = resultMap
}
