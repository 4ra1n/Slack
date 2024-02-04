package core

import (
	"context"
	"net"
	"os"
	"slack-wails/lib/util"
	"time"

	"github.com/miekg/dns"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"gopkg.in/yaml.v2"
)

func Resolution(domain string, servers []string, timeout int) (ips, cname []string, err error) {
	cname, err = LookupCNAME(domain, servers, timeout)
	ips, _ = LookupHost(domain, servers, timeout)
	return util.RemoveDuplicates[string](ips), cname, err
}

func LookupHost(domain string, domainServers []string, timeout int) (ips []string, err error) {
	for _, domainServer := range domainServers {
		ips, err = LookupHostWithServers(domain, domainServer, timeout)
		if err == nil {
			return ips, nil
		}
	}
	return nil, err
}

func LookupHostWithServers(domain, domainServers string, timeout int) ([]string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Duration(timeout) * time.Second,
			}
			return d.DialContext(ctx, "tcp", domainServers)
		},
	}
	ips, err := r.LookupHost(context.Background(), domain)
	if err == nil {
		return ips, nil
	}
	return []string{}, err
}

func LookupCNAME(domain string, domainServers []string, timeout int) (cnames []string, err error) {
	for _, domainServer := range domainServers {
		cnames, err = LookupCNAMEWithServer(domain, domainServer, timeout)
		if err == nil {
			return cnames, nil
		}
	}
	return nil, err
}

func LookupCNAMEWithServer(domain, domainServer string, timeout int) ([]string, error) {
	c := dns.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	var CNAMES []string
	m := dns.Msg{}
	// 最终都会指向一个ip 也就是typeA, 这样就可以返回所有层的cname.
	m.SetQuestion(domain+".", dns.TypeA)
	r, _, err := c.Exchange(&m, domainServer)
	if err != nil {
		return nil, err
	}
	for _, ans := range r.Answer {
		record, isType := ans.(*dns.CNAME)
		if isType {
			CNAMES = append(CNAMES, record.Target)
		}
	}
	return CNAMES, nil
}

func ReadCDNFile(cdnFile string) map[string][]string {
	yamlData, err := os.ReadFile(util.HomeDir() + cdnFile)
	if err != nil {
		logger.NewDefaultLogger().Debug(err.Error())
	}
	data := make(map[string][]string)
	err = yaml.Unmarshal(yamlData, &data)
	if err != nil {
		logger.NewDefaultLogger().Debug(err.Error())
	}
	return data
}
