package ifs

import "time"

type Processor struct {
	processes map[string]*Process
	config    *Configuration
	notifier  chan string
	quit      chan struct{}
}

func NewProcessor(config *Configuration) *Processor {
	n := make(chan string)
	processes := make(map[string]*Process)
	for _, c := range config.Connections {
		processes[c.Name] = NewProcess(c.Name, config, c, &ConnectionInfo{}, n)
	}

	return &Processor{
		processes: processes,
		config:    config,
		notifier:  n,
		quit:      make(chan struct{}),
	}
}

func (proc *Processor) _SetRoute() {
	routeSet := false
	for _, c := range proc.config.Connections {
		if !routeSet && proc.processes[c.Name].IsUp() {
			SetDefautRoute(proc.config.UseSudo, proc.config.Ping, c)
			proc.processes[c.Name].info.Active = true
			routeSet = true
		} else {
			proc.processes[c.Name].info.Active = false
		}
	}
}

func (proc *Processor) Start() {
	go func() {
		for _, p := range proc.processes {
			p.Start()
		}

		ticker := time.NewTicker(60000 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-proc.quit:
				proc.quit = nil
				return
			case <-proc.notifier:
				proc._SetRoute()
			case <-ticker.C:
				proc._SetRoute()
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (proc *Processor) Stop() error {
	for _, p := range proc.processes {
		p.Stop()
	}

	close(proc.quit)
	return nil
}

func (proc *Processor) GetConnectionInfo() map[string]*ConnectionInfo {
	ci := make(map[string]*ConnectionInfo)
	for _, c := range proc.config.Connections {
		p := proc.processes[c.Name]
		ci[c.Name] = p.GetConnectionInfo()
	}

	return ci
}
