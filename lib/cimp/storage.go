package cimp

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type Config struct {
	Address string
}

type ConsulStorage struct {
	client *api.Client
}

func InitStorage(cfg Config) (*ConsulStorage, error) {
	clientCfg := api.DefaultConfig()
	clientCfg.Address = cfg.Address

	client, err := api.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	return &ConsulStorage{
		client: client,
	}, nil
}

func (cs *ConsulStorage) Save(kv KV) error {
	var ops api.TxnOps
	for k, v := range kv {
		op := &api.TxnOp{
			KV: &api.KVTxnOp{
				Verb:  api.KVSet,
				Key:   k,
				Value: []byte(fmt.Sprintf("%v", v)),
			},
		}
		ops = append(ops, op)
	}

	ok, _, _, err := cs.client.Txn().Txn(ops, nil)
	if !ok {
		return fmt.Errorf("execute consul SET-transaction: %w", err)
	}

	return nil
}

func (cs *ConsulStorage) List(prefix string) (KV, error) {
	pairs, _, err := cs.client.KV().List(prefix, nil)
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

func (cs *ConsulStorage) Delete(kv KV) error {
	for k := range kv {
		_, err := cs.client.KV().Delete(k, nil)
		if err != nil {
			return fmt.Errorf("delete %q from consul: %w", k, err)
		}
	}

	return nil
}
