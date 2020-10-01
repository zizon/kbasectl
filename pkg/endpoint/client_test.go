package endpoint

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/zizon/kbasectl/pkg/panichain"
	"k8s.io/klog"
)

func TestClient(t *testing.T) {
	flagset := &flag.FlagSet{}
	klog.InitFlags(flagset)
	flagset.Set("v", "10")

	config := NewDefaultConfig()
	t.Logf("read config:%v", config)

	client := NewAPIClient(config)
	resp := client.Do(http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/api/v1/namespaces",
		},
	})

	direct, err := ioutil.ReadAll(resp.Body)
	panichain.Propogate(err)

	t.Logf("direct:%s", string(direct))

	defer panichain.Catch(func(err error) error {
		t.Log(err)
		return nil
	})
	panichain.Propogate(errors.New("intented"))
}
