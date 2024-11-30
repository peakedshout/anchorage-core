package client

import (
	"github.com/peakedshout/go-pandorasbox/tool/bio"
	"github.com/peakedshout/go-pandorasbox/xrpc"
	"net"
	"sync"
	"time"
)

type connInfo struct {
	PublicNetwork string
	PublicAddress string
	TimeStamp     time.Time
	Cert          []byte
	Key           []byte
}

func newConn(stream xrpc.Stream, laddr, raddr net.Addr, df func()) net.Conn {
	c := &_conn{
		Stream: stream,
		br:     nil,
		laddr:  laddr,
		raddr:  raddr,
		df:     df,
	}
	c.br = bio.NewBufferReader(c)
	return c
}

type _conn struct {
	xrpc.Stream
	br           *bio.BufferReader
	laddr, raddr net.Addr
	df           func()
	closer       sync.Once
}

func (c *_conn) Read(b []byte) (n int, err error) {
	return c.br.Read(b)
}

func (c *_conn) Write(b []byte) (n int, err error) {
	for i := 0; i < len(b); i += 32 * 1024 {
		j := i + 32*1024
		if j > len(b) {
			j = len(b)
		}
		err = c.Stream.Send(b[i:j])
		if err != nil {
			return i, err
		}
	}
	return len(b), nil
}

func (c *_conn) Close() error {
	defer c.closer.Do(c.df)
	return c.Stream.Close()
}

func (c *_conn) LocalAddr() net.Addr {
	return c.laddr
}

func (c *_conn) RemoteAddr() net.Addr {
	return c.raddr
}

func (c *_conn) SetDeadline(t time.Time) error {
	return nil
}

func (c *_conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *_conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *_conn) ReadPacket() ([]byte, error) {
	var b []byte
	err := c.Stream.Recv(&b)
	return b, err
}
