package sdk

import (
	"errors"
	"github.com/peakedshout/anchorage-core/pkg/config"
)

type Config struct {
	Server []*ServerConfig     `json:"server" yaml:"server" comment:"server part"`
	Client []*ClientConfig     `json:"client" yaml:"client" comment:"client part"`
	Logger config.LoggerConfig `json:"logger" yaml:"logger" comment:"logger config"`
}

type ServerConfig struct {
	Enable               bool `json:"enable" yaml:"enable" comment:"loaded then to work"`
	*config.ServerConfig `json:"config" yaml:"config" comment:"server config"`
}

type ClientConfig struct {
	*ClientConfigUnit `json:"config" yaml:"config" comment:"client config"`
	Listen            []*ListenConfig `json:"listen" yaml:"listen" comment:"listen service list"`
	Dial              []*DialConfig   `json:"dial" yaml:"dial" comment:"dial service list"`
	Proxy             []*ProxyConfig  `json:"proxy" yaml:"proxy" comment:"proxy service list"`
	Plugin            []*PluginConfig `json:"plugin" yaml:"plugin" comment:"plugin list"`
}

type ClientConfigUnit struct {
	Enable               bool `json:"enable" yaml:"enable" comment:"loaded then to work"`
	*config.ClientConfig `json:"config" yaml:"config" comment:"client config"`
}

type ListenConfig struct {
	Enable     bool             `json:"enable" yaml:"enable" comment:"loaded then to work"`
	Node       string           `json:"node" yaml:"node" comment:"register a specified node"`
	Name       string           `json:"name" yaml:"name" comment:"service name"`
	Notes      string           `json:"notes" yaml:"notes" comment:"service notes"`
	Auth       *config.AuthInfo `json:"auth" yaml:"auth" comment:"service auth ( username  password )"`
	SwitchHide bool             `json:"switchHide" yaml:"switchHide" comment:"not to be discovered by others"`
	SwitchLink bool             `json:"switchLink" yaml:"switchLink" comment:"whether or not to allow the link"`
	SwitchUP2P bool             `json:"switchUP2P" yaml:"switchUP2P" comment:"whether support udp p2p to link"`
	SwitchTP2P bool             `json:"switchTP2P" yaml:"switchTP2P" comment:"whether support tcp p2p to link"`

	OutNetwork *NetworkConfig `json:"outNetwork" yaml:"outNetwork" comment:"out network config"`
	Multi      bool           `json:"multi" yaml:"multi" comment:"whether support multi io to link"`
	Plugin     string         `json:"plugin" yaml:"plugin" comment:"plugin name"`
}

func (lc *ListenConfig) Check() error {
	if lc == nil {
		return errors.New("nil listen config")
	}
	if lc.Name == "" {
		return errors.New("nil name")
	}
	return nil
}

type DialConfig struct {
	Enable      bool             `json:"enable" yaml:"enable" comment:"loaded then to work"`
	Node        []string         `json:"node" yaml:"node" comment:"node links"`
	Link        string           `json:"link" yaml:"link"  comment:"link service name"`
	PP2PNetwork string           `json:"PP2PNetwork" yaml:"PP2PNetwork" comment:"specify the p2p network type"`
	SwitchUP2P  bool             `json:"switchUP2P" yaml:"switchUP2P" comment:"whether support udp p2p to link"`
	SwitchTP2P  bool             `json:"switchTP2P" yaml:"switchTP2P" comment:"whether support tcp p2p to link"`
	ForceP2P    bool             `json:"forceP2P" yaml:"forceP2P" comment:"whether force p2p to link"`
	Auth        *config.AuthInfo `json:"auth" yaml:"auth" comment:"service auth ( username  password )"`

	InNetwork  *NetworkConfig `json:"inNetwork" yaml:"inNetwork" comment:"in network config"`
	OutNetwork *NetworkConfig `json:"outNetwork" yaml:"outNetwork" comment:"out network config"`
	Multi      int            `json:"multi" yaml:"multi" comment:"whether support multi io count to link (set -1 to close)"`
	MultiIdle  int            `json:"multiIdle" yaml:"multiIdle" comment:"whether support multi io idle count to link"`
	Plugin     string         `json:"plugin" yaml:"plugin" comment:"plugin name"`
}

func (dc *DialConfig) Check() error {
	if dc == nil {
		return errors.New("nil dial config")
	}
	if dc.Link == "" {
		return errors.New("nil link")
	}
	if dc.InNetwork == nil {
		return errors.New("nil in network")
	}
	return nil
}

type NetworkConfig struct {
	Network string `json:"network" yaml:"network" comment:"must be tcp or udp"`
	Address string `json:"address" yaml:"address"`
}

type ProxyConfig struct {
	Enable     bool           `json:"enable" yaml:"enable" comment:"loaded then to work"`
	Node       []string       `json:"node" yaml:"node" comment:"node links"`
	InNetwork  *NetworkConfig `json:"inNetwork" yaml:"inNetwork" comment:"in network config"`
	OutNetwork *NetworkConfig `json:"outNetwork" yaml:"outNetwork" comment:"out network config"`

	Multi  int    `json:"multi" yaml:"multi" comment:"whether support multi io count to link"`
	Plugin string `json:"plugin" yaml:"plugin" comment:"plugin name"`
}

func (pc *ProxyConfig) Check() error {
	if pc == nil {
		return errors.New("nil proxy config")
	}
	if pc.InNetwork == nil {
		return errors.New("nil in network")
	}
	return nil
}

type PluginConfig struct {
	Name string           `json:"name" yaml:"name" comment:"plugin name"`
	Type []string         `json:"type" yaml:"type" comment:"plugin type (dial,listen,proxy)"`
	List []map[string]any `json:"list" yaml:"list" comment:"list plugin setting"`
	//Args map[string]string `json:"args" yaml:"args" comment:"plugin args"`
}

func (pc *PluginConfig) Check() error {
	if pc == nil {
		return errors.New("nil plugin config")
	}
	if pc.Name == "" {
		return errors.New("plugin nil name")
	}
	return nil
}

const (
	PluginTypeDial   = "dial"
	PluginTypeListen = "listen"
	PluginTypeProxy  = "proxy"
)

//const (
//	PluginTypeSocks  = "socks"
//	PluginTypeSocks4 = "socks4"
//	PluginTypeSocks5 = "socks5"
//	PluginTypeHttp   = "http"
//	PluginTypeHttps  = "https"
//
//	PluginArgsAuthUserName = "username"
//	PluginArgsAuthPassword = "password"
//)
