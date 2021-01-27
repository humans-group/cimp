package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/humans-group/consul-importer/lib/cimp"
)

func main() {
	pathRaw := flag.String("p", "./config.yaml", "Path to yaml with config-file which should be imported")
	formatRaw := flag.String("f", "", "File format: json, yaml. If empty - got from extension. Default: yaml")
	consulEndpoint := flag.String("c", "127.0.0.1:8500", "Consul endpoint in format: `address:port`")
	flag.Parse()
	if pathRaw == nil || formatRaw == nil || consulEndpoint == nil {
		panic("Impossible! Flags with defaults can't be nil")
	}

	path, err := filepath.Abs(*pathRaw)
	check(err)

	format, err := cimp.InitFormat(*formatRaw, path)
	check(err)

	kv := cimp.NewKV()
	check(kv.FillFromFile(path, format))

	serverName, err := kv.GetString("server/name")
	check(err)

	kv.AddPrefix(fmt.Sprintf("services/%s/", serverName))

	storage, err := cimp.InitStorage(cimp.Config{Address: *consulEndpoint})
	check(err)

	check(storage.Save(kv))
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}
