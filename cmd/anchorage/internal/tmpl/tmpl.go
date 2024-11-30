package tmpl

import (
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/anchorage-core/pkg/sdk"
	"github.com/peakedshout/go-pandorasbox/tool/hyaml"
)

func makeServerTmpl() error {
	cfg := &sdk.ServerConfig{
		Enable: false,
		ServerConfig: &config.ServerConfig{
			NodeInfo: config.NodeConfig{
				NodeName: "",
				BaseNetwork: []config.BaseNetworkConfig{
					{
						Network: "",
						Address: "",
					},
				},
				ExNetworks: []config.ExNetworkConfig{
					{
						Network:            "",
						CertFile:           "",
						KeyFile:            "",
						CertRaw:            "",
						KeyRaw:             "",
						InsecureSkipVerify: false,
					},
				},
				Crypto: []config.CryptoConfig{
					{
						Name:     "",
						Crypto:   "",
						KeyFiles: nil,
						Keys:     nil,
						Priority: 0,
					},
				},
				HandshakeTimeout: 0,
				HandleTimeout:    0,
				Auth: &config.AuthInfo{
					UserName: "",
					Password: "",
				},
			},
			SyncNodes: []config.NodeConfig{
				{
					NodeName:         "",
					BaseNetwork:      nil,
					ExNetworks:       nil,
					Crypto:           nil,
					HandshakeTimeout: 0,
					HandleTimeout:    0,
					Auth:             nil,
				},
			},
			SyncTimeInterval: 0,
			LinkTimeout:      0,
		},
	}
	return hyaml.SavePathT("server.yaml", cfg)
}

func makeClientTmpl() error {
	cfg := &sdk.ClientConfigUnit{
		Enable: false,
		ClientConfig: &config.ClientConfig{
			Nodes: []config.NodeConfig{
				{
					NodeName: "",
					BaseNetwork: []config.BaseNetworkConfig{
						{
							Network: "",
							Address: "",
						},
					},
					ExNetworks: []config.ExNetworkConfig{
						{
							Network:            "",
							CertFile:           "",
							KeyFile:            "",
							CertRaw:            "",
							KeyRaw:             "",
							InsecureSkipVerify: false,
						},
					},
					Crypto: []config.CryptoConfig{
						{
							Name:     "",
							Crypto:   "",
							KeyFiles: nil,
							Keys:     nil,
							Priority: 0,
						},
					},
					HandshakeTimeout: 0,
					HandleTimeout:    0,
					Auth: &config.AuthInfo{
						UserName: "",
						Password: "",
					},
				},
			},
		},
	}
	return hyaml.SavePathT("client.yaml", cfg)
}

func makeListenTmpl() error {
	cfg := &sdk.ListenConfig{
		Enable: false,
		Node:   "",
		Name:   "",
		Notes:  "",
		Auth: &config.AuthInfo{
			UserName: "",
			Password: "",
		},
		SwitchHide: false,
		SwitchLink: false,
		SwitchUP2P: false,
		SwitchTP2P: false,
		OutNetwork: &sdk.NetworkConfig{
			Network: "",
			Address: "",
		},
		Multi:  false,
		Plugin: "",
	}
	return hyaml.SavePathT("listen.yaml", cfg)
}

func makeDialTmpl() error {
	cfg := &sdk.DialConfig{
		Enable:      false,
		Node:        nil,
		Link:        "",
		PP2PNetwork: "",
		SwitchUP2P:  false,
		SwitchTP2P:  false,
		ForceP2P:    false,
		Auth: &config.AuthInfo{
			UserName: "",
			Password: "",
		},
		InNetwork: &sdk.NetworkConfig{
			Network: "",
			Address: "",
		},
		OutNetwork: &sdk.NetworkConfig{
			Network: "",
			Address: "",
		},
		Multi:     0,
		MultiIdle: 0,
		Plugin:    "",
	}
	return hyaml.SavePathT("dial.yaml", cfg)
}

func makeProxyTmpl() error {
	cfg := &sdk.ProxyConfig{
		Enable: false,
		Node:   nil,
		InNetwork: &sdk.NetworkConfig{
			Network: "",
			Address: "",
		},
		OutNetwork: &sdk.NetworkConfig{
			Network: "",
			Address: "",
		},
		Plugin: "",
	}
	return hyaml.SavePathT("proxy.yaml", cfg)
}

func makePluginTmpl() error {
	cfg := &sdk.PluginConfig{
		Name: "",
		Type: []string{},
		List: []map[string]any{},
	}
	return hyaml.SavePathT("plugin.yaml", cfg)
}
