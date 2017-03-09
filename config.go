package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
)

type config struct {
	addrListen       string
	allow            SpecList
	deny             SpecList
	interruptTimeout time.Duration
	gracefulTimeout  time.Duration
}

func newConfig() *config {
	c := &config{}
	c.allow.Clear()
	c.deny.Clear()
	c.interruptTimeout = 1 * time.Second
	c.gracefulTimeout = 5 * time.Second
	return c
}

func (self *config) MustParse() {
	var strAllow, strDeny string
	// flag.StringVar(&configPath, "config", "", "/path/to/config.toml")
	flag.StringVar(&self.addrListen, "listen", "", "addr:port")
	flag.StringVar(&strAllow, "allow", "", "rule-allow 'cidr:port->cidr:port,...'")
	flag.StringVar(&strDeny, "deny", "", "rule-deny 'cidr:port->cidr:port,...'")
	flag.Parse()

	errs := make([]string, 0)
	errs = append(errs, self.allow.FromString("allow", strAllow)...)
	errs = append(errs, self.deny.FromString("deny", strDeny)...)
	if self.addrListen == "" {
		errs = append(errs, fmt.Sprintf("config parse: require listen"))
	}
	if self.allow.Len() == 0 {
		errs = append(errs, fmt.Sprintf("config parse: nothing is allowed by default"))
	}
	if len(errs) > 0 {
		log.Fatal(strings.Join(errs, "\n"))
	}
}

func (self *config) String() string {
	return fmt.Sprintf("listen=%s allow=%v deny=%s",
		self.addrListen,
		self.allow.String(),
		self.deny.String())
}
