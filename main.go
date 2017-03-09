package main

import (
	"github.com/armon/go-socks5"
	"github.com/coreos/go-systemd/daemon"
	"golang.org/x/net/context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ruleset struct {
	config *config
}

func ipToNet(ip net.IP) net.IPNet {
	n := net.IPNet{
		IP:   ip,
		Mask: make(net.IPMask, len(ip)),
	}
	for i, _ := range ip {
		n.Mask[i] = 0xff
	}
	return n
}

func (self *ruleset) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	src := &NetPort{ipToNet(req.RemoteAddr.IP), req.RemoteAddr.Port}
	dst := &NetPort{ipToNet(req.DestAddr.IP), req.DestAddr.Port}
	allowed := self.config.allow.Match("allow", src, dst)
	denied := self.config.deny.Match("deny", src, dst)
	ok := allowed && !denied
	log.Printf("access-check src=%s dst=%s ok=%v", req.RemoteAddr.Address(), req.DestAddr.Address(), ok)
	return ctx, ok
}

func MustSdNotify(state string) {
	if _, err := daemon.SdNotify(false, state); err != nil {
		panic(err)
	}
}

func main() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	MustSdNotify("READY=1")
	if watchdogTimeout, err := daemon.SdWatchdogEnabled(false); err != nil && watchdogTimeout > 0 {
		go func() {
			const watchdogState = "WATCHDOG=1"
			for {
				time.Sleep(watchdogTimeout / 2)
				MustSdNotify(watchdogState)
			}
		}()
	}

	conf := newConfig()
	conf.MustParse()
	log.Printf("config: %s", conf.String())

	socksConf := &socks5.Config{
		Rules:  &ruleset{config: conf},
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	server, err := socks5.New(socksConf)
	if err != nil {
		panic(err)
	}
	connch := make(chan net.Conn, 8)
	var listener net.Listener
	if listener, err = net.Listen("tcp", conf.addrListen); err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	donech := make(chan struct{})
	stopListen := make(chan struct{}, 1)

	terminate := func(timeout time.Duration, exitCode int) {
		begin := time.Now()
		MustSdNotify("STOPPING=1")
		stopListen <- struct{}{}
		listener.Close()
		go func() { wg.Wait(); donech <- struct{}{} }()
		select {
		case <-donech:
		case <-time.After(timeout):
		}
		cleanDuration := time.Now().Sub(begin)
		log.Printf("remaining work finished in %s", cleanDuration)
		os.Exit(exitCode)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			select {
			case <-stopListen:
				return
			default:
				if err == nil {
					log.Printf("accept() remote=%s", conn.RemoteAddr().String())
					wg.Add(1)
					connch <- conn
				} else {
					log.Fatal(err)
				}
			}
		}
	}()

	for {
		select {
		case conn := <-connch:
			go func() {
				server.ServeConn(conn)
				wg.Done()
			}()
		case sig := <-sigch:
			if sig == os.Interrupt {
				log.Printf("caught SIGINT")
				terminate(conf.interruptTimeout, 1)
			}
			if sig == syscall.SIGTERM {
				log.Printf("caught SIGTERM")
				terminate(conf.gracefulTimeout, 0)
			}
			log.Fatalf("unexpected signal: %d", sig)
		}
	}
}
