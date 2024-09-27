package macclients

import (
	"log"
	"os/exec"
	"strings"
	"net"
	"runtime"
)

type client struct{
	ip	net.IP
	mac net.HardwareAddr
}

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

func parse_arp_output(stdout string) (clients []client) {
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
			log.Println(err)
		} else {
			client := client{
				ip: ip,
				mac: mac,
			}
			clients = append(clients, client)			
		}

	}

	return clients
}

func GetMACTable() {
	stdout, err := run_arp()

	if err != nil {
		log.Println(err)
		return
	}

	clients := parse_arp_output(stdout)
	for _, client := range clients {
		log.Println(client)
	}
}
