package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/akamensky/argparse"
	homecommon "github.com/homebackend/go-homebackend-common/pkg"
	"github.com/homebackend/go-internet-failover-service/pkg/ifs"
)

var requiredCommands = []string{
	"ip",
	"sudo",
	"iptables",
	"ping",
}

const (
	PROG_NAME = "goifs"
	CONF_FILE = "/etc/goifs/ifs.yaml"
	ADDR_FILE = "goifs.addr"
)

func service(c *string, t *bool, n *bool) {
	homecommon.CheckPrerequisites(homecommon.O_LINUX, *c, requiredCommands)
	pidFile := homecommon.CreatePidFile()

	config := homecommon.GetConf[ifs.Configuration](*c)
	sort.Slice(config.Connections, func(i, j int) bool {
		return config.Connections[i].Rank < config.Connections[j].Rank
	})

	if !*t {
		for _, c := range config.Connections {
			ifs.NetworkStart(config.UseSudo, config.CleanIfRequired, c)
		}
	}

	defer func() {
		homecommon.StopIpc(PROG_NAME)
		pidFile.Unlock()

		if *t || *n {
			log.Printf("Skipping cleanup on exit")
			return
		}
		log.Printf("Performing exit cleanup")
		for _, c := range config.Connections {
			ifs.NetworkStop(config.UseSudo, c)
		}
	}()

	p := ifs.NewProcessor(config)
	sigc := homecommon.Signal()

	p.Start()

	cs := new(ifs.Status)
	cs.CI = p.GetConnectionInfo()
	if err := homecommon.StartIpc(PROG_NAME, cs); err != nil {
		os.Exit(1)
	}

	log.Printf("Service started.")

	for {
		select {
		case s := <-sigc:
			log.Printf("Signal captured: %s", s)
			p.Stop()
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func status() {
	pid := homecommon.GetPid(PROG_NAME)
	if ci, err := ifs.IpcGetConnectionStatus(PROG_NAME); err != nil {
		log.Fatalf("Error getting connection status: %s", err)
	} else {
		names := make([]string, len(ci))
		i := 0
		for n, _ := range ci {
			names[i] = n
			i++
		}
		slices.Sort(names)
		log.Printf("goifs is running with process id: %d", pid)
		log.Println("Connection Details::")
		w := tabwriter.NewWriter(os.Stdout, 0, 1, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
		fmt.Fprintln(w, "Name\tGateway\tIs Up\tSuccesses\tFailures\tTotal\tConsecutive Successes\tConsecutive Failures\tActive")
		for _, n := range names {
			i := ci[n]
			fmt.Fprintf(w, "%s\t%s\t%t\t%d\t%d\t%d\t%d\t%d\t%t\n",
				n,
				i.Gateway,
				i.IsUp,
				i.Success,
				i.Failure,
				i.Success+i.Failure,
				i.ConsecutiveSuccesses,
				i.ConsecutiveFailures,
				i.Active,
			)
		}
		w.Flush()
	}
}

func main() {
	parser := argparse.NewParser(os.Args[0], "Sets up internet connection failover")

	startCommand := parser.NewCommand("start", "Start the internet failover service")
	stopCommand := parser.NewCommand("stop", "Stop the internet faillover service")
	statusCommand := parser.NewCommand("status", "Show status of the internet failover service")

	c := startCommand.String("c", "configuration-file", &argparse.Options{
		Required: false,
		Default:  CONF_FILE,
		Help:     "Configuration File",
	})

	t := startCommand.Flag("t", "try-with-existing-network", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "Try running with existing configuration. I.e. It assumes required network namespaces and iptables already exist.",
	})

	n := startCommand.Flag("n", "no-exit-cleanup", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     "Do not cleanup namespace and iptable configuration on exit.",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	if startCommand.Happened() {
		service(c, t, n)
	} else if stopCommand.Happened() {
		homecommon.Stop(PROG_NAME)
	} else if statusCommand.Happened() {
		status()
	}
}
