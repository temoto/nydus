package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type NetPort struct {
	net  net.IPNet
	port int
}

func (self *NetPort) String() string {
	port := "*"
	if self.port > 0 {
		port = strconv.Itoa(self.port)
	}
	return self.net.String() + ":" + port
}

// parse 1.2.3.4/24:80
func (self *NetPort) FromString(s string) error {
	const invalid = "Invalid cidr:port '%s'. Examples: 1.2.3.4/24:80 [2c08::f1]/64:443. Any port: '*'. Error: "
	isep := strings.IndexByte(s, '/')
	iport := strings.LastIndexByte(s, ':')
	if iport <= isep {
		return fmt.Errorf(invalid+"no / or port", s)
	}
	np := NetPort{}
	s1, s2 := s[:iport], s[iport+1:]
	var err error
	var n *net.IPNet
	if _, n, err = net.ParseCIDR(s1); err != nil {
		return fmt.Errorf(invalid+err.Error(), s)
	}
	np.net = *n
	if s2 == "*" {
		np.port = 0
	} else {
		if np.port, err = strconv.Atoi(s2); err != nil {
			return fmt.Errorf(invalid+err.Error(), s)
		}
	}
	*self = np
	return nil
}

func (self *NetPort) Match(name string, target *NetPort) bool {
	okIP := self.net.Contains(target.net.IP)
	okPort := self.port == 0 || self.port == target.port
	return okIP && okPort
}

type Spec struct {
	src NetPort
	dst NetPort
}

func (self *Spec) String() string {
	return self.src.String() + "->" + self.dst.String()
}

func (self *Spec) FromString(s string) error {
	i := strings.Index(s, "->")
	if i < 0 {
		return fmt.Errorf("Spec '%s' is invalid. Expected 'cidr:port->cidr:port'. Separator '->' not found.", s)
	}
	s1, s2 := s[:i], s[i+2:]
	spec := Spec{}
	if err := spec.src.FromString(s1); err != nil {
		return err
	}
	if err := spec.dst.FromString(s2); err != nil {
		return err
	}
	*self = spec
	return nil
}

// src,dst have all-ff mask
func (self *Spec) Match(name string, src, dst *NetPort) bool {
	okSrc := self.src.Match(name, src)
	okDst := self.dst.Match(name, dst)
	return okSrc && okDst
}

type SpecList struct {
	sync.Mutex
	l     []Spec
	index map[string]int
}

func NewSpecList() *SpecList {
	self := &SpecList{}
	self.Clear()
	return self
}

func (self *SpecList) Clear() {
	self.Lock()
	self.l = make([]Spec, 0, 8)
	self.index = make(map[string]int)
	self.Unlock()
}

func (self *SpecList) unsafe_add(s Spec) {
	ss := s.String()
	if _, ok := self.index[ss]; !ok {
		self.l = append(self.l, s)
		self.index[ss] = len(self.l) - 1
	}
}

func (self *SpecList) Add(s Spec) {
	self.Lock()
	self.unsafe_add(s)
	self.Unlock()
}

func (self *SpecList) Has(s Spec) bool {
	self.Lock()
	_, ok := self.index[s.String()]
	self.Unlock()
	return ok
}

func (self *SpecList) Len() int {
	self.Lock()
	n := len(self.l)
	self.Unlock()
	return n
}

func (self *SpecList) String() string {
	self.Lock()
	if len(self.l) == 0 {
		self.Unlock()
		return ""
	}
	ss := make([]string, len(self.l))
	for i, s := range self.l {
		ss[i] = s.String()
	}
	self.Unlock()
	return strings.Join(ss, ",")
}

func (self *SpecList) FromString(name, value string) (errs []string) {
	if value == "" {
		return nil
	}
	errs = make([]string, 0, 4)
	ss := strings.Split(value, ",")
	for _, s := range ss {
		if s == "" {
			log.Fatalf("config parse: name=%s has invalid empty value. value='%s'", name, value)
		}
		spec := Spec{}
		if err := spec.FromString(s); err != nil {
			errs = append(errs, fmt.Sprintf("config parse: name=%s item=%s error=%s", name, s, err))
			continue
		}
		self.Add(spec)
	}
	return errs
}

func (self *SpecList) Match(name string, src, dst *NetPort) bool {
	for _, spec := range self.l {
		if spec.Match(name, src, dst) {
			return true
		}
	}
	return false
}
