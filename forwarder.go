package forwarder

import (
	"context"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

// SetUser changes the user used to make the SSH connection.
func (tun *ForwardConfig) SetUser(user string) *ForwardConfig {
	tun.User = user
	return tun
}

// SetConnState specifies an optional callback function that is called when a SSH tunnel changes state.
// See the ConnState type and associated constants for details.
func (tun *ForwardConfig) SetConnState(connStateFun func(*ForwardConfig, ConnState)) {
	tun.connState = connStateFun
}

func (tun *ForwardConfig) SetTunneledConnState(tunneledConnStateFun func(*ForwardConfig, *TunneledConnState)) {
	tun.tunneledConnState = tunneledConnStateFun
}

// NewUnix does the same as New but using unix sockets.
func NewUnix(localUnixSocket string, server string, remoteUnixSocket string) *ForwardConfig {
	sshTun := defaultSSHTun(server)
	sshTun.Local = NewUnixEndpoint(localUnixSocket)
	sshTun.Remote = NewUnixEndpoint(remoteUnixSocket)
	return sshTun
}

func NewUnixEndpoint(socket string) *Endpoint {
	return &Endpoint{
		unixSocket: socket,
	}
}

// NewUnixRemote does the same as NewRemote but using unix sockets.
func NewUnixRemote(localUnixSocket string, server string, remoteUnixSocket string) *ForwardConfig {
	sshTun := NewUnix(localUnixSocket, server, remoteUnixSocket)
	sshTun.forwardType = Remote
	return sshTun
}

func defaultSSHTun(server string) *ForwardConfig {
	return &ForwardConfig{
		aliveConnMutex:    &sync.RWMutex{},
		Server:            NewServerEndpoint(server, 22),
		User:              "root",
		timeout:           time.Second * 2,
		forwardType:       Remote,
		connState:         func(_config *ForwardConfig, _state ConnState) {},
		tunneledConnState: func(_config *ForwardConfig, _state *TunneledConnState) {},
	}
}

func NewServerEndpoint(host string, port int) *Endpoint {
	return &Endpoint{
		host: host,
		port: port,
	}
}

func (tun *ForwardConfig) initSSHConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            tun.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         tun.timeout,
	}

	authMethod, err := tun.getSSHAuthMethod()
	if err != nil {
		return nil, err
	}

	config.Auth = []ssh.AuthMethod{authMethod}

	return config, nil
}

// Stop closes all connections and makes Start exit `gracefully`.
func (tun *ForwardConfig) Stop() {
	if tun.started {
		tun.cancel()
	}
}

func (tun *ForwardConfig) notifyStop() {
	tun.started = false
	tun.connState(tun, StateStopped)
}

// CleanTargetSocketFile delete the target socket file before forward
func (tun *ForwardConfig) CleanTargetSocketFile() error {
	sshClient, err := ssh.Dial(tun.Server.Type(), tun.Server.String(), tun.SSHConfig)
	if err != nil {
		return fmt.Errorf("SSH Dial Error: %v", err)
	}
	mySession, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("SSH NewSession Error: %v", err)
	}
	mySession.Stdin = nil
	rSocket := tun.Remote.String()

	cmdline := fmt.Sprintf("sh -c 'mkdir -p %s && rm -f %s '", filepath.Dir(rSocket), rSocket)
	if err = mySession.Run(cmdline); err != nil {
		return fmt.Errorf("SSH Run Error: %v", err)
	}
	defer mySession.Close()
	return nil
}

func (tun *ForwardConfig) Start(ctx context.Context) error {
	if tun.started {
		return fmt.Errorf("already started")
	}
	tun.started = true

	defer tun.notifyStop()
	return tun.start(ctx)
}

func (tun *ForwardConfig) start(ctx context.Context) error {
	tun.ctx, tun.cancel = context.WithCancel(ctx)

	config, err := tun.initSSHConfig()
	if err != nil {
		return fmt.Errorf("ssh config failed: %w", err)
	}
	tun.SSHConfig = config

	tun.connState(tun, StateStarting)

	var listener net.Listener

	sshClient, err := ssh.Dial(tun.Server.Type(), tun.Server.String(), tun.SSHConfig)
	if err != nil {
		return fmt.Errorf("ssh dial %s to %s failed: %w", tun.Server.Type(), tun.Server.String(), err)
	}

	listener, err = sshClient.Listen(tun.Remote.Type(), tun.Remote.String())
	if err != nil {
		return fmt.Errorf("remote listen %s on %s failed: %w", tun.Remote.Type(), tun.Remote.String(), err)
	}
	defer listener.Close()

	tun.connState(tun, StateStarted)

	g := errgroup.Group{}
	g.Go(func() error {
		return tun.listen(listener)
	})

	return g.Wait()
}

func (tun *ForwardConfig) listen(listener net.Listener) error {
	for {
		if tun.ctx.Err() != nil {
			return fmt.Errorf("forward context cancelled: %w", tun.ctx.Err())
		}
		if conn, err := listener.Accept(); err == nil {
			tun.addConn()
			go func(conn net.Conn) {
				defer tun.removeConn()
				tun.forward(conn)
			}(conn)
		}
	}
}

func (tun *ForwardConfig) fromEndpoint() *Endpoint {
	if tun.forwardType == Remote {
		return tun.Remote
	}

	return tun.Local
}

func (tun *ForwardConfig) addConn() {
	tun.aliveConnMutex.Lock()
	defer tun.aliveConnMutex.Unlock()

	tun.active += 1
}

func (tun *ForwardConfig) removeConn() {
	tun.aliveConnMutex.Lock()
	defer tun.aliveConnMutex.Unlock()

	tun.active -= 1
}

func (tun *ForwardConfig) GetAliveConnCount() int {
	tun.aliveConnMutex.RLock()
	defer tun.aliveConnMutex.RUnlock()

	return tun.active
}

func (tun *ForwardConfig) tunneledState(state *TunneledConnState) {
	tun.tunneledConnState(tun, state)
}

func (tun *ForwardConfig) toEndpoint() *Endpoint {
	if tun.forwardType == Remote {
		return tun.Local
	}

	return tun.Remote
}

func (tun *ForwardConfig) forwardFromName() string {
	if tun.forwardType == Remote {
		return "Remote"
	}

	return "Local"
}

func (tun *ForwardConfig) forwardToName() string {
	if tun.forwardType == Remote {
		return "Local"
	}

	return "Remote"
}

func (tun *ForwardConfig) forward(fromConn net.Conn) {
	from := fromConn.RemoteAddr().String()

	tun.tunneledState(&TunneledConnState{
		From: from,
		Info: fmt.Sprintf("accepted %s connection", tun.fromEndpoint().Type()),
	})

	dialFunc := tun.SSHClient.Dial
	if tun.forwardType == Remote {
		dialFunc = net.Dial
	}

	toConn, err := dialFunc(tun.toEndpoint().Type(), tun.toEndpoint().String())
	if err != nil {
		tun.tunneledState(&TunneledConnState{
			From: from,
			Error: fmt.Errorf("%s dial %s to %s failed: %w", tun.forwardToName(),
				tun.toEndpoint().Type(), tun.toEndpoint().String(), err),
		})

		_ = fromConn.Close()
		return
	}

	connStr := fmt.Sprintf("%s -(%s)> %s <(ssh)> %s -(%s)> %s", from, tun.fromEndpoint().Type(),
		tun.fromEndpoint().String(), tun.Server.String(), tun.toEndpoint().Type(), tun.toEndpoint().String())

	tun.tunneledState(&TunneledConnState{
		From:   from,
		Info:   fmt.Sprintf("connection established: %s", connStr),
		Ready:  true,
		Closed: false,
	})

	connCtx, connCancel := context.WithCancel(tun.ctx)
	errGroup := &errgroup.Group{}

	errGroup.Go(func() error {
		defer connCancel()
		_, err = io.Copy(toConn, fromConn)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			return fmt.Errorf("failed copying bytes from %s to %s: %w", tun.forwardToName(), tun.forwardFromName(), err)
		}
		return nil
	})

	errGroup.Go(func() error {
		defer connCancel()
		_, err = io.Copy(fromConn, toConn)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			return fmt.Errorf("failed copying bytes from %s to %s: %w", tun.forwardFromName(), tun.forwardToName(), err)
		}
		return nil
	})

	<-connCtx.Done()
	_ = fromConn.Close()
	_ = toConn.Close()

	err = errGroup.Wait()
	if err != nil {
		tun.tunneledState(&TunneledConnState{
			From:   from,
			Error:  err,
			Closed: true,
		})
	}

	tun.tunneledState(&TunneledConnState{
		From:   from,
		Info:   fmt.Sprintf("connection closed: %s", connStr),
		Closed: true,
	})
}
