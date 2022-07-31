package elastic

import (
	"io"

	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type SearchRequest struct {
	Body           io.Reader
	Pretty         bool
	Indexes        []string
	TrackTotalHits bool
}

func (sq *SearchRequest) ToOpenSearchReq() *opensearchapi.SearchRequest {
	r := &opensearchapi.SearchRequest{
		Body:        sq.Body,
		Pretty:      sq.Pretty,
		Index:       sq.Indexes,
		TrackScores: &sq.TrackTotalHits,
	}
	return r
}

type SearchResponse struct {
	RawResponse *opensearchapi.Response
}

func NewOSResponse(res *opensearchapi.Response) *SearchResponse {
	return &SearchResponse{}
}

type ESClient interface {
	Search(sReq *SearchRequest) (*SearchResponse, error)
}
