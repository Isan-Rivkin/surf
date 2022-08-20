package elastic

import (
	"io"

	"github.com/Jeffail/gabs/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type SearchRequest struct {
	Body           io.Reader
	Pretty         bool
	Indexes        []string
	TrackTotalHits bool
}

func NewSearchRequest(body io.Reader, indexes []string, trackTotalHits bool) *SearchRequest {
	return &SearchRequest{
		Pretty:         true,
		Body:           body,
		Indexes:        indexes,
		TrackTotalHits: trackTotalHits,
	}
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
	Container   *gabs.Container
}

func NewOSResponse(res *opensearchapi.Response) *SearchResponse {
	return &SearchResponse{
		RawResponse: res,
	}
}

type ESClient interface {
	Search(sReq *SearchRequest) (*SearchResponse, error)
}
