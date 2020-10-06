package volume

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/zizon/kbasectl/pkg/endpoint"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateStorage(t *testing.T) {
	endpoint.SetLogLevel(10)

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
	cephVolume := NewCephVolume(
		"ceph-test",
		CephMount{
			Mount: Mount{
				Type:     "ceph",
				Source:   "/template",
				ReadOnly: true,
				Capacity: 1024,
			},

			Monitors:       config.Ceph.Monitors,
			User:           config.Ceph.User,
			TokenName:      tokenName,
			TokenNamespace: endpoint.Namespace,
		})
	CreateStorage(client, "ceph-test", cephVolume)

	// 4. local volume
	localVolume := NewLocalVolume("ceph-test", Mount{
		Type:     "local",
		Source:   "/tmp",
		ReadOnly: false,
		Capacity: 1,
	})
	CreateStorage(client, "ceph-test", localVolume)

	// 5. cleanup
	client.Do(http.Request{
		Method: "DELETE",
		URL: &url.URL{
			Path: "/api/v1/namespaces/ceph-test",
		},
	})
}
