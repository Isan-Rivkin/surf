package elastic

import (
	"crypto/tls"

	"net/http"

	log "github.com/sirupsen/logrus"
)

type ActiveModifierTransport interface {
	RoundTrip(req *http.Request) (*http.Response, error)
}

type LogzioTransport struct {
	infoPath     string
	searchPath   string
	accountIds   []string
	underlyingTs http.RoundTripper
}

func NewLogzioTransport(infoPath, searchPath string, accountIds []string, underlyingTs http.RoundTripper) ActiveModifierTransport {
	return &LogzioTransport{
		infoPath:     infoPath,
		searchPath:   searchPath,
		accountIds:   accountIds,
		underlyingTs: underlyingTs,
	}
}

func (t *LogzioTransport) isSearchRequest(req *http.Request) bool {
	return req.Method == "POST"
}

func (t *LogzioTransport) isAppendESHEader() bool {
	return false
}

func (t *LogzioTransport) getUnderlyingTransport() http.RoundTripper {
	if t.underlyingTs != nil {
		return t.underlyingTs
	}
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}
func (t *LogzioTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	httpTransport := t.getUnderlyingTransport()

	req.URL.Path = t.infoPath
	if t.isSearchRequest(req) {
		req.URL.Path = t.searchPath
		if len(t.accountIds) > 0 {
			q := req.URL.Query()
			for _, accId := range t.accountIds {
				q.Add("accountIds", accId)
			}
			req.URL.RawQuery = q.Encode()
			q = req.URL.Query()
			log.Debugf("query params %v", q)
		}
	}

	res, err := httpTransport.RoundTrip(req)

	if t.isAppendESHEader() {
		res.Header.Set("X-Elastic-Product", "Elasticsearch")
	}

	return res, err
}
