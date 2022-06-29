package consul

import (
	"fmt"
	"strings"

	c "github.com/hashicorp/consul/api"
)

type Client interface {
	List(prefix string) (c.KVPairs, error)
	GetSchemeType() string
	GetConsulAddr() string
	GetConsulUIBaseAddr() (string, error)
	GetCurrentDatacenter() (string, error)
	ListDatacenters() ([]string, error)
}

type ConsulClient struct {
	client *c.Client
	config *c.Config
}

func NewClient(address string, datacenter string) (Client, error) {
	config := c.Config{
		Address:    address,
		Datacenter: datacenter,
	}
	client, err := c.NewClient(&config)
	if err != nil {
		return nil, err
	}
	return &ConsulClient{
		client: client,
		config: &config,
	}, err
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

// http or https scheme type
func (client *ConsulClient) GetSchemeType() string {
	return client.config.Scheme
}

func (client *ConsulClient) GetConsulAddr() string {
	return client.config.Address
}

func (client *ConsulClient) GetConsulUIBaseAddr() (string, error) {
	dc, err := client.GetCurrentDatacenter()

	if err != nil {
		return "", fmt.Errorf("failed getting datacenter to construct consul ui base address - %s", err.Error())
	}

	addr := client.config.Address

	if !strings.HasPrefix(addr, client.GetSchemeType()) {
		addr = client.config.Scheme + "://" + addr
	}

	uiBaseUrl := fmt.Sprintf("%s/ui", addr)

	if dc != "" {
		uiBaseUrl = fmt.Sprintf("%s/%s", uiBaseUrl, dc)
	}

	return uiBaseUrl, nil
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

	// cache dc for future requests
	client.config.Datacenter = dc

	return dc, nil
}

func (client *ConsulClient) ListDatacenters() ([]string, error) {
	return client.client.Catalog().Datacenters()
}

func GenerateKVWebURL(uiBaseAddress, key string) string {
	return fmt.Sprintf("%s/kv/%s/edit", uiBaseAddress, key)
}
