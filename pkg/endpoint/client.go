package endpoint

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/zizon/kbasectl/pkg/panichain"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var (
	flagset = &flag.FlagSet{}
)

func init() {
	klog.InitFlags(flagset)
}

type Client interface {
	Do(request http.Request) http.Response
	Config() Config
}

type client struct {
	http.RoundTripper
	host   string
	header http.Header
	config Config
}

func NewRequestBody(obj interface{}) io.ReadCloser {
	raw, err := json.Marshal(obj)
	panichain.Propogate(err)

	if klog.V(5) {
		klog.Infof("reqeust body:%s", string(raw))
	}

	return ioutil.NopCloser(bytes.NewBuffer(raw))
}

func NewAPIClient(config Config) Client {
	rt, err := rest.TransportFor(&config.Rest)
	panichain.Propogate(err)

	client := client{
		rt,
		config.Rest.Host,
		http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", (&config.Rest).BearerToken)},
		},
		config,
	}
	return client
}

func SetLogLevel(level int) {
	flagset.Set("v", fmt.Sprintf("%d", level))
}

func (client client) Do(request http.Request) http.Response {
	request.URL.Scheme = "https"
	request.URL.Host = client.host

	if request.Header == nil {
		request.Header = make(http.Header)
	}
	for key, value := range client.header {
		provided, exists := request.Header[key]
		if exists {
			provided = append(provided, value...)
		} else {
			provided = value
		}

		request.Header[key] = provided
	}
	resp, err := client.RoundTrip(&request)
	panichain.Propogate(err)
	return *resp
}

func (client client) Config() Config {
	return client.config
}
