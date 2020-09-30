package endpoint

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func TestClient(t *testing.T) {
	flagset := &flag.FlagSet{}
	klog.InitFlags(flagset)
	flagset.Set("v", "10")

	config, err := NewDefaultConfig()
	if err != nil {
		t.Errorf("fail to create config :%v", err)
		return
	}
	t.Logf("read config:%v", config)

	client, err := NewAPIClient(config)
	if err != nil {
		t.Errorf("fail create api client:%v", err)
		return
	}

	response, err := client.NewRequest("/api", schema.GroupVersion{Version: "v1"}).Verb("GET").Resource("namespaces").
		DoRaw(context.TODO())
	if err != nil {
		t.Errorf("fail to list namesapce:%v", err)
		return
	}

	t.Logf("response:%s", string(response))

	rt, _ := rest.TransportFor(&config.Rest)
	resp, err := rt.RoundTrip(&http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   config.Rest.Host,
			Path:   fmt.Sprintf("/api/v1/namespaces"),
		},
		Header: http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", (&config.Rest).BearerToken)},
		},
	})
	if err != nil {
		t.Errorf("fail plain request:%v", err)
		return
	}

	direct, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("fail plain response:%v", err)
		return
	}
	t.Logf("direct:%s", string(direct))
}
