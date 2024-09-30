package macclients

import (
	"errors"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

type Client struct{
	IP	net.IP
	MAC net.HardwareAddr
}

var (
	Clients = getMACTable()
)

func FixMacOSMACNotation(s string) string {
    var e int
    var sb strings.Builder
    for i := 0; i < len(s); i++ {
        r := s[i]
        if r == ':' {
            for j := e; j < 2; j++ {
                sb.WriteString("0")
            }
            sb.WriteString(s[i-e : i])
            sb.WriteString(":")
            e = 0
            continue
        }
        e++
    }
    for j := e; j < 2; j++ {
        sb.WriteString("0")
    }
    sb.WriteString(s[len(s)-e:])
    return sb.String()
}

func run_arp() (string, error) {
	cmd := exec.Command("arp", "-na")
	stdout, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return string(stdout), err
}

func parse_arp_output(stdout string) (clients []Client) {
	lines := strings.Split(stdout, "\n")

	for _, line := range lines {
		splited := strings.Split(line, " ")
		if len(splited) < 4 {
			continue
		}

		ip := net.ParseIP(strings.Trim(splited[1], "()"))
		if strings.Contains(runtime.GOOS, "darwin") {
			splited[3] = FixMacOSMACNotation(splited[3])
		}
		mac, err := net.ParseMAC(strings.ToUpper(splited[3]))
		if err != nil {
			// log.Println(splited[1], err)
		} else {
			client := Client{
				IP: ip,
				MAC: mac,
			}
			clients = append(clients, client)			
		}

	}

	return clients
}

func getMACTable() (clients []Client) {
	stdout, err := run_arp()

	if err != nil {
		log.Println(err)
		return nil
	}

	clients = parse_arp_output(stdout)

	return clients
}

func GetInfoFromIP(remoteAddr string) (client Client, err error) {
	splited := strings.Split(remoteAddr, ":")
	if len(splited) < 2 {
		return client, errors.New("Error parsing IP")
	}
	ip := splited[0]
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return client, errors.New("IP parsing failed")
	}

	for _, client = range Clients {
		if strings.Compare(client.IP.String(), ip) == 0 {
			return client, nil
		}
	}

	Clients = getMACTable()

	for _, client = range Clients {
		if strings.Compare(client.IP.String(), ip) == 0 {
			return client, nil
		}
	}

	return client, errors.New("GetInfoFromIP: not found for: " + ip)
}
