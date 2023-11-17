package ifs

import "log"

const (
	MAX_COUNT = 100
)

type ConnectionInfo struct {
	Gateway              string
	Active               bool
	Success              uint
	Failure              uint
	IsUp                 bool
	ConsecutiveSuccesses uint
	ConsecutiveFailures  uint
	LastSuccess          bool
}

func (ci *ConnectionInfo) UpdateSuccess() {
	if ci.Success+ci.Failure < MAX_COUNT {
		ci.Success++
	} else {
		if ci.Failure > 0 {
			ci.Failure--
			ci.Success++
		}
	}

	if ci.LastSuccess {
		if ci.ConsecutiveSuccesses < MAX_COUNT {
			ci.ConsecutiveSuccesses++
		}
	} else {
		ci.LastSuccess = true
		ci.ConsecutiveSuccesses = 1
		ci.ConsecutiveFailures = 0
	}
}

func (ci *ConnectionInfo) UpdateFailure() {
	if ci.Success+ci.Failure < MAX_COUNT {
		ci.Failure++
	} else {
		if ci.Success > 0 {
			ci.Failure++
			ci.Success--
		}
	}

	if !ci.LastSuccess {
		if ci.ConsecutiveFailures < MAX_COUNT {
			ci.ConsecutiveFailures++
		}
	} else {
		ci.LastSuccess = false
		ci.ConsecutiveSuccesses = 0
		ci.ConsecutiveFailures = 1
	}
}

func (ci *ConnectionInfo) NeedMoreInfo() bool {
	return ci.Success+ci.Failure < MAX_COUNT
}

func (ci *ConnectionInfo) Evaluate(ns string,
	maxPacketLoss uint, minPacketLoss uint,
	maxSuccessivePktsLost uint, minSuccessivePktsRcved uint) bool {

	if ci.IsUp && (ci.Failure >= maxPacketLoss || (!ci.LastSuccess && ci.ConsecutiveFailures >= maxSuccessivePktsLost)) {
		log.Printf("Network `%s` is now down", ns)
		ci.IsUp = false
		return true
	}

	if !ci.IsUp && (ci.Failure <= minPacketLoss && (ci.LastSuccess && ci.ConsecutiveSuccesses >= minSuccessivePktsRcved)) {
		log.Printf("Network `%s` is now up", ns)
		ci.IsUp = true
		return true
	}

	return false
}
