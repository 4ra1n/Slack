package core

import (
	"regexp"
	"strings"
)

var (
	FscanRegs = map[string][]string{
		"FTP":       {"[+] ftp"},
		"SSH":       {"[+] SSH"},
		"Mssql":     {"[+] mssql"},
		"Oracle":    {"[+] oracle"},
		"Mysql":     {"[+] mysql"},
		"RDP":       {"[+] RDP"},
		"Redis":     {"[+] Redis"},
		"Postgres":  {"[+] Postgres"},
		"Mongodb":   {"[+] Mongodb"},
		"Memcached": {"[+] Memcached"},
		"MS17-010":  {"[+] MS17-010"},
		"POC":       {"poc"},
		"DC INFO":   {"DC"},
		"INFO":      {"[+] InfoScan"},
		"Vcenter":   {"ID_VC_Welcome"},
		"Camera":    {"len:2512", "len:600", "len:481", "len:480"},
	}
	NetInfoReg = regexp.MustCompile(`\[\*]((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}(\s+\[\-\>](\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[a-zA-Z0-9\-]+))+`)
)

func MatchLine(name string, contains, lines []string) string {
	var temp []string
	var result string
	for _, v := range lines {
		for _, c := range contains {
			if strings.Contains(strings.ToLower(v), strings.ToLower(c)) {
				temp = append(temp, v)
			}
		}
	}
	if len(temp) > 0 {
		result += "[" + name + "]\n"
		for _, v := range temp {
			result += v + "\n"
		}
		result += "\n\n"
	}
	return result
}
