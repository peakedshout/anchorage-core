package sdk

import (
	"context"
	"fmt"
	"github.com/peakedshout/anchorage-core/pkg/config"
	"github.com/peakedshout/anchorage-core/pkg/sdk/plugin"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"github.com/peakedshout/go-pandorasbox/tool/hyaml"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"github.com/peakedshout/go-pandorasbox/xnet/fasttool"
	"github.com/peakedshout/go-pandorasbox/xnet/proxy/socks"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func newAddr() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("127.0.0.1:%d", r.Intn(10000)+10000)
}

func TestSdkManager(t *testing.T) {
	ln, err := fasttool.EchoTcpListener()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	hs, addh, err := fasttool.EchoHttp()
	if err != nil {
		t.Fatal(err)
	}
	defer hs.Close()
	addr1 := newAddr()
	addr2 := newAddr()
	addr3 := newAddr()
	addr4 := newAddr()
	addr5 := newAddr()
	addr0 := newAddr()

	addr6 := newAddr()
	addr7 := newAddr()
	tmp := hyaml.Load[Config]("./cfg.yaml", &Config{
		Server: []*ServerConfig{{
			Enable: true,
			ServerConfig: &config.ServerConfig{
				NodeInfo: config.NodeConfig{
					NodeName: "node1",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr1,
					}, {
						Network: "udp",
						Address: addr2,
					}},
				},
				SyncNodes: []config.NodeConfig{{
					NodeName: "node2",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr3,
					}, {
						Network: "udp",
						Address: addr4,
					}},
				}},
			}}, {
			Enable: true,
			ServerConfig: &config.ServerConfig{
				NodeInfo: config.NodeConfig{
					NodeName: "node2",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr3,
					}, {
						Network: "udp",
						Address: addr4,
					}},
				},
				SyncNodes: []config.NodeConfig{{
					NodeName: "node1",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr1,
					}, {
						Network: "udp",
						Address: addr2,
					}},
				}},
			}},
		},
		Client: []*ClientConfig{
			{
				ClientConfigUnit: &ClientConfigUnit{
					Enable: true,
					ClientConfig: &config.ClientConfig{
						Nodes: []config.NodeConfig{{
							NodeName: "node1",
							BaseNetwork: []config.BaseNetworkConfig{{
								Network: "tcp",
								Address: addr1,
							}},
							ExNetworks:       nil,
							Crypto:           nil,
							HandshakeTimeout: 0,
							HandleTimeout:    0,
						}},
					},
				},
				Listen: []*ListenConfig{{
					Enable:     true,
					Name:       "test1",
					Notes:      "wwwwww",
					Auth:       nil,
					SwitchHide: false,
					SwitchLink: true,
					SwitchUP2P: false,
					SwitchTP2P: false,
					Multi:      true,
					Plugin:     "",
				}},
				Dial: []*DialConfig{{
					Enable:      true,
					Node:        nil,
					Link:        "test1",
					PP2PNetwork: "",
					SwitchUP2P:  false,
					SwitchTP2P:  false,
					ForceP2P:    false,
					Auth:        nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr0,
					},
					OutNetwork: &NetworkConfig{
						Network: ln.Addr().Network(),
						Address: ln.Addr().String(),
					},
					Multi:     0,
					MultiIdle: 0,
					Plugin:    "",
				}},
				Proxy: []*ProxyConfig{{
					Enable: true,
					Node:   nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr5,
					},
					OutNetwork: &NetworkConfig{
						Network: ln.Addr().Network(),
						Address: ln.Addr().String(),
					},
					Plugin: "",
				}},
			}, {
				ClientConfigUnit: &ClientConfigUnit{
					Enable: true,
					ClientConfig: &config.ClientConfig{
						Nodes: []config.NodeConfig{{
							NodeName: "node2",
							BaseNetwork: []config.BaseNetworkConfig{{
								Network: "tcp",
								Address: addr3,
							}},
							ExNetworks:       nil,
							Crypto:           nil,
							HandshakeTimeout: 0,
							HandleTimeout:    0,
						}},
					},
				},
				Listen: nil,
				Dial: []*DialConfig{{
					Enable:      true,
					Node:        nil,
					Link:        "test1",
					PP2PNetwork: "",
					SwitchUP2P:  false,
					SwitchTP2P:  false,
					ForceP2P:    false,
					Auth:        nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr6,
					},
					OutNetwork: nil,
					Multi:      0,
					MultiIdle:  0,
					Plugin:     "socks",
				}},
				Proxy: []*ProxyConfig{{
					Enable: true,
					Node:   nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr7,
					},
					OutNetwork: nil,
					Plugin:     "http",
				}},
				Plugin: []*PluginConfig{{
					Name: "socks",
					Type: []string{PluginTypeDial, PluginTypeProxy},
					List: []map[string]any{
						{
							"name":       plugin.NameSocks,
							"v4":         true,
							"v5":         true,
							"CMDCONNECT": true,
							"S5Auth": [][2]string{
								{"user", "pass"},
							},
						},
					},
				}, {
					Name: "http",
					Type: []string{PluginTypeDial, PluginTypeProxy},
					List: []map[string]any{
						{
							"name": plugin.NameHttpProxy,
							"Auth": [][2]string{
								{"user", "pass"},
							},
						},
					},
				}},
			},
		},
	})
	_ = tmp.Save()
	manager, err := newSdkManager(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Stop()
	time.Sleep(2 * time.Second)
	fmt.Println(manager.sList[0].server.GetRouteView())

	conn, err := net.Dial("tcp", addr0)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}

	conn2, err := net.Dial("tcp", addr5)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn2.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn2, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}

	dr, err := socks.SOCKS5CONNECTP("tcp", addr6, &socks.S5AuthPassword{
		User:     "user",
		Password: "pass",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	conn3, err := dr.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn3.Close()
	for i := 0; i < 10000; i++ {
		b := []byte(uuid.NewIdn(4096))
		_, err = conn3.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 4096)
		_, err = io.ReadFull(conn3, buf)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != string(buf) {
			t.Fatal()
		}
	}

	parse, err := url.Parse(fmt.Sprintf("http://user:pass@%s", addr7))
	if err != nil {
		t.Fatal(err)
	}

	c := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parse)}}
	resp, err := c.Get(addh)
	if err != nil || resp.StatusCode != 200 {
		t.Fatal(err)
	}

	sv := manager.GetServerView()
	fmt.Println(hjson.MustMarshalStr(sv))
	cv := manager.GetClientView()
	fmt.Println(hjson.MustMarshalStr(cv))
}

func testCfg(ln net.Listener) *Config {
	addr1 := newAddr()
	addr2 := newAddr()
	addr3 := newAddr()
	addr4 := newAddr()
	addr5 := newAddr()
	addr0 := newAddr()

	addr6 := newAddr()
	addr7 := newAddr()
	return &Config{
		Server: []*ServerConfig{{
			Enable: true,
			ServerConfig: &config.ServerConfig{
				NodeInfo: config.NodeConfig{
					NodeName: "node1",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr1,
					}, {
						Network: "udp",
						Address: addr2,
					}},
				},
				SyncNodes: []config.NodeConfig{{
					NodeName: "node2",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr3,
					}, {
						Network: "udp",
						Address: addr4,
					}},
				}},
			}}, {
			Enable: true,
			ServerConfig: &config.ServerConfig{
				NodeInfo: config.NodeConfig{
					NodeName: "node2",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr3,
					}, {
						Network: "udp",
						Address: addr4,
					}},
				},
				SyncNodes: []config.NodeConfig{{
					NodeName: "node1",
					BaseNetwork: []config.BaseNetworkConfig{{
						Network: "tcp",
						Address: addr1,
					}, {
						Network: "udp",
						Address: addr2,
					}},
				}},
			}},
		},
		Client: []*ClientConfig{
			{
				ClientConfigUnit: &ClientConfigUnit{
					Enable: true,
					ClientConfig: &config.ClientConfig{
						Nodes: []config.NodeConfig{{
							NodeName: "node1",
							BaseNetwork: []config.BaseNetworkConfig{{
								Network: "tcp",
								Address: addr1,
							}},
							ExNetworks:       nil,
							Crypto:           nil,
							HandshakeTimeout: 0,
							HandleTimeout:    0,
						}},
					},
				},
				Listen: []*ListenConfig{{
					Enable:     true,
					Name:       "test1",
					Notes:      "wwwwww",
					Auth:       nil,
					SwitchHide: false,
					SwitchLink: true,
					SwitchUP2P: false,
					SwitchTP2P: false,
					Multi:      true,
					Plugin:     "",
				}},
				Dial: []*DialConfig{{
					Enable:      true,
					Node:        nil,
					Link:        "test1",
					PP2PNetwork: "",
					SwitchUP2P:  false,
					SwitchTP2P:  false,
					ForceP2P:    false,
					Auth:        nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr0,
					},
					OutNetwork: &NetworkConfig{
						Network: ln.Addr().Network(),
						Address: ln.Addr().String(),
					},
					Multi:     0,
					MultiIdle: 0,
					Plugin:    "",
				}},
				Proxy: []*ProxyConfig{{
					Enable: true,
					Node:   nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr5,
					},
					OutNetwork: &NetworkConfig{
						Network: ln.Addr().Network(),
						Address: ln.Addr().String(),
					},
					Plugin: "",
				}},
			}, {
				ClientConfigUnit: &ClientConfigUnit{
					Enable: true,
				},
				Listen: nil,
				Dial: []*DialConfig{{
					Enable:      true,
					Node:        nil,
					Link:        "test1",
					PP2PNetwork: "",
					SwitchUP2P:  false,
					SwitchTP2P:  false,
					ForceP2P:    false,
					Auth:        nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr6,
					},
					OutNetwork: nil,
					Multi:      0,
					MultiIdle:  0,
					Plugin:     "socks",
				}},
				Proxy: []*ProxyConfig{{
					Enable: true,
					Node:   nil,
					InNetwork: &NetworkConfig{
						Network: "tcp",
						Address: addr7,
					},
					OutNetwork: nil,
					Plugin:     "http",
				}},
				Plugin: []*PluginConfig{{
					Name: "socks",
					Type: []string{PluginTypeDial, PluginTypeProxy},
					List: []map[string]any{
						{
							"name":       plugin.NameSocks,
							"v4":         true,
							"v5":         true,
							"CMDCONNECT": true,
							"S5Auth": [][2]string{
								{"user", "pass"},
							},
						},
					},
				}, {
					Name: "http",
					Type: []string{PluginTypeDial, PluginTypeProxy},
					List: []map[string]any{
						{
							"name": plugin.NameHttpProxy,
							"Auth": [][2]string{
								{"user", "pass"},
							},
						},
					},
				}}},
		},
	}
}

func TestSdkManagerServer(t *testing.T) {
	ln, err := fasttool.EchoTcpListener()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	hs, addh, err := fasttool.EchoHttp()
	if err != nil {
		t.Fatal(err)
	}
	defer hs.Close()

	tmpConfig := testCfg(ln)

	tmp := &hyaml.Config[Config]{Config: &Config{}}
	sdk, err := newSdkManager(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	_ = addh
	fmt.Println(hjson.MustMarshalStr(sdk.GetServerView()))
	s1, err := sdk.AddServer(tmpConfig.Server[0])
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(sdk.GetServerView()))
	s2, err := sdk.AddServer(tmpConfig.Server[1])
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(sdk.GetServerView()))
	err = sdk.DelServer(s1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(sdk.GetServerView()))
	fmt.Println(sdk.GetServerView()[0].Status)
	err = sdk.StopServer(s2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sdk.GetServerView()[0].Status)
	err = sdk.StartServer(s2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sdk.GetServerView()[0].Status)
}
