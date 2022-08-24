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

func (osc *OpenSearchClient) ListIndexes() (ESIndicesResponse, error) {
	res, err := osc.client.Indices.GetMapping()
	if err != nil {
		return nil, fmt.Errorf("failed api elastic requestic getting indices mapping %s", err.Error())
	}

	if res == nil || res.Body == nil {
		return nil, fmt.Errorf("indices response body is empty %s", err.Error())
	}

	result, err := NewESIndexRespObj(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed parsing indices response body %s", err.Error())
	}

	return result, nil
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
