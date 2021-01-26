package cimp

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type Config struct {
	Address string
}

type ConsulRepo struct {
	client *api.Client
}

func InitRepo(cfg Config) (*ConsulRepo, error) {
	clientCfg := api.DefaultConfig()
	clientCfg.Address = cfg.Address

	client, err := api.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	return &ConsulRepo{
		client: client,
	}, nil
}

func (repo *ConsulRepo) Save(kv KV) error {
	for k, v := range kv {
		kvPair := &api.KVPair{
			Key:   k,
			Value: []byte(fmt.Sprintf("%v", v)),
		}
		_, err := repo.client.KV().Put(kvPair, nil)
		if err != nil {
			return fmt.Errorf("put %q to consul: %w", k, err)
		}
	}

	return nil
}

func (repo *ConsulRepo) List(prefix string) (KV, error) {
	pairs, _, err := repo.client.KV().List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("list with prefix %q from consul: %w", prefix, err)
	}

	kv := NewKV()
	for k, v := range pairs {
		if v == nil {
			return nil, fmt.Errorf("pair #%d is nil", k+1)
		}
		kv.AddPair(*v)
	}

	return kv, nil
}

func (repo *ConsulRepo) Delete(kv KV) error {
	for k := range kv {
		_, err := repo.client.KV().Delete(k, nil)
		if err != nil {
			return fmt.Errorf("delete %q from consul: %w", k, err)
		}
	}

	return nil
}
