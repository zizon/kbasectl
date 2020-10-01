package ceph

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

func NewVolume(path string, readonly bool, capacityMb int) Volume {
	name := volumeName(path, readonly)

	capacity := resource.NewScaledQuantity(int64(capacityMb), resource.Mega)

	pv := v1.PersistentVolume{
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
	}

	pvc := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: func() []v1.PersistentVolumeAccessMode {
				if readonly {
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

	return volume{
		pv:       pv,
		capacity: *capacity,
		pvc:      pvc,
	}
}
