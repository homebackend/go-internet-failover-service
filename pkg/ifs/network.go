package ifs

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	selfcommon "github.com/homebackend/go-homebackend-common/pkg"
)

func NetworkStart(sudo bool, clean bool, connection *Connection) {
	ifName := fmt.Sprintf("%sa", connection.Name)
	peerIfName := fmt.Sprintf("%sb", connection.Name)
	ip := fmt.Sprintf("%s/%d", connection.IP, connection.Mask)
	peerIp := fmt.Sprintf("%s/%d", connection.PeerIp, connection.Mask)
	mark := fmt.Sprintf("%d", connection.Mark)

	_, out, _ := IpExecute(sudo, []string{"netns"})
	if strings.Contains(out, connection.Name) {
		log.Printf("Namespace %s already exists.", connection.Name)
		if clean {
			log.Printf("Performing namespace %s cleanup", connection.Name)
			IpExecute(sudo, []string{"netns", "del", connection.Name})
			_, out, _ := IpExecute(sudo, []string{"link"})
			if strings.Contains(out, ifName) {
				log.Printf("Performing link %s cleanup", ifName)
				IpExecute(sudo, []string{"link", "del", ifName})
			}
			time.Sleep(1 * time.Second)
		} else {
			os.Exit(1)
		}
	}
	IpExecute(sudo, []string{"netns", "add", connection.Name})

	IpExecute(sudo, []string{"link", "add", ifName, "type", "veth", "peer", "name", peerIfName})
	IpExecute(sudo, []string{"link", "set", peerIfName, "netns", connection.Name})

	IpExecute(sudo, []string{"addr", "add", ip, "dev", ifName})
	IpExecuteNs(sudo, true, connection.Name, []string{"addr", "add", peerIp, "dev", peerIfName})
	IpExecute(sudo, []string{"link", "set", ifName, "up"})
	IpExecuteNs(sudo, true, connection.Name, []string{"link", "set", peerIfName, "up"})

	IpExecuteNs(sudo, true, connection.Name, []string{"route", "add", "default", "via", connection.IP})

	IptAddExecute(sudo, []string{"FORWARD", "-o", connection.GwIf, "-i", ifName, "-j", "ACCEPT"})
	IptAddExecute(sudo, []string{"FORWARD", "-i", connection.GwIf, "-o", ifName, "-j", "ACCEPT"})
	IptAddExecuteTable(sudo, "nat", []string{"POSTROUTING", "-s", ip, "-o", connection.GwIf, "-j", "MASQUERADE"})
	IptAddExecuteTable(sudo, "mangle", []string{"PREROUTING", "-i", ifName, "-j", "MARK", "--set-mark", mark})

	_, out, _ = IpExecute(sudo, []string{"rule"})
	if strings.Contains(out, fmt.Sprintf("lookup %d", connection.Mark)) {
		log.Printf("Lookup rule for `%s` already exists.", connection.Name)
		if clean {
			log.Printf("Performing lookup rule for `%s` cleanup", connection.Name)
			IpExecute(sudo, []string{"rule", "del", "table", mark})
		} else {
			os.Exit(1)
		}
	}
	IpExecute(sudo, []string{"rule", "add", "fwmark", mark, "table", mark})

	code, _, _ := IpExecuteNs(sudo, false, "", []string{"route", "list", "table", mark})
	if code == 0 {
		log.Printf("Routing table for `%s/%s` already exists.", connection.Name, mark)
		if clean {
			log.Printf("Performing routing table for `%s/%s` cleanup", connection.Name, mark)
			IpExecute(sudo, []string{"route", "flush", "table", mark})
		} else {
			os.Exit(1)
		}
	}
	IpExecute(sudo, []string{"route", "add", "default", "via", connection.GwIp, "table", mark})

	dir := fmt.Sprintf("/etc/netns/%s", connection.Name)
	selfcommon.Execute(sudo, false, []string{"mkdir", "-p", dir})
	selfcommon.Execute(sudo, false, []string{"bash", "-c", fmt.Sprintf("echo nameserver 1.1.1.1 > \"%s/resolv.conf\"", dir)})
}

func NetworkStop(sudo bool, connection *Connection) {
	ifName := fmt.Sprintf("%sa", connection.Name)
	cidr := fmt.Sprintf("%s/%d", connection.IP, connection.Mask)
	ipo, _, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Printf("Error parsing IP: %s", cidr)
	}
	ip := fmt.Sprintf("%s/%d", ipo.String(), connection.Mask)
	mark := fmt.Sprintf("%d", connection.Mark)

	IptDeleteExecute(sudo, []string{"FORWARD", "-o", connection.GwIf, "-i", ifName, "-j", "ACCEPT"})
	IptDeleteExecute(sudo, []string{"FORWARD", "-i", connection.GwIf, "-o", ifName, "-j", "ACCEPT"})
	IptDeleteExecuteTable(sudo, "nat", []string{"POSTROUTING", "-s", ip, "-o", connection.GwIf, "-j", "MASQUERADE"})
	IptDeleteExecuteTable(sudo, "mangle", []string{"PREROUTING", "-i", ifName, "-j", "MARK", "--set-mark", mark})

	IpExecute(sudo, []string{"rule", "del", "table", mark})
	IpExecute(sudo, []string{"route", "flush", "table", mark})
	IpExecute(sudo, []string{"netns", "del", connection.Name})

	selfcommon.Execute(sudo, false, []string{"rm", "-rf", fmt.Sprintf("/etc/netns/%s", connection.Name)})
}

func SetDefautRoute(sudo bool, ping string, connection *Connection) bool {
	_, out, _ := IpExecute(sudo, []string{"route", "get", ping})
	if !strings.Contains(out, fmt.Sprintf("via %s dev %s", connection.GwIp, connection.GwIf)) {
		log.Printf("Setting default route to %s", connection.GwIp)
		IpExecute(sudo, []string{"route", "del", "default"})
		IpExecute(sudo, []string{"route", "add", "default", "via", connection.GwIp, "dev", connection.GwIf})
	}

	return true
}

func Ping(sudo bool, ns string, pingIp string) bool {
	code, _, _ := selfcommon.Execute(sudo, false, []string{"ip", "netns", "exec", ns, "ping", "-q", "-c", "1", pingIp})
	return code == 0
}

func Curl(sudo bool, ns string, url string) bool {
	code, _, _ := selfcommon.Execute(sudo, false, []string{"ip", "netns", "exec", ns, "curl", "-Is", url})
	return code == 0
}
