Forward unix domain socket file over SSH
=====================

Origin code from [daytona](https://github.com/daytonaio/daytona/tree/c9476324ae5cca3a8d1e772c72297de41213c83a/pkg/tailscale/tunnel)

This is a simple implementation of a forward UDF over SSH. It is used to forward a UDF server to a remote machine.

```go
package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"oosshagent/pkg/forwarder"
)

func main() {
	myForwarder := forwarder.NewUnixRemote("/var/folders/dt/wqkv0wf13nl6n8jbf98jggjh0000gn/T/ssh-px37G3xkyelq/agent.4973", "192.168.1.250", "/tmp/my_socks")
	myForwarder.SetTunneledConnState(func(tun *forwarder.ForwardConfig, state *forwarder.TunneledConnState) {
		logrus.Infof("%v", state)
	})

	myForwarder.SetKeyFile("/Users/danhexon/.ssh/id_ed25519")
	myForwarder.SetUser("ihexon")
	myForwarder.SetPort(22)

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

	if err := myForwarder.Start(context.Background()); err != nil {
		logrus.Infof("SSH tunnel error: %v", err)
	}
}

```


