package main

import (
	"flag"
	"path/filepath"

	"github.com/humans-group/cimp/lib/cimp"
)

func main() {
	pathRaw := flag.String("p", "./config.yaml", "Path to config-file which should be imported")
	formatRaw := flag.String("f", "", "File format: json, yaml, edn. If empty - got from extension. Default: yaml")
	arrayValueFormatRaw := flag.String("a", "", "Array value format: json, yaml, edn. If empty - got from extension. Default: yaml")
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

	arrayValueFormat, err := cimp.NewFormat(*arrayValueFormatRaw, path)
	check(err)

	kv := cimp.NewKV("", arrayValueFormat)
	check(kv.FillFromFile(path, format))

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
