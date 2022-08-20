package elastic

import (
	"context"
	"fmt"

	ops "github.com/opensearch-project/opensearch-go"
)

type OpenSearchClient struct {
	client *ops.Client
}

func NewOpenSearchClient(conf ops.Config) (ESClient, error) {
	client, err := ops.NewClient(conf)
	return &OpenSearchClient{client: client}, err
}

func (osc *OpenSearchClient) Search(sReq *SearchRequest) (*SearchResponse, error) {
	resp, err := sReq.ToOpenSearchReq().Do(context.Background(), osc.client)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed searching from elastic %s", resp.String())
	}
	return NewOSResponse(resp), nil
}
