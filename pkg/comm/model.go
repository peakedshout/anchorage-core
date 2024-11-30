package comm

import (
	"github.com/peakedshout/go-pandorasbox/xnet/xnetutil"
	"time"
)

const (
	LinkStatusInit = "init"
	LinkStatusWork = "work"
	LinkStatusWait = "wait"
	LinkStatusDead = "dead"
)

type LinkView struct {
	Id                uint64                  `json:"id"`
	Link              string                  `json:"link"`
	Status            string                  `json:"status"`
	SIT               [3]string               `json:"SIT"`
	InitSTSessionId   [2]string               `json:"initSTSessionId"`
	WorkSTStreamId    [2]string               `json:"workSTStreamId"`
	WorkSTMonitorInfo [2]xnetutil.MonitorInfo `json:"workSTMonitorInfo"`
	LinkId            string                  `json:"linkId"` // identification bit instead of reliable uuid
	LinkList          []string                `json:"linkList"`
	NodeList          []string                `json:"nodeList"`
}

type ServiceRouteView struct {
	NodeView    map[string]map[string][]ServiceRouteViewUnit `json:"nodeView"`    //k1 node k2 service v unit
	ServiceView map[string][]ServiceRouteViewUnit            `json:"serviceView"` //k1 service v unit
}

type ServiceRouteViewUnit struct {
	Name     string        `json:"name"`
	Node     string        `json:"node"`
	Notes    string        `json:"notes"`
	Auth     bool          `json:"auth"`
	Settings Settings      `json:"settings"`
	Delay    time.Duration `json:"delay"`
}

type ProxyRequest struct {
	Node    []string
	Network string
	Address string
}

func (pr *ProxyRequest) GetNode(localNode string) (string, []string) {
	if len(pr.Node) == 0 {
		return "", nil
	}
	for i, n := range pr.Node {
		if n != "" && n != localNode {
			return n, pr.Node[i+1:]
		}
	}
	return "", nil
}
