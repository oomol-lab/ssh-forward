package main

import (
	"context"
	forwarder "github.com/oomol-lab/ssh-forward"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	myForwarder := forwarder.NewUnixRemote("/var/folders/dt/wqkv0wf13nl6n8jbf98jggjh0000gn/T/ssh-px37G3xkyelq/agent.4973", "192.168.1.250", "/tmp/my_socks/ssh-auth.sock")
	myForwarder.SetTunneledConnState(func(tun *forwarder.ForwardConfig, state *forwarder.TunneledConnState) {
		logrus.Infof("%v", state)
	})

	myForwarder.SetKeyFile("/Users/danhexon/.ssh/id_ed25519").SetUser("ihexon").SetPort(22)

	// We set a callback to know when the tunnel is ready
	myForwarder.SetConnState(func(tun *forwarder.ForwardConfig, state forwarder.ConnState) {
		switch state {
		case forwarder.StateStarting:
			logrus.Infoln("STATE is Starting")
		case forwarder.StateStarted:
			logrus.Infoln("STATE is Started")
		case forwarder.StateStopped:
			logrus.Infoln("STATE is Stopped")
		}
	})

	if err := myForwarder.Start(ctx); err != nil {
		logrus.Infof("SSH tunnel error: %v", err)
	}
}
