package endpoint

import (
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

type Client interface {
	Get(apiPath string) (*http.Response, error)
}

type client struct {
	http.RoundTripper
	host   string
	header http.Header
}

func NewAPIClient(config Config) (Client, error) {
	rt, err := rest.TransportFor(&config.Rest)
	if err != nil {
		klog.Errorf("fail create client with config: %v reason: %v", config, err)
		return &client{}, err
	}
	client := client{
		rt,
		config.Rest.Host,
		http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", (&config.Rest).BearerToken)},
		},
	}
	return client, nil
}

func (client client) Get(apiPath string) (*http.Response, error) {
	return client.RoundTrip(&http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   client.host,
			Path:   apiPath,
		},
		Header: client.header,
	})
}
