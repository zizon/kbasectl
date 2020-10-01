package endpoint

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

type Client interface {
	Do(request http.Request) (*http.Response, error)
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
		return nil, err
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

func (client client) Do(request http.Request) (*http.Response, error) {
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
	return client.RoundTrip(&request)
}
