package elastic

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	osc "github.com/opensearch-project/opensearch-go"
	osct "github.com/opensearch-project/opensearch-go/opensearchtransport"
)

type ConfigBuilder struct {
	headers      http.Header
	trans        http.RoundTripper
	username     string
	password     string
	urls         []string
	customLogger osct.Logger
}

func NewConf() *ConfigBuilder {
	return &ConfigBuilder{}
}

func (cb *ConfigBuilder) WithTransport(ts http.RoundTripper) {
	cb.trans = ts
}

func (cb *ConfigBuilder) WithBasicAuthToken(token string) *ConfigBuilder {
	cb.WithHeader("Authorization", fmt.Sprintf("Basic %s", token))
	return cb
}

func (cb *ConfigBuilder) WithHeader(k, v string) *ConfigBuilder {
	if cb.headers == nil {
		cb.headers = http.Header{}
	}
	cb.headers[k] = append(cb.headers[k], v)
	return cb
}
func (cb *ConfigBuilder) WithUserAuth(username, password string) *ConfigBuilder {
	cb.username = username
	cb.password = password
	return cb
}
func (cb *ConfigBuilder) WithURL(url string) *ConfigBuilder {
	cb.urls = append(cb.urls, url)
	return cb
}

func (cb *ConfigBuilder) WithColoredLogger() *ConfigBuilder {
	cb.customLogger = &osct.ColorLogger{
		Output:             os.Stdout,
		EnableRequestBody:  true,
		EnableResponseBody: true,
	}
	return cb
}

func (cb *ConfigBuilder) WithCurlLogger() *ConfigBuilder {
	cb.customLogger = &osct.CurlLogger{
		Output:             os.Stdout,
		EnableRequestBody:  true,
		EnableResponseBody: true,
	}
	return cb
}

func (cb *ConfigBuilder) Build() *osc.Config {
	conf := &osc.Config{}

	// set transport
	if cb.trans == nil {
		cb.trans = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // For testing only. Use certificate for validation.
		}
	}
	conf.Transport = cb.trans

	// set url
	conf.Addresses = append(conf.Addresses, cb.urls...)

	// set auth - if empty its just default
	conf.Username = cb.username
	conf.Password = cb.password

	// set headers
	conf.Header = cb.headers

	// set logger
	if cb.customLogger != nil {
		conf.Logger = cb.customLogger
	}

	return conf
}
