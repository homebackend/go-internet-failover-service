package ifs

import (
	"time"
)

const (
	SLOW_PING = 500 * time.Millisecond
	FAST_PING = 100 * time.Millisecond
)

type Process struct {
	name       string
	config     *Configuration
	connection *Connection
	info       *ConnectionInfo
	notifier   chan string
	quit       chan struct{}
}

func NewProcess(name string, config *Configuration, connection *Connection, info *ConnectionInfo, n chan string) *Process {
	info.Gateway = connection.GwIp

	return &Process{
		name:       name,
		config:     config,
		connection: connection,
		info:       info,
		notifier:   n,
		quit:       make(chan struct{}),
	}
}

func (p *Process) Start() {
	go func() {
		for {
			select {
			case <-p.quit:
				p.quit = nil
				return
			default:
				result := Ping(p.config.UseSudo, p.name, p.config.Ping)
				sleep := SLOW_PING
				if result {
					p.info.UpdateSuccess()
				} else {
					p.info.UpdateFailure()
				}

				// If ping failed or we don't have required count of results
				// ping again quickly, else take time.
				if !result || p.info.NeedMoreInfo() {
					sleep = FAST_PING
				}

				change := p.info.Evaluate(p.name, p.config.MaxPacketLoss, p.config.MinPacketLoss, p.config.MaxSuccessivePacketsLost,
					p.config.MinSuccessivePacketsRecvd)

				//log.Printf("%s: %v", p.name, p.info)

				if change {
					p.notifier <- p.name
				}

				time.Sleep(sleep)
			}
		}
	}()
}

func (p *Process) Stop() error {
	close(p.quit)
	return nil
}

func (p *Process) IsUp() bool {
	return p.info.IsUp
}

func (p *Process) GetConnectionInfo() *ConnectionInfo {
	return p.info
}
