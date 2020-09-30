package endpoint

import (
	"flag"
	"io/ioutil"
	"testing"

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

	resp, err := client.Get("/api/v1/namespaces")
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
