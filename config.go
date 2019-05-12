package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
)

type Address struct {
	Host string
	Port int
}

type Config struct {
	BindAddresses []Address
	Destination   Address
}

var config *Config

func GetConfiguration() *Config {
	if config != nil {
		return config
	}

	bindHostAddr := flag.String("bind", "0.0.0.0", "Bind hosts. Example: 192.168.0.10")
	bindPortAddr := flag.Int("port", 0, "Bind port. Example: 8080")
	destAddr := flag.String("dest", "", "Destination address. Example: 10.0.0.2:8080")

	flag.Parse()

	// parse bind hosts
	hosts := strings.Split(*bindHostAddr, ",")
	bindHosts := make([]string, 0)
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		bindHosts = append(bindHosts, host)
	}

	if len(bindHosts) == 0 {
		log.Panicln("no hosts to bind")
	}

	// parse bind port
	if *bindPortAddr <= 0 || *bindPortAddr > 65535 {
		log.Panicf("invalid port number %d", *bindPortAddr)
	}

	bindAddr := make([]Address, len(hosts))
	for idx, host := range hosts {
		bindAddr[idx] = Address{Host: host, Port: *bindPortAddr}
	}

	dPortIdx := strings.LastIndex(*destAddr, ":")
	if dPortIdx == -1 {
		log.Fatalln("Missing destination port")
	}

	dHost := (*destAddr)[0:dPortIdx]
	dPort, err := strconv.Atoi((*destAddr)[dPortIdx:])
	if len(dHost) == 0 {
		log.Fatalln("Invalid destination address")
	}
	if err != nil {
		log.Fatalln("Invalid destination port")
	}

	config = &Config{
		BindAddresses: bindAddr,
		Destination: Address{
			Host: dHost,
			Port: dPort,
		},
	}

	return config
}
