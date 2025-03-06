package forwarder

import "fmt"

func (e *Endpoint) String() string {
	if e.unixSocket != "" {
		return e.unixSocket
	}
	return fmt.Sprintf("%s:%d", e.host, e.port)
}

func (e *Endpoint) Type() string {
	if e.unixSocket != "" {
		return endpointTypeUnixSocket
	}
	return endpointTypeUnknow
}

const (
	endpointTypeUnixSocket = "unix"
	endpointTypeUnknow     = "unknow"
)

type Endpoint struct {
	host       string
	port       int
	unixSocket string
}
