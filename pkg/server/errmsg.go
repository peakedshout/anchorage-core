package server

import "github.com/peakedshout/go-pandorasbox/tool/xerror"

var (
	ErrNodeAuthFailed       = xerror.New("node auth failed")
	ErrRegisterListenerInfo = xerror.New("register listener info err: %v")
	ErrInvalidLinkId        = xerror.New("invalid link id")
	ErrNotFoundService      = xerror.New("not found service: %s")
	ErrNotFoundNode         = xerror.New("not found node: %s")
	ErrInvalidNode          = xerror.New("invalid node: %s")
	ErrLinkNodeFailed       = xerror.New("link node failed")
)
