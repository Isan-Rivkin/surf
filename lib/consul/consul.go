package consul

import (
	c "github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	client *c.Client
}

func (client *ConsulClient) List(prefix string) (c.KVPairs, error) {
	kv := client.client.KV()
	query := c.QueryOptions{}
	pairs, _, err := kv.List(prefix, &query)
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (client *ConsulClient) GetConsulAddr() string {
	return c.DefaultConfig().Address
}

func NewClient(address string) *ConsulClient {
	config := c.Config{
		Address: address,
	}
	client, _ := c.NewClient(&config)
	return &ConsulClient{
		client: client,
	}
}
