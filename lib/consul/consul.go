package consul

import (
	"fmt"

	c "github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	client *c.Client
	config *c.Config
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
	return fmt.Sprintf("%s://consul.service.%s.consul:8500", client.config.Scheme, client.config.Datacenter)
}

func NewClient(address string, datacenter string) *ConsulClient {
	config := c.Config{
		Address:    address,
		Datacenter: datacenter,
	}
	client, _ := c.NewClient(&config)
	return &ConsulClient{
		client: client,
		config: &config,
	}
}

// Clickable URLs influenced of the termLink library https://github.com/savioxavier/termlink/blob/master/termlink.go#L165
func GetKeyURL(address string, datacenter string, key string) string {
	return fmt.Sprintf("\x1b]8;;%s/ui/%s/kv/%s/edit\x07%s\x1b]8;;\x07", address, datacenter, key, key)
}
