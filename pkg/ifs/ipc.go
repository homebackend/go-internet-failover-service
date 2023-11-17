package ifs

import (
	homecommon "github.com/homebackend/go-homebackend-common/pkg"
)

type Status struct {
	CI map[string]*ConnectionInfo
}

func (s *Status) GetStatus(args *homecommon.Nothing, ci *map[string]ConnectionInfo) error {
	c := make(map[string]ConnectionInfo, len(s.CI))
	for n, i := range s.CI {
		c[n] = *i
	}
	*ci = c
	return nil
}

func IpcGetConnectionStatus(progName string) (map[string]ConnectionInfo, error) {
	return homecommon.IpcGetData[map[string]ConnectionInfo](progName, "Status.GetStatus", 0)
}
