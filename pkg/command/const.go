package command

import "github.com/peakedshout/anchorage-core/pkg/sdk"

const (
	CmdCore        = sdk.CoreName
	CmdCoreVersion = sdk.CoreVersion
)

const (
	CmdHClientName    = "call-client-name"
	CmdHClientVersion = "call-client-version"
)

const (
	CmdPing = "ping"
	CmdInfo = "info"

	CmdStop   = "stop"
	CmdReload = "reload"
	CmdConfig = "config"
	CmdUpdate = "update"

	CmdViewServer           = "view_server"
	CmdViewServerById       = "view_server_id"
	CmdViewServerSession    = "view_server_session"
	CmdViewServerRoute      = "view_server_route"
	CmdViewServerLink       = "view_server_link"
	CmdViewServerSync       = "view_server_sync"
	CmdViewServerProxy      = "view_server_proxy"
	CmdViewClient           = "view_client"
	CmdViewClientUnit       = "view_client_unit"
	CmdViewClientById       = "view_client_id"
	CmdViewClientUnitById   = "view_client_unit_id"
	CmdViewClientListenById = "view_client_listen_id"
	CmdViewClientDialById   = "view_client_dial_id"
	CmdViewClientProxyById  = "view_client_proxy_id"
	CmdViewClientSession    = "view_client_session"
	CmdViewClientProxyT     = "view_client_proxyT"
	CmdViewClientProxyTUnit = "view_client_proxyT_unit"

	CmdAddServer    = "add_server"
	CmdDelServer    = "del_server"
	CmdStartServer  = "start_server"
	CmdStopServer   = "stop_server"
	CmdReloadServer = "reload_server"
	CmdUpdateServer = "update_server"
	CmdConfigServer = "config_server"

	CmdAddClient        = "add_client"
	CmdAddClientUnit    = "add_client_unit"
	CmdDelClient        = "del_client"
	CmdStartClient      = "start_client"
	CmdStartClientUnit  = "start_client_unit"
	CmdStopClient       = "stop_client"
	CmdReloadClient     = "reload_client"
	CmdReloadClientUnit = "reload_client_unit"
	CmdUpdateClient     = "update_client"
	CmdUpdateClientUnit = "update_client_unit"
	CmdConfigClient     = "config_client"
	CmdConfigClientUnit = "config_client_unit"

	CmdAddProxy    = "add_proxy"
	CmdDelProxy    = "del_proxy"
	CmdStartProxy  = "start_proxy"
	CmdStopProxy   = "stop_proxy"
	CmdReloadProxy = "reload_proxy"
	CmdUpdateProxy = "update_proxy"
	CmdConfigProxy = "config_proxy"

	CmdAddListen    = "add_listen"
	CmdDelListen    = "del_listen"
	CmdStartListen  = "start_listen"
	CmdStopListen   = "stop_listen"
	CmdReloadListen = "reload_listen"
	CmdUpdateListen = "update_listen"
	CmdConfigListen = "config_listen"

	CmdAddDial    = "add_dial"
	CmdDelDial    = "del_dial"
	CmdStartDial  = "start_dial"
	CmdStopDial   = "stop_dial"
	CmdReloadDial = "reload_dial"
	CmdUpdateDial = "update_dial"
	CmdConfigDial = "config_dial"

	CmdGetPluginUnit  = "get_plugin_unit"
	CmdListPluginUnit = "list_plugin_unit"
	CmdAddPlugin      = "add_plugin"
	CmdDelPlugin      = "del_plugin"
	CmdUpdatePlugin   = "update_plugin"
	CmdConfigPlugin   = "config_plugin"

	CmdLog = "logger"
)

var FlagList = []string{
	CmdPing, CmdInfo,
	CmdStop, CmdReload, CmdConfig, CmdUpdate,
	CmdViewServer, CmdViewServerById, CmdViewServerSession, CmdViewServerRoute, CmdViewServerLink, CmdViewServerSync, CmdViewServerProxy,
	CmdViewClient, CmdViewClientUnit, CmdViewClientById, CmdViewClientUnitById, CmdViewClientListenById, CmdViewClientDialById, CmdViewClientProxyById, CmdViewClientSession, CmdViewClientProxyT, CmdViewClientProxyTUnit,
	CmdAddServer, CmdDelServer, CmdStartServer, CmdStopServer, CmdReloadServer, CmdUpdateServer, CmdConfigServer,
	CmdAddClient, CmdAddClientUnit, CmdDelClient, CmdStartClient, CmdStartClientUnit, CmdStopClient, CmdReloadClient, CmdReloadClientUnit, CmdUpdateClient, CmdUpdateClientUnit, CmdConfigClient, CmdConfigClientUnit,
	CmdAddProxy, CmdDelProxy, CmdStartProxy, CmdStopProxy, CmdReloadProxy, CmdUpdateProxy, CmdConfigProxy,
	CmdAddListen, CmdDelListen, CmdStartListen, CmdStopListen, CmdReloadListen, CmdUpdateListen, CmdConfigListen,
	CmdAddDial, CmdDelDial, CmdStartDial, CmdStopDial, CmdReloadDial, CmdUpdateDial, CmdConfigDial,
	CmdGetPluginUnit, CmdListPluginUnit, CmdAddPlugin, CmdDelPlugin, CmdUpdatePlugin, CmdConfigPlugin,
	CmdLog,
}

type Info struct {
	Core    string   `json:"core" yaml:"core"`
	Version string   `json:"version" yaml:"version"`
	Flag    []string `json:"flag" yaml:"flag"`
}
