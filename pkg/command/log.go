package command

import (
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
)

func (c *Cmd) getLogger(ctx *xhttp.Context) error {
	r, err := c._sdk.GetLogger(ctx)
	if err != nil {
		return err
	}

	b := make([]byte, 4*1024)
	var n int
	for {
		n, err = r.Read(b)
		if err != nil {
			return err
		}
		_, err = ctx.WriteFlush(b[:n])
		if err != nil {
			return err
		}
	}
}
