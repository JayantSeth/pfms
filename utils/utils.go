package utils

import (
	"bufio"
	"errors"
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
}

func (n Node) ExecuteCommands(commands []string, readWait int) (string, error) {
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
		msg := fmt.Sprintf("\nFailed to connect to host: %s on port 22, error: %s\n", n.IpAddress, err.Error())
		return "", errors.New(msg)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		msg := fmt.Sprintf("Failed to create a session with client: %v", err.Error())
		return "", errors.New(msg)
	}
	defer session.Close()
	stdin, err := session.StdinPipe()
	if err != nil {
		msg := fmt.Sprintf("Unable to setup stdin for session: %v", err.Error())
		return "", errors.New(msg)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		msg := fmt.Sprintf("Unable to setup stdout for session: %v", err.Error())
		return "", errors.New(msg)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		msg := fmt.Sprintf("Unable to setup stderr for session: %v", err.Error())
		return "", errors.New(msg)
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
		time.Sleep(time.Duration(readWait) * time.Second)
	}

	// Close stdin to signal end of input
	stdin.Close()

	// Wait for the session to finish (optional, depending on your needs)
	session.Wait()
	return output, nil
}

func (n Node) GetSourceIp() string {
	return n.IpAddress
}

type Operations interface {
	DoPing() (map[string]bool, error)
	GetSourceIp() string
}

type Linux struct {
	Node Node
}

func (l Linux) GetSourceIp() string {
	return l.Node.IpAddress
}

func DoPingCh (o Operations, ch chan Result, wg *sync.WaitGroup) {
	defer wg.Done()
	result, err := o.DoPing()
	if err != nil {
		ch <- Result{IpAddress: fmt.Sprintf("%s %s", o.GetSourceIp(), err.Error()), Result: result}
	} else {
		ch <- Result{IpAddress: o.GetSourceIp(), Result: result}	
	}
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
		go DoPingCh(o, ch, &wg)
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


func (n Node) ExtractPingResult(output string) map[string]bool {
	data := strings.Split(output, "PING")
	ipRegex := regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])`)
	pingResultRegex := regexp.MustCompile(`(\d{1,3})%\spacket\sloss,`)
	resultMap := make(map[string]bool)
	for _, ip := range n.DestIps {
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

func (l Linux) DoPing() (map[string]bool, error) {
	commands := []string{}
	for _, ip := range l.Node.DestIps {
		commands = append(commands, fmt.Sprintf("ping %s -c 3", ip))
	}
	output, err := l.Node.ExecuteCommands(commands, 1)	
	if err != nil {
		map1 := make(map[string]bool)
		return map1, err
	}
	return l.Node.ExtractPingResult(output), nil

}

type Arista struct {
	Node Node
}

func (a Arista) GetSourceIp() string {
	return a.Node.IpAddress
}

func (a Arista) DoPing() (map[string]bool, error) {
	commands := []string{"en"}
	for _, ip := range a.Node.DestIps {
		commands = append(commands, fmt.Sprintf("ping %s repeat 3", ip))
	}
	output, err := a.Node.ExecuteCommands(commands, 1)
	if err != nil {
		map1 := make(map[string]bool)
		return map1, err
	}
	return a.Node.ExtractPingResult(output), nil
}

type Cisco struct {
	Node Node
}

func (c Cisco) GetSourceIp() string {
	return c.Node.IpAddress
}

func (c Cisco) DoPing() (map[string]bool, error) {
	commands := []string{}
	for _, ip := range c.Node.DestIps {
		commands = append(commands, fmt.Sprintf("ping %s timeout 1 r 3", ip))
	}
	output, err := c.Node.ExecuteCommands(commands, 3)
	if err != nil {
		map1 := make(map[string]bool)
		return map1, err
	}
	return c.ExtractPingResult(output), nil
}

func (c Cisco) ExtractPingResult(output string) map[string]bool {
	data := strings.Split(output, "Echos")
	ipRegex := regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9]),`)
	pingResultRegex := regexp.MustCompile(`(\d{1,3})\spercent\s\(`)
	resultMap := make(map[string]bool)
	for _, ip := range c.Node.DestIps {
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
