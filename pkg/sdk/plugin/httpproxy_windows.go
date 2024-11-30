package plugin

import (
	"golang.org/x/sys/windows/registry"
	"net"
)

func init() {
	httpproxy_osProxySetting = httpproxy_osProxySettingWindows
}

func httpproxy_osProxySettingWindows(ln net.Listener) (func() error, error) {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	str := `localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;192.168.*`

	err = key.SetStringValue("ProxyOverride", str)
	if err != nil {
		return nil, err
	}

	err = key.SetStringValue("ProxyServer", ln.Addr().String())
	if err != nil {
		return nil, err
	}
	err = key.SetDWordValue("ProxyEnable", 1)
	if err != nil {
		return nil, err
	}
	return httpproxy_osProxySettingWindowsD, nil
}

func httpproxy_osProxySettingWindowsD() error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
	if err != nil {
		panic(err)
	}
	defer key.Close()
	err = key.SetDWordValue("ProxyEnable", 0)
	if err != nil {
		return err
	}
	return nil
}
