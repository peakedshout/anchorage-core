package internal

import (
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	bindTag  = "Barg" // bing name and short
	helpTag  = "Harg" // help notes
	groupTag = "Garg" // group , one must be group all
	onlyTag  = "Oarg" // only , <= 1
	mustTag  = "Marg" // must , >= 1
)

var gMM tmap.SyncMap[*cobra.Command, *sync.Map]

// GetKeyT if you BindKey, will not return nil; if BindKey value is not *T will be panic
func GetKeyT[T any](cmd *cobra.Command, key string) *T {
	value, ok := gMM.Load(cmd)
	if !ok {
		return nil
	}
	load, _ := tmap.Load[string, *T](value, key)
	return load
}

// GetKey if you BindKey, will not return nil
func GetKey(cmd *cobra.Command, key string) any {
	value, ok := gMM.Load(cmd)
	if !ok {
		return nil
	}
	load, o := value.Load(key)
	if !o {
		return nil
	}
	return load
}

// BindKey a must be struct pointer and not nil; key cant nil
func BindKey(cmd *cobra.Command, key string, a any) {
	if key == "" {
		panic("nil key")
	}
	value, ok := gMM.Load(cmd)
	if !ok {
		value = &sync.Map{}
		gMM.Store(cmd, value)
	}
	BindArgs(cmd, a)
	value.Store(key, a)
}

// BindArgs a must be struct pointer and not nil
func BindArgs(cmd *cobra.Command, a any) {
	valueOf := reflect.ValueOf(a)
	typeOf := reflect.TypeOf(a)
	if valueOf.Kind() != reflect.Pointer {
		panic(fmt.Sprintf("invaild type: %T", a))
	}
	for valueOf.Kind() == reflect.Pointer {
		if valueOf.IsNil() {
			panic(fmt.Sprintf("nil pointer: %T", a))
		}
		valueOf = valueOf.Elem()
		typeOf = typeOf.Elem()
	}
	if valueOf.Kind() != reflect.Struct {
		panic(fmt.Sprintf("invaild type: %T", a))
	}
	gMap := make(map[string][]string)
	oMap := make(map[string][]string)
	mMap := make(map[string][]string)
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)
		fieldv := valueOf.Field(i)
		value := field.Tag.Get(bindTag)
		if value == "" {
			continue
		}
		help := field.Tag.Get(helpTag)
		split := strings.Split(value, ",")
		name, sname := "", ""
		if len(split) == 2 {
			name, sname = split[0], split[1]
		} else {
			name = split[0]
		}
		fa := fieldv.Addr().Interface()
		err := bindCmd(cmd, fa, name, sname, help)
		if err != nil {
			panic(err)
		}
		setAction(field.Tag.Get(groupTag), name, gMap)
		setAction(field.Tag.Get(onlyTag), name, oMap)
		setAction(field.Tag.Get(mustTag), name, mMap)
	}
	glist := expandAction(gMap)
	for _, i := range glist {
		cmd.MarkFlagsRequiredTogether(i...)
	}
	olist := expandAction(oMap)
	for _, i := range olist {
		cmd.MarkFlagsMutuallyExclusive(i...)
	}
	mlist := expandAction(mMap)
	for _, i := range mlist {
		cmd.MarkFlagsOneRequired(i...)
	}
}

func bindCmd(cmd *cobra.Command, a any, name, sname, help string) error {
	switch a.(type) {
	case *string:
		d := a.(*string)
		cmd.Flags().StringVarP(d, name, sname, *d, help)
	case *[]string:
		d := a.(*[]string)
		cmd.Flags().StringSliceVarP(d, name, sname, *d, help)
	case *bool:
		d := a.(*bool)
		cmd.Flags().BoolVarP(d, name, sname, *d, help)
	case *[]bool:
		d := a.(*[]bool)
		cmd.Flags().BoolSliceVarP(d, name, sname, *d, help)
	case *int:
		d := a.(*int)
		cmd.Flags().IntVarP(d, name, sname, *d, help)
	case *[]int:
		d := a.(*[]int)
		cmd.Flags().IntSliceVarP(d, name, sname, *d, help)
	case *int8:
		d := a.(*int8)
		cmd.Flags().Int8VarP(d, name, sname, *d, help)
	case *int16:
		d := a.(*int16)
		cmd.Flags().Int16VarP(d, name, sname, *d, help)
	case *int32:
		d := a.(*int32)
		cmd.Flags().Int32VarP(d, name, sname, *d, help)
	case *[]int32:
		d := a.(*[]int32)
		cmd.Flags().Int32SliceVarP(d, name, sname, *d, help)
	case *int64:
		d := a.(*int64)
		cmd.Flags().Int64VarP(d, name, sname, *d, help)
	case *[]int64:
		d := a.(*[]int64)
		cmd.Flags().Int64SliceVarP(d, name, sname, *d, help)
	case *uint:
		d := a.(*uint)
		cmd.Flags().UintVarP(d, name, sname, *d, help)
	case *[]uint:
		d := a.(*[]uint)
		cmd.Flags().UintSliceVarP(d, name, sname, *d, help)
	case *uint8:
		d := a.(*uint8)
		cmd.Flags().Uint8VarP(d, name, sname, *d, help)
	case *uint16:
		d := a.(*uint16)
		cmd.Flags().Uint16VarP(d, name, sname, *d, help)
	case *uint32:
		d := a.(*uint32)
		cmd.Flags().Uint32VarP(d, name, sname, *d, help)
	case *uint64:
		d := a.(*uint64)
		cmd.Flags().Uint64VarP(d, name, sname, *d, help)
	case *float32:
		d := a.(*float32)
		cmd.Flags().Float32VarP(d, name, sname, *d, help)
	case *[]float32:
		d := a.(*[]float32)
		cmd.Flags().Float32SliceVarP(d, name, sname, *d, help)
	case *float64:
		d := a.(*float64)
		cmd.Flags().Float64VarP(d, name, sname, *d, help)
	case *[]float64:
		d := a.(*[]float64)
		cmd.Flags().Float64SliceVarP(d, name, sname, *d, help)
	case *time.Duration:
		d := a.(*time.Duration)
		cmd.Flags().DurationVarP(d, name, sname, *d, help)
	case *[]time.Duration:
		d := a.(*[]time.Duration)
		cmd.Flags().DurationSliceVarP(d, name, sname, *d, help)
	case *net.IP:
		d := a.(*net.IP)
		cmd.Flags().IPVarP(d, name, sname, *d, help)
	case *[]net.IP:
		d := a.(*[]net.IP)
		cmd.Flags().IPSliceVarP(d, name, sname, *d, help)
	case *net.IPNet:
		d := a.(*net.IPNet)
		cmd.Flags().IPNetVarP(d, name, sname, *d, help)
	case *net.IPMask:
		d := a.(*net.IPMask)
		cmd.Flags().IPMaskVarP(d, name, sname, *d, help)
	case *[]byte:
		d := a.(*[]byte)
		cmd.Flags().BytesBase64VarP(d, name, sname, *d, help)
	case *map[string]string:
		d := a.(*map[string]string)
		cmd.Flags().StringToStringVarP(d, name, sname, *d, help)
	case *map[string]int:
		d := a.(*map[string]int)
		cmd.Flags().StringToIntVarP(d, name, sname, *d, help)
	case *map[string]int64:
		d := a.(*map[string]int64)
		cmd.Flags().StringToInt64VarP(d, name, sname, *d, help)
	case pflag.Value:
		d := a.(pflag.Value)
		cmd.Flags().VarP(d, name, sname, help)
	default:
		return fmt.Errorf("type %T not supported", a)
	}
	return nil
}

func setAction(action, name string, actionMap map[string][]string) {
	if action == "" {
		return
	}
	split := strings.Split(action, ",")
	for _, s := range split {
		actionMap[s] = append(actionMap[action], name)
	}
}

func expandAction(actionMap map[string][]string) [][]string {
	list := make([][]string, 0, len(actionMap))
	for _, i := range actionMap {
		list = append(list, i)
	}
	return list
}
