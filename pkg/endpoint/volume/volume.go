package volume

import (
	"fmt"
	"regexp"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	safeVolumeName = regexp.MustCompile("[/\\.]")
)

type Volume interface {
	Get() v1.PersistentVolume
	Claim() v1.PersistentVolumeClaim
	Capcity() resource.Quantity
}

func volumeName(path string, readonly bool) string {
	return fmt.Sprintf("ceph-volume-%s-%s", func() string {
		if readonly {
			return "read"
		}

		return "readwrite"
	}(), safeVolumeName.ReplaceAllString(path, "-"))
}

type volume struct {
	pv       v1.PersistentVolume
	pvc      v1.PersistentVolumeClaim
	capacity resource.Quantity
}

func (v volume) Get() v1.PersistentVolume {
	return v.pv
}

func (v volume) Claim() v1.PersistentVolumeClaim {
	return v.pvc
}

func (v volume) Capcity() resource.Quantity {
	return v.capacity
}

func FromKubernetesVolume(pv v1.PersistentVolume, pvc v1.PersistentVolumeClaim) Volume {
	return volume{
		pv:       pv,
		pvc:      pvc,
		capacity: *pv.Spec.Capacity.Storage(),
	}
}

type Mountable interface {
	Mount() Mount

	Apply(v1.PersistentVolume) v1.PersistentVolume

	VolumeNmae() string
}

type Mount struct {
	Type      string
	Source    string
	ReadOnly  bool
	Capacity  int
	applyFunc func(v1.PersistentVolume) v1.PersistentVolume
}

func (m Mount) Mount() Mount {
	return m
}

func (m Mount) Apply(v v1.PersistentVolume) v1.PersistentVolume {
	return m.applyFunc(v)
}

func (m Mount) VolumeNmae() string {
	return fmt.Sprintf("%s-volume-%s-%s", m.Type, func() string {
		if m.ReadOnly {
			return "read"
		}

		return "readwrite"
	}(), safeVolumeName.ReplaceAllString(m.Source, "-"))
}

func NewVolume(mountable Mountable) Volume {
	mount := mountable.Mount()
	name := mountable.VolumeNmae()

	capacity := resource.NewScaledQuantity(int64(mount.Capacity), resource.Mega)

	pv := mount.Apply(
		v1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1.PersistentVolumeSpec{
				Capacity: v1.ResourceList{
					v1.ResourceStorage: *capacity,
				},
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteMany, v1.ReadOnlyMany,
				},
				PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
			},
		},
	)

	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: func() []v1.PersistentVolumeAccessMode {
				if mount.ReadOnly {
					return []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany}
				}

				return []v1.PersistentVolumeAccessMode{v1.ReadWriteMany}
			}(),
			Resources: v1.ResourceRequirements{
				Requests: pv.Spec.Capacity,
			},
			VolumeName: pv.ObjectMeta.Name,
		},
	}

	return FromKubernetesVolume(pv, pvc)
}
