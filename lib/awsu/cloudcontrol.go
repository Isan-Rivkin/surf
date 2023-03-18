package awsu

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
)

type CloudControlAPI interface{}

type CloudControlClient struct {
	c *cloudcontrol.Client
}

func NewDCloudControlClient(c *cloudcontrol.Client) CloudControlAPI {
	return &CloudControlClient{
		c: c,
	}
}

func (cc *CloudControlClient) client() *cloudcontrol.Client {
	return cc.c
}
