package volume

import (
	"github.com/zizon/kbasectl/pkg/endpoint"
	v1 "k8s.io/api/core/v1"
)

const (
	tokenName = "ceph-token"
)

type CephMount struct {
	Path     string
	ReadOnly bool
	Capacity int

	Monitors  []string
	User      string
	TokenName string
}

func CephVolume(mount CephMount) Volume {
	v := NewVolume(
		mount.Path,
		mount.ReadOnly,
		mount.Capacity,
	)

	pv := v.Get()
	pv.Spec.PersistentVolumeSource.CephFS = &v1.CephFSPersistentVolumeSource{
		Monitors: mount.Monitors,
		Path:     mount.Path,
		User:     mount.User,
		SecretRef: &v1.SecretReference{
			Name:      mount.TokenName,
			Namespace: endpoint.Namespace,
		},
	}

	pvc := v.Claim()

	return FromKubernetesVolume(pv, pvc)
}
