package internal

import _ "embed"

//go:embed tmpl/server.yaml
var tmplServer []byte

func GetTmplServer() []byte {
	return tmplServer
}

//go:embed tmpl/client.yaml
var tmplClient []byte

func GetTmplClient() []byte {
	return tmplClient
}

//go:embed tmpl/dial.yaml
var tmplDial []byte

func GetTmplDial() []byte {
	return tmplDial
}

//go:embed tmpl/listen.yaml
var tmplListen []byte

func GetTmplListen() []byte {
	return tmplListen
}

//go:embed tmpl/plugin.yaml
var tmplPlugin []byte

func GetTmplPlugin() []byte {
	return tmplPlugin
}

//go:embed tmpl/proxy.yaml
var tmplProxy []byte

func GetTmplProxy() []byte {
	return tmplProxy
}
