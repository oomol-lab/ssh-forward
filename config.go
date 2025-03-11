package forwarder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type ForwardType int

const (
	Remote ForwardType = iota
)

type ConnState int

const (
	// StateStopped represents a stopped tunnel. A call to Start will make the state to transition to StateStarting.
	StateStopped ConnState = iota

	// StateStarting represents a tunnel initializing and preparing to listen for connections.
	// A successful initialization will make the state to transition to StateStarted, otherwise it will transition to StateStopped.
	StateStarting

	// StateStarted represents a tunnel ready to accept connections.
	// A call to stop or an error will make the state to transition to StateStopped.
	StateStarted
)

type TunneledConnState struct {
	// From is the address initating the connection.
	From string
	// Info holds a message with info on the state of the connection (useful for debug purposes).
	Info string
	// Error holds an error on the connection or nil if the connection is successful.
	Error error
	// Ready indicates if the connection is established.
	Ready bool
	// Closed indicates if the connection is closed.
	Closed bool
}

func (s *TunneledConnState) String() string {
	out := fmt.Sprintf("[%s] ", s.From)
	if s.Info != "" {
		out += s.Info
	}
	if s.Error != nil {
		out += fmt.Sprintf("Error: %v", s.Error)
	}
	return out
}

type ForwardConfig struct {
	mutex             *sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	User              string
	Server            *Endpoint
	started           bool
	Local             *Endpoint
	Remote            *Endpoint
	forwardType       ForwardType
	timeout           time.Duration
	active            int
	connState         func(*ForwardConfig, ConnState)
	tunneledConnState func(*ForwardConfig, *TunneledConnState)
	SSHClient         *ssh.Client
	SSHConfig         *ssh.ClientConfig
	authKeyFile       string
}
