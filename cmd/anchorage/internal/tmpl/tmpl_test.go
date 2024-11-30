package tmpl

import "testing"

func errCheck(err error) {
	if err != nil {
		panic(err)
	}
}

func TestTmpl(t *testing.T) {
	errCheck(makeServerTmpl())
	errCheck(makeClientTmpl())
	errCheck(makeListenTmpl())
	errCheck(makeDialTmpl())
	errCheck(makeProxyTmpl())
	errCheck(makePluginTmpl())
}
