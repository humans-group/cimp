package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"

	"github.com/humans-group/cimp/lib/cimp"
	"github.com/humans-group/cimp/lib/tree"
)

func main() {
	pathRaw := flag.String("p", "./config.yaml", "Path to config-file which should be imported")
	formatRaw := flag.String("f", "", "File format: json, yaml, edn. If empty - got from extension. Default: yaml")
	consulEndpoint := flag.String("c", "127.0.0.1:8500", "Consul endpoint in format `address:port`")
	prefixRaw := flag.String("pref", "", "Prefix for all keys")

	flag.Parse()
	if pathRaw == nil || formatRaw == nil || consulEndpoint == nil || prefixRaw == nil {
		panic("Impossible! Flags with defaults can't be nil")
	}

	path, err := filepath.Abs(*pathRaw)
	check(err)

	format, err := cimp.NewFormat(*formatRaw, path)
	check(err)

	cfgRaw, err := ioutil.ReadFile(path)
	check(err)

	kv := cimp.NewKV(tree.New())
	unmarshaler := cimp.NewUnmarshaler(kv, format)
	err = unmarshaler.Unmarshal(cfgRaw)
	check(err)

	kv.AddPrefix(*prefixRaw)

	storage, err := cimp.NewStorage(cimp.Config{Address: *consulEndpoint})
	check(err)

	check(storage.Save(kv))
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}
