package comm

import (
	"context"
	"errors"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/mslice"
	"github.com/peakedshout/go-pandorasbox/xrpc"
)

func NewStreamManager(nm map[string][]*NodeUnit) *StreamManager {
	var list []*NodeUnit
	for _, units := range nm {
		list = append(list, units...)
	}
	return &StreamManager{
		nm:   nm,
		list: list,
	}
}

type StreamManager struct {
	nm   map[string][]*NodeUnit
	list []*NodeUnit
}

func (sm *StreamManager) GetStream(ctx context.Context, node string, header string, data any) (*NodeUnit, xrpc.Stream, error) {
	units, err := sm.getNodeUnit(node)
	if err != nil {
		return nil, nil, err
	}
	lens := len(units)
	list := mslice.MakeRandRangeSlice(0, lens)
	for _, i := range list {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		nu := units[i]
		stream, err := nu.Stream(ctx, header)
		if err != nil {
			continue
		}
		err = stream.Send(data)
		if err != nil {
			continue
		}
		return nu, stream, nil
	}
	return nil, nil, errors.New("get stream failed")
}

func (sm *StreamManager) getNodeUnit(node string) ([]*NodeUnit, error) {
	if node == "" {
		return sm.list, nil
	}
	units, ok := sm.nm[node]
	if !ok {
		return nil, fmt.Errorf("not found node: %s", node)
	}
	return units, nil
}
