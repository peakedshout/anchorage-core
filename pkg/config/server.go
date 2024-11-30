package config

import (
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/xnet"
	"net"
)

type ServerConfig struct {
	NodeInfo         NodeConfig   `json:"nodeInfo" yaml:"nodeInfo" comment:"current server node config"`
	SyncNodes        []NodeConfig `json:"syncNodes" yaml:"syncNodes" comment:"sync server node config"`
	SyncTimeInterval uint         `json:"syncTimeInterval" yaml:"syncTimeInterval" comment:"sync server time interval (unit ms)"` // ms
	LinkTimeout      uint         `json:"linkTimeout" yaml:"linkTimeout" comment:"link time out (unit ms)"`                       // ms
	ProxyMulti       int          `json:"proxyMulti" yaml:"proxyMulti" comment:"whether support proxy multi io count to link"`
}

func (sc *ServerConfig) Check() error {
	if sc == nil {
		return errors.New("nil server config")
	}
	var errs []error
	errs = append(errs, sc.NodeInfo.Check())
	for _, node := range sc.SyncNodes {
		errs = append(errs, node.Check())
	}
	return errors.Join(errs...)
}

type NodeConfig struct {
	NodeName         string              `json:"nodeName" yaml:"nodeName" comment:"must be"`
	BaseNetwork      []BaseNetworkConfig `json:"baseNetwork" yaml:"baseNetwork"  comment:"base network list"`
	ExNetworks       []ExNetworkConfig   `json:"exNetworks" yaml:"exNetworks" comment:"expand network list"`
	Crypto           []CryptoConfig      `json:"crypto" yaml:"crypto" comment:"crypto config"`
	HandshakeTimeout uint                `json:"handshakeTimeout" yaml:"handshakeTimeout" comment:"handshake timeout (unit ms)"` // ms
	HandleTimeout    uint                `json:"handleTimeout" yaml:"handleTimeout" comment:"handle timeout (unit ms)"`          // ms
	Auth             *AuthInfo           `json:"auth" yaml:"auth" comment:"node auth ( username  password )"`
}

func (nc *NodeConfig) Check() error {
	if nc == nil {
		return errors.New("nil node config")
	}
	var errs []error
	if nc.NodeName == "" {
		errs = append(errs, errors.New("nil node name"))
	}
	for _, one := range nc.BaseNetwork {
		errs = append(errs, one.Check())
	}
	for _, one := range nc.ExNetworks {
		errs = append(errs, one.Check())
	}
	for _, one := range nc.Crypto {
		errs = append(errs, one.Check())
	}
	return errors.Join(errs...)
}

type BaseNetworkConfig struct {
	Network string `json:"network" yaml:"network" comment:"must be tcp,udp(quic)"`
	Address string `json:"address" yaml:"address" comment:"work address"`
}

func (bc *BaseNetworkConfig) Check() error {
	if bc == nil {
		return errors.New("nil base network config")
	}
	var errs []error
	if xnet.GetStdBaseNetwork(bc.Network) == "" {
		errs = append(errs, fmt.Errorf("not support network: %s", bc.Network))
	}
	_, _, err := net.SplitHostPort(bc.Address)
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

type ExNetworkConfig struct {
	Network            string `json:"network" yaml:"network" comment:"must be tcp,udp,quic,tls,ws,wss,http,https; if tcp or udp will without; tls, wss and https must have cert; if ws and http has cert will up grader to wss or https; quic optional cert."`
	CertFile           string `json:"certFile" yaml:"certFile" comment:"cert file path"`
	KeyFile            string `json:"keyFile" yaml:"keyFile" comment:"key file path"`
	CertRaw            string `json:"certRaw" yaml:"certRaw" comment:"raw cert"`
	KeyRaw             string `json:"keyRaw" yaml:"keyRaw" comment:"raw key"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify" comment:"cert insecure skip verify"`
}

func (enc *ExNetworkConfig) Check() error {
	if enc == nil {
		return errors.New("nil ex network config")
	}
	var errs []error
	if (enc.CertFile == "" && enc.KeyFile != "") || (enc.CertFile != "" && enc.KeyFile == "") ||
		(len(enc.CertRaw) == 0 && len(enc.KeyRaw) != 0) || (len(enc.CertRaw) != 0 && len(enc.KeyRaw) == 0) {
		errs = append(errs, errors.New("certificate and key do not match"))
	}
	switch enc.Network {
	case "tcp", "udp", "quic", "ws", "http":
	case "tls", "wss", "https":
	default:
		errs = append(errs, fmt.Errorf("not support network: %s", enc.Network))
	}
	return errors.Join(errs...)
}

type CryptoConfig struct {
	Name     string   `json:"name" yaml:"name" comment:"crypto expand name"`
	Crypto   string   `json:"crypto" yaml:"crypto" comment:"crypto type"`
	KeyFiles []string `json:"keyFiles" yaml:"keyFiles" comment:"key flies; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key."`
	Keys     []string `json:"keys" yaml:"keys" comment:"raw key; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key."`
	Priority int8     `json:"priority" yaml:"priority" comment:"crypto priority"`
}

func (cc *CryptoConfig) Check() error {
	if cc == nil {
		return errors.New("nil crypto config")
	}
	if cc.Crypto != "_" {
		symmetric, err := pcrypto.IsSymmetric(cc.Crypto)
		if err != nil {
			return fmt.Errorf("invalid crypto type: %s", cc.Crypto)
		}
		if symmetric {
			if len(cc.Keys) == 1 && len(cc.KeyFiles) == 0 {
			} else if len(cc.KeyFiles) == 1 && len(cc.Keys) == 0 {
			} else {
				return errors.New("invalid key group")
			}
		} else {
			if len(cc.Keys) == 2 && len(cc.KeyFiles) == 0 {
			} else if len(cc.KeyFiles) == 2 && len(cc.Keys) == 0 {
			} else {
				return errors.New("invalid key group")
			}
		}
	} else {
		if len(cc.Keys) != 0 || len(cc.KeyFiles) != 0 {
			return errors.New("invalid key group")
		}
	}
	return nil
}

type AuthInfo struct {
	UserName string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

func (ai *AuthInfo) Equal(obj *AuthInfo) bool {
	if (ai == nil && obj != nil) || (ai != nil && obj == nil) {
		return false
	} else if ai == nil && obj == nil {
		return true
	} else if ai.UserName == obj.UserName && ai.Password == ai.Password {
		return true
	} else {
		return false
	}
}
