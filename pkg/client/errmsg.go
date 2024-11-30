package client

import "github.com/peakedshout/go-pandorasbox/tool/xerror"

var (
	ErrNotFoundNode   = xerror.New("not found node: %s")
	ErrNilNodes       = xerror.New("nil nodes")
	ErrDialNodeFailed = xerror.New("dial node failed")
	ErrLinkRefuse     = xerror.New("link refuse")
)
