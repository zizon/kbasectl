package volume

import (
	v1 "k8s.io/api/core/v1"
)

type CephMount struct {
	Mount

	Monitors []string
	User     string

	TokenName      string
	TokenNamespace string
}

func NewCephVolume(namespace string, mount CephMount) Volume {
	mount.Mount.applyFunc = func(v v1.PersistentVolume) v1.PersistentVolume {
		v.Spec.PersistentVolumeSource.CephFS = &v1.CephFSPersistentVolumeSource{
			Monitors: mount.Monitors,
			Path:     mount.Source,
			User:     mount.User,
			SecretRef: &v1.SecretReference{
				Name:      mount.TokenName,
				Namespace: mount.TokenNamespace,
			},
		}
		return v
	}
	return NewPersistentVolume(namespace, mount.Mount)
}

func NewLocalVolume(namespace string, mount Mount) Volume {
	mount.applyFunc = func(v v1.PersistentVolume) v1.PersistentVolume {
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
			Path: mount.Source,
		}
		return v
	}

	return NewPersistentVolume(namespace, mount)
}
