package comm

import "github.com/peakedshout/anchorage-core/pkg/config"

const (
	CallRegister = "CallRegister"

	CallLink    = "CallLink"
	CallLinkReq = "CallLinkReq"

	CallSync    = "CallSync"
	CallSyncMap = "CallSyncMap"

	CallRouteView = "CallRouteView"

	CallProxy = "CallProxy"
)

const (
	NodeName   = "nodeName"
	KeyLinkId  = "linkId"
	KeyLinkLId = "linkLId"
	ProxyInfo  = "proxyInfo"
)

type RegisterListenerInfo struct {
	Node     string
	Name     string
	Notes    string
	Auth     *config.AuthInfo
	Settings Settings
}

func (rli *RegisterListenerInfo) Equal(info *LinkRequest) bool {
	if rli.Settings.SwitchHide || !rli.Settings.SwitchLink ||
		!rli.Auth.Equal(info.Auth) {
		return false
	}
	if info.ForceP2P && !rli.P2PCheck(info) {
		return false
	}
	return true
}

func (rli *RegisterListenerInfo) P2PCheck(info *LinkRequest) (enable bool) {
	if !info.SwitchUP2P && !info.SwitchTP2P {
		return false
	}
	if !rli.Settings.SwitchUP2P && !rli.Settings.SwitchTP2P {
		return false
	}
	// todo ???
	//if rli.Network != info.Network {
	//	return false
	//}
	return true
}

type Settings struct {
	SwitchHide bool
	SwitchLink bool
	SwitchUP2P bool
	SwitchTP2P bool
}

type LinkRequest struct {
	Node []string
	Link string

	PP2PNetwork string
	SwitchUP2P  bool
	SwitchTP2P  bool
	ForceP2P    bool
	Auth        *config.AuthInfo

	BoxId  uint64
	BoxLId string
}

func (lr *LinkRequest) GetNode() string {
	for _, node := range lr.Node {
		return node
	}
	return ""
}

func (lr *LinkRequest) RemoveNode() {
	if len(lr.Node) != 0 {
		lr.Node = lr.Node[1:]
	}
}

type LinkResponse struct {
	P2PNetwork string
	BoxId      uint64
	BoxLId     string
}
