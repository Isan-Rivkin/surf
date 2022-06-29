package consul

import (
	"fmt"
	"strings"

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
	if client.config.Datacenter == "" {
		return c.DefaultConfig().Address
	}
	return fmt.Sprintf("%s://consul.service.%s.consul:8500", client.config.Scheme, client.config.Datacenter)
}

func (client *ConsulClient) GetCurrentDatacenter() (string, error) {
	// if previously set
	if client.config.Datacenter != "" {
		return client.config.Datacenter, nil
	}
	// fetch from agent info
	info, err := client.client.Agent().Self()
	if err != nil {
		return "", nil
	}

	conf, exist := info["Config"]

	if !exist {
		return "", fmt.Errorf("GetCurrentDatacenter - no key Config in response %v", info)
	}

	dc, ok := conf["Datacenter"].(string)

	if !ok {
		return "", fmt.Errorf("GetCurrentDatacenter - no key Datacenter in response %v", conf["Datacenter"])
	}

	return dc, nil
}

func (client *ConsulClient) ListDatacenters() ([]string, error) {
	return client.client.Catalog().Datacenters()
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

func GenerateWebURL(address string, datacenter string, key string) string {
	// clickable url will not work without a protocol in all terminals
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = fmt.Sprintf("http://%s", address)

	}
	return fmt.Sprintf("%s/ui/%s/kv/%s/edit", address, datacenter, key)
}
