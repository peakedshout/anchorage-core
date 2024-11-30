package internal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
	"testing"
	"time"
)

func TestBindArgs(t *testing.T) {
	cmd := &cobra.Command{
		Use: "",
	}
	type testStruct struct {
		A string        `Barg:"xa,a" Harg:"testStruct A string"`
		B int           `Barg:"xb,b" Harg:"testStruct B int"`
		C float64       `Barg:"xc,c" Harg:"testStruct C float64"`
		D []byte        `Barg:"xd,d" Harg:"testStruct D []byte"`
		F time.Duration `Barg:"xf,f" Harg:"testStruct F time.Duration"`
	}
	ts := testStruct{}
	BindArgs(cmd, &ts)
	cmd.Usage()
	bs := []byte{12, 45, 66, 33}
	b64 := base64.StdEncoding.EncodeToString(bs)
	err := cmd.Flags().Parse([]string{"-a", "hhhh", "-b", "666", "-c", "0.2123", "-d", b64, "-f", "1000ms"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(bytes.Equal(bs, ts.D))
	fmt.Printf("%#v\n", ts)
}

func TestBindArgs2(t *testing.T) {
	cmd := &cobra.Command{
		Use: "",
	}
	type testStruct struct {
		A string  `Barg:"xa,a" Harg:"testStruct A string" Garg:"g"`
		B int     `Barg:"xb,b" Harg:"testStruct B int" Garg:"g" Marg:"m"`
		C float64 `Barg:"xc,c" Harg:"testStruct C float64" Marg:"m" Oarg:"o"`
		D bool    `Barg:"xd,d" Harg:"testStruct D bool" Oarg:"o"`
	}
	ts := testStruct{}
	BindArgs(cmd, &ts)
	cmd.Usage()
}

func TestBindKey(t *testing.T) {
	cmd := &cobra.Command{
		Use: "",
	}
	type testStruct struct {
		A string  `Barg:"xa,a" Harg:"testStruct A string" Garg:"g"`
		B int     `Barg:"xb,b" Harg:"testStruct B int" Garg:"g" Marg:"m"`
		C float64 `Barg:"xc,c" Harg:"testStruct C float64" Marg:"m" Oarg:"o"`
		D bool    `Barg:"xd,d" Harg:"testStruct D bool" Oarg:"o"`
	}
	ts := testStruct{}
	BindKey(cmd, "tk", &ts)
	cmd.Usage()
	keyT := GetKeyT[testStruct](cmd, "tk")
	if keyT == nil || keyT != &ts {
		t.Fatal()
	}
}
