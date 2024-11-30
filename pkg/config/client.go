package config

import "errors"

type ClientConfig struct {
	Nodes []NodeConfig `json:"nodes" yaml:"nodes" comment:"dial to server node list"`
}

func (cc *ClientConfig) Check() error {
	if cc == nil {
		return errors.New("nil client config")
	}
	errs := make([]error, 0, len(cc.Nodes))
	for _, node := range cc.Nodes {
		errs = append(errs, node.Check())
	}
	return errors.Join(errs...)
}
