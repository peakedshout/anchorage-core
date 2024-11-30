package internal

import (
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var viperMap tmap.SyncMap[*cobra.Command, *viper.Viper]

func GetViper(cmd *cobra.Command) *viper.Viper {
	value, ok := viperMap.Load(cmd)
	if !ok {
		value = viper.New()
	}
	return value
}

func NewViper(cmd *cobra.Command) *viper.Viper {
	value := viper.New()
	viperMap.Store(cmd, value)
	return value
}

func BindViper(cmd *cobra.Command) *viper.Viper {
	value, ok := viperMap.Load(cmd)
	if !ok {
		value = viper.New()
		viperMap.Store(cmd, value)
	}
	return value
}
