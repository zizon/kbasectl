package endpoint

import (
	"fmt"
	"net/http"

	"github.com/zizon/kbasectl/pkg/panichain"
	"k8s.io/client-go/rest"
)

type Client interface {
	Do(request http.Request) http.Response
}

type client struct {
	http.RoundTripper
	host   string
	header http.Header
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
	}
	return client
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
