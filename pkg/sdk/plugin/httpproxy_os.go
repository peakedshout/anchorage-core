package plugin

import (
	"net"
)

var httpproxy_osProxySetting func(ln net.Listener) (func() error, error)
