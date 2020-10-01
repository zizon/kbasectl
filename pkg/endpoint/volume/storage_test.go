package volume

import (
	"flag"
	"net/http"
	"net/url"
	"testing"

	"github.com/zizon/kbasectl/pkg/endpoint"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func TestCreateStorage(t *testing.T) {
	flagset := &flag.FlagSet{}
	klog.InitFlags(flagset)
	flagset.Set("v", "10")

	config := endpoint.NewDefaultConfig()
	client := endpoint.NewAPIClient(config)

	tokenName := "ceph-token"

	// 1. ensure namespace
	endpoint.MaybeCreate(client,
		endpoint.NamesapcedObject{
			API:     "api",
			Version: "v1",
			Kind:    "namespaces",
			Name:    endpoint.Namespace,
			Object: v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: endpoint.Namespace,
				},
			},
		},
	)

	// 2. ensure token
	endpoint.MaybeCreate(client,
		endpoint.NamesapcedObject{
			API:       "api",
			Version:   "v1",
			Namespace: endpoint.Namespace,
			Kind:      "secrets",
			Name:      tokenName,
			Object: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: tokenName,
				},
				StringData: map[string]string{
					config.Ceph.User: config.Ceph.Token,
				},
			},
		},
	)

	// 3. prepare volume
	mount := CephMount{
		Path:     "/template",
		ReadOnly: true,
		Capacity: 1024,

		Monitors:  config.Ceph.Monitors,
		User:      config.Ceph.User,
		TokenName: tokenName,
	}
	cephVolume := CephVolume(mount)

	// 4. create pv
	CreateStorage(client, "ceph-test", cephVolume)

	// 5. cleanup
	client.Do(http.Request{
		Method: "DELETE",
		URL: &url.URL{
			Path: "/api/v1/namespaces/ceph-test",
		},
	})
}
