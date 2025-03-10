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
	if e.host != "" && e.port != 0 {
		return endpointTypeTcp
	}

	return endpointTypeUnknown
}

const (
	endpointTypeUnixSocket = "unix"
	endpointTypeTcp        = "tcp"
	endpointTypeUnknown    = "unknown"
)

type Endpoint struct {
	host       string
	port       int
	unixSocket string
}
