package volume

import (
	"github.com/zizon/kbasectl/pkg/endpoint"
	v1 "k8s.io/api/core/v1"
)

type CephMount struct {
	Mount

	Monitors  []string
	User      string
	TokenName string
}

func NewCephVolume(mount CephMount) Volume {
	mount.Mount.ApplyFunc = func(v v1.PersistentVolume) v1.PersistentVolume {
		v.Spec.PersistentVolumeSource.CephFS = &v1.CephFSPersistentVolumeSource{
			Monitors: mount.Monitors,
			Path:     mount.Path,
			User:     mount.User,
			SecretRef: &v1.SecretReference{
				Name:      mount.TokenName,
				Namespace: endpoint.Namespace,
			},
		}
		return v
	}
	return NewVolume(mount.Mount)
}

func NewLocalVolume(mount Mount) Volume {
	mount.ApplyFunc = func(v v1.PersistentVolume) v1.PersistentVolume {
		v.Spec.NodeAffinity = &v1.VolumeNodeAffinity{
			Required: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{{
					// eventualy all nodes are volunteer
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      "kubernetes.io/os",
						Operator: v1.NodeSelectorOpExists,
					}},
				}},
			},
		}
		v.Spec.PersistentVolumeSource.Local = &v1.LocalVolumeSource{
			Path: mount.Path,
		}
		return v
	}

	return NewVolume(mount)
}

func NewMemoryVolume(mount Mount) Volume {
	return NewVolume(mount)
}
