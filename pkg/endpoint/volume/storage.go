package volume

import (
	"github.com/zizon/kbasectl/pkg/endpoint"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func CreateStorage(client endpoint.Client, namespace string, volume Volume) {
	// 1. namespace
	endpoint.MaybeCreate(
		client,
		endpoint.NamesapcedObject{
			API:     "api",
			Version: "v1",
			Kind:    "namespaces",
			Name:    namespace,
			Object: v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			},
		},
	)

	// 2. pv
	pv := volume.Get()
	klog.Infof("ensureing pv: %v", pv)
	endpoint.MaybeCreate(client,
		endpoint.NamesapcedObject{
			API:     "api",
			Version: "v1",
			Kind:    "persistentvolumes",
			Name:    pv.Name,
			Object:  pv,
		},
	)

	// 3. pvc
	pvc := volume.Claim()
	endpoint.MaybeCreate(client,
		endpoint.NamesapcedObject{
			API:       "api",
			Version:   "v1",
			Namespace: namespace,
			Kind:      "persistentvolumeclaims",
			Name:      pvc.Name,
			Object:    pvc,
		},
	)
}
