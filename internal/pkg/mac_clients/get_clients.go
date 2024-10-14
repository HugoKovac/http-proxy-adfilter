package macclients

import (
	"errors"
	"log"
	"net"
	"os"
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

func run_arp_openwrt() (string, error) {
	cmd := exec.Command("cat", "/proc/net/arp")
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

func parse_arp_output_openwrt(stdout string) (clients []Client) {
	lines := strings.Split(stdout, "\n")

	for _, line := range lines {
		splited := strings.Fields(line)
		if len(splited) < 4 {
			continue
		}

		ip := net.ParseIP(splited[0])
		mac, err := net.ParseMAC(strings.ToUpper(splited[3]))
		if err != nil {
			// log.Println(splited, err)
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
	run_fn, parse_fn := func () (func() (string, error), func (string) ([]Client)){
		if _, err := os.Open("/etc/openwrt_version"); err == nil {
			log.Println("openwrt")
			return run_arp_openwrt, parse_arp_output_openwrt
		} else {
			log.Println("normal")
			return run_arp, parse_arp_output
		}
	}()
	
	stdout, err := run_fn()

	if err != nil {
		log.Println(err)
		return nil
	}

	clients = parse_fn(stdout)

	return clients
}

func GetInfoFromIP(remoteAddr string) (client Client, err error) {
	splited := strings.Split(remoteAddr, ":")
	if len(splited) < 2 {
		return client, errors.New("error parsing ip: " + remoteAddr)
	}
	ip := splited[0]
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return client, errors.New("ip parsing failed")
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
