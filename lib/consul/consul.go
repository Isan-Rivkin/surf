package consul

import (
	c "github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	_client *c.Client
}

func (client *ConsulClient) List(prefix string) (c.KVPairs, error) {
	kv := client._client.KV()
	query := c.QueryOptions{}
	pairs, _, err := kv.List(prefix, &query)
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (client *ConsulClient) GetVaultAddr() string {
	return c.DefaultConfig().Address
}

func NewClient(address string) *ConsulClient {
	config := c.Config{
		Address: address,
	}
	client, _ := c.NewClient(&config)
	return &ConsulClient{
		_client: client,
	}
}
