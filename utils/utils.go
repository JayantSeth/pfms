package utils

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Node struct {
	Name      string   `yaml:"name"`
	Type      string   `yaml:"type"`
	IpAddress string   `yaml:"ip"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Port      string   `yaml:"ssh_port"`
	DestIps   []string `yaml:"dest_ips"`
	Result    map[string]bool
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
			Ciphers:      Ciphers,
			KeyExchanges: KeyExchanges,
			MACs:         Macs,
		},
	}

	client, err := ssh.Dial("tcp", n.IpAddress+":"+n.Port, config)
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

func (n Node) GetSourceIp() string {
	return n.IpAddress
}

func (n Node) GetDestIps() []string {
	return n.DestIps
}

type Operations interface {
	DoPing(dstIps []string) map[string]bool
	GetSourceIp() string
	GetDestIps() []string
}

type Linux struct {
	Node *Node
}

func (l Linux) GetDestIps() []string {
	return l.Node.DestIps
}

func (l Linux) GetSourceIp() string {
	return l.Node.IpAddress
}

func DoPingCh (o Operations, dstIps []string, ch chan Result, wg *sync.WaitGroup) {
	defer wg.Done()
	result := o.DoPing(dstIps)
	ch <- Result{IpAddress: o.GetSourceIp(), Result: result}	
}

type Result struct {
	IpAddress string
	Result map[string]bool
}

func DoMultiplePing(ons []Operations) map[string]map[string]bool {
	ch := make(chan Result)
	var wg sync.WaitGroup
	for _, o := range ons {
		wg.Add(1)
		go DoPingCh(o, o.GetDestIps(), ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	result := make(map[string]map[string]bool)
	for r := range ch {
		result[r.IpAddress] = r.Result
	}
	return result
}


func (n *Node) ExtractPingResult(dstIps []string, output string) map[string]bool {
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
		if pingResult[1] == "0" {
			resultMap[ip] = true
		} else {
			resultMap[ip] = false
		}
	}
	return resultMap
}

func (l *Linux) DoPing(dstIps []string) map[string]bool {
	commands := []string{}
	for _, ip := range dstIps {
		commands = append(commands, fmt.Sprintf("ping %s -c 3", ip))
	}
	output := l.Node.ExecuteCommands(commands)
	return l.Node.ExtractPingResult(dstIps, output)

}

type Arista struct {
	Node *Node
}

func (a Arista) GetDestIps() []string {
	return a.Node.DestIps
}

func (a Arista) GetSourceIp() string {
	return a.Node.IpAddress
}

func (a *Arista) DoPing(dstIps []string) map[string]bool {
	commands := []string{}
	for _, ip := range dstIps {
		commands = append(commands, fmt.Sprintf("ping %s", ip))
	}
	output := a.Node.ExecuteCommands(commands)
	return a.Node.ExtractPingResult(dstIps, output)
}

type Cisco struct {
	Node *Node
}

func (c Cisco) GetDestIps() []string {
	return c.Node.DestIps
}

func (c Cisco) GetSourceIp() string {
	return c.Node.IpAddress
}

func (c *Cisco) DoPing(dstIps []string) map[string]bool {
	commands := []string{}
	for _, ip := range dstIps {
		commands = append(commands, fmt.Sprintf("ping %s", ip))
	}
	output := c.Node.ExecuteCommands(commands)
	return c.ExtractPingResult(dstIps, output)
}

func (c *Cisco) ExtractPingResult(dstIps []string, output string) map[string]bool {
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
		if pingResult[1] == "0" {
			resultMap[ip] = false
		} else {
			resultMap[ip] = true
		}
	}
	return resultMap
}
