package ifs

import selfcommon "github.com/homebackend/go-homebackend-common/pkg"

func IpExecuteNs(sudo bool, check bool, ns string, command []string) (int, string, string) {
	if ns != "" {
		command = append([]string{"netns", "exec", ns, "ip"}, command...)
	}

	command = append([]string{"ip"}, command...)

	return selfcommon.Execute(sudo, check, command)
}

func IpExecute(sudo bool, command []string) (int, string, string) {
	return IpExecuteNs(sudo, true, "", command)
}

func IptExecute(sudo bool, table string, delete bool, command []string) (int, string, string) {
	var checkCommand []string
	if table == "" {
		checkCommand = append([]string{"iptables", "-C"}, command...)
	} else {
		checkCommand = append([]string{"iptables", "-t", table, "-C"}, command...)
	}

	code, _, _ := selfcommon.Execute(sudo, false, checkCommand)
	if (!delete && code != 0) || (delete && code == 0) {
		flag := "-A"
		if delete {
			flag = "-D"
		}

		if table == "" {
			command = append([]string{"iptables", flag}, command...)
		} else {
			command = append([]string{"iptables", "-t", table, flag}, command...)
		}

		return selfcommon.Execute(sudo, true, command)
	}

	return code, "", ""
}

func IptAddExecuteTable(sudo bool, table string, command []string) (int, string, string) {
	return IptExecute(sudo, table, false, command)
}

func IptDeleteExecuteTable(sudo bool, table string, command []string) (int, string, string) {
	return IptExecute(sudo, table, true, command)
}

func IptAddExecute(sudo bool, command []string) (int, string, string) {
	return IptAddExecuteTable(sudo, "", command)
}

func IptDeleteExecute(sudo bool, command []string) (int, string, string) {
	return IptDeleteExecuteTable(sudo, "", command)
}
