package sdk

import (
	"github.com/peakedshout/anchorage-core/pkg/config"
)

type ServerView struct {
	Id     string               `json:"id"`
	Status bool                 `json:"status"`
	Enable bool                 `json:"enable"`
	Config *config.ServerConfig `json:"config"`
}

type ClientView struct {
	Id     string               `json:"id"`
	Status bool                 `json:"status"`
	Enable bool                 `json:"enable"`
	Config *config.ClientConfig `json:"config"`
	Listen []*ListenView        `json:"listen"`
	Dial   []*DialView          `json:"dial"`
	Proxy  []*ProxyView         `json:"proxy"`
	Plugin []*PluginConfig      `json:"plugin"`
}

type ListenView struct {
	Id     string `json:"id"`
	Status bool   `json:"status"`
	*ListenConfig
}

type DialView struct {
	Id     string `json:"id"`
	Status bool   `json:"status"`
	*DialConfig
}

type ProxyView struct {
	Id     string `json:"id"`
	Status bool   `json:"status"`
	*ProxyConfig
}
