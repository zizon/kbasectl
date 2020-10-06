package generator

import (
	"fmt"
	"path"

	"github.com/zizon/kbasectl/pkg/endpoint/container"
	"github.com/zizon/kbasectl/pkg/endpoint/volume"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	Namespace string
	Name      string

	Labels map[string]string
	Envs   map[string]string

	RunAS      int
	Entrypoint string
	WorkDir    string

	Image     string
	CPU       int
	MemoryMb  int
	IngressMb int
	EgressMb  int

	Replica int

	CephBinds   []CephBind
	LocalBinds  []LocalBind
	MemoryBinds []MemoryBind

	ConfigBind ConfigMapBind
}

type CephBind struct {
	Monitors []string
	User     string

	TokenName      string
	TokenNamespace string

	Source string

	ReadOnly   bool
	CapacityMb int

	MountTo  string
	SubPaths []string
}

type LocalBind struct {
	Source string

	ReadOnly   bool
	CapacityMb int

	MountTo  string
	SubPaths []string
}

type MemoryBind struct {
	CapacityMb int

	MountTo string
}

type ConfigMapBind struct {
	name      string
	MountTo   string
	ConfigMap map[string]string
}

type volumeBind struct {
	volumeName string
	source     v1.VolumeSource
	readOnly   bool
	mountTo    string
	subPath    []string
}

type volumeBinable interface {
	toVolumeBind() volumeBind
}

func (ceph CephBind) toVolumeBind() volumeBind {
	name := volume.Mount{
		Type:     CephBindVolumeType,
		Source:   ceph.Source,
		ReadOnly: ceph.ReadOnly,
	}.VolumeNmae()
	return volumeBind{
		volumeName: name,
		readOnly:   ceph.ReadOnly,
		mountTo:    ceph.MountTo,
		subPath:    ceph.SubPaths,
		source: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: name,
				ReadOnly:  ceph.ReadOnly,
			},
		},
	}
}

func (local LocalBind) toVolumeBind() volumeBind {
	name := volume.Mount{
		Type:     LocalBindVolumeType,
		Source:   local.Source,
		ReadOnly: local.ReadOnly,
	}.VolumeNmae()
	return volumeBind{
		volumeName: name,
		readOnly:   local.ReadOnly,
		mountTo:    local.MountTo,
		source: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: name,
				ReadOnly:  local.ReadOnly,
			},
		},
		subPath: local.SubPaths,
	}
}

func (memory MemoryBind) toVolumeBind() volumeBind {
	name := volume.Mount{
		Type:     MemoryBindVolumeType,
		Source:   memory.MountTo,
		ReadOnly: false,
	}.VolumeNmae()
	return volumeBind{
		volumeName: name,
		readOnly:   false,
		mountTo:    memory.MountTo,
		source: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium:    v1.StorageMediumMemory,
				SizeLimit: resource.NewScaledQuantity(int64(memory.CapacityMb), resource.Mega),
			},
		},
		subPath: []string{""},
	}
}

func (configMap ConfigMapBind) toVolumeBind() volumeBind {
	return volumeBind{
		volumeName: configMap.name,
		readOnly:   true,
		mountTo:    configMap.MountTo,
		source: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMap.name,
				},
			},
		},
	}
}

const (
	CephBindVolumeType      = "ceph"
	LocalBindVolumeType     = "local"
	MemoryBindVolumeType    = "memory"
	ConfigMapBindVolumeType = "configmap"
)

func (config Config) ConfigMapName() string {
	return fmt.Sprintf("%s-configmap", config.Name)
}

func GenerateDeployment(config Config) appv1.Deployment {
	// collect pv
	volumeBinds := []volumeBind{}

	// ceph
	for _, bindable := range config.CephBinds {
		volumeBinds = append(volumeBinds, bindable.toVolumeBind())
	}

	// local
	for _, bindable := range config.LocalBinds {
		volumeBinds = append(volumeBinds, bindable.toVolumeBind())
	}

	// memory
	for _, bindable := range config.MemoryBinds {
		volumeBinds = append(volumeBinds, bindable.toVolumeBind())
	}

	// config map
	config.ConfigBind.name = config.ConfigMapName()
	volumeBinds = append(volumeBinds, config.ConfigBind.toVolumeBind())

	podVolumes := map[string]v1.Volume{}
	mounts := map[string]v1.VolumeMount{}
	// convert volume bind to kubernetes
	for _, bind := range volumeBinds {
		podVolumes[bind.volumeName] = v1.Volume{
			Name:         bind.volumeName,
			VolumeSource: bind.source,
		}

		if len(bind.subPath) == 0 {
			bind.subPath = append(bind.subPath, "")
		}

		for _, subPath := range bind.subPath {
			lookup := fmt.Sprintf("%s-%s", bind.volumeName, subPath)
			mounts[lookup] = v1.VolumeMount{
				Name:        bind.volumeName,
				ReadOnly:    bind.readOnly,
				MountPath:   path.Join(bind.mountTo, subPath),
				SubPathExpr: subPath,
			}
		}
	}

	// create deployment template
	deployment := container.NewDeployment(container.PodConfig{
		Namespace: config.Namespace,
		IngressMb: config.IngressMb,
		EgressMb:  config.EgressMb,
		RunAS:     config.RunAS,
		Labels:    config.Labels,
		Container: container.ContainerConfig{
			Name:     config.Name,
			Image:    config.Image,
			CPU:      config.CPU,
			MemoryMb: config.MemoryMb,

			WorkDir:    config.WorkDir,
			Entrypoint: config.Entrypoint,
		},
	}, config.Replica)

	// set back volume
	slicePodVolumes := []v1.Volume{}
	for _, v := range podVolumes {
		slicePodVolumes = append(slicePodVolumes, v)
	}
	deployment.Spec.Template.Spec.Volumes = slicePodVolumes

	// set back mounts
	sliceVolumeMounts := []v1.VolumeMount{}
	for _, m := range mounts {
		sliceVolumeMounts = append(sliceVolumeMounts, m)
	}
	container := deployment.Spec.Template.Spec.Containers[0]
	container.VolumeMounts = sliceVolumeMounts

	// set env
	for key, value := range config.Envs {
		container.Env = append(container.Env, v1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	// setback
	deployment.Spec.Template.Spec.Containers = []v1.Container{container}
	annotateWithVersion(&deployment)
	return deployment
}

func GeneratePersistenVolume(config Config) []v1.PersistentVolume {
	volumes := map[string]v1.PersistentVolume{}

	// ceph
	for _, cephVolume := range config.CephBinds {
		pv := volume.NewCephVolume(config.Namespace, volume.CephMount{
			Monitors: cephVolume.Monitors,
			User:     cephVolume.User,

			TokenName:      cephVolume.TokenName,
			TokenNamespace: cephVolume.TokenNamespace,

			Mount: volume.Mount{
				Type:     CephBindVolumeType,
				Source:   cephVolume.Source,
				ReadOnly: cephVolume.ReadOnly,
				Capacity: cephVolume.CapacityMb,
			},
		}).Get()
		volumes[pv.Name] = pv
	}

	// local
	for _, localVolume := range config.LocalBinds {
		pv := volume.NewLocalVolume(config.Namespace, volume.Mount{
			Type:     LocalBindVolumeType,
			Source:   localVolume.Source,
			ReadOnly: localVolume.ReadOnly,
			Capacity: localVolume.CapacityMb,
		}).Get()
		volumes[pv.Name] = pv
	}

	sliceVolumes := []v1.PersistentVolume{}
	for _, v := range volumes {
		annotateWithVersion(&v)
		sliceVolumes = append(sliceVolumes, v)
	}
	return sliceVolumes
}

func GeneratePersistenVolumeClaim(config Config) []v1.PersistentVolumeClaim {
	volumeClaims := map[string]v1.PersistentVolumeClaim{}

	// ceph
	for _, cephVolume := range config.CephBinds {
		pvc := volume.NewCephVolume(config.Namespace, volume.CephMount{
			Monitors: cephVolume.Monitors,
			User:     cephVolume.User,

			TokenName:      cephVolume.TokenName,
			TokenNamespace: cephVolume.TokenNamespace,

			Mount: volume.Mount{
				Type:     CephBindVolumeType,
				Source:   cephVolume.Source,
				ReadOnly: cephVolume.ReadOnly,
				Capacity: cephVolume.CapacityMb,
			},
		}).Claim()
		volumeClaims[pvc.Name] = pvc
	}

	// local
	for _, localClaim := range config.LocalBinds {
		pvc := volume.NewLocalVolume(config.Namespace, volume.Mount{
			Type:     LocalBindVolumeType,
			Source:   localClaim.Source,
			ReadOnly: localClaim.ReadOnly,
			Capacity: localClaim.CapacityMb,
		}).Claim()
		volumeClaims[pvc.Name] = pvc
	}

	sliceClaims := []v1.PersistentVolumeClaim{}
	for _, v := range volumeClaims {
		annotateWithVersion(&v)
		sliceClaims = append(sliceClaims, v)
	}
	return sliceClaims
}

func GenerateNamespace(config Config) []v1.Namespace {
	namespaces := map[string]v1.Namespace{
		config.Namespace: {
			ObjectMeta: metav1.ObjectMeta{
				Name: config.Namespace,
			},
		},
	}

	for _, ceph := range config.CephBinds {
		namespaces[ceph.TokenNamespace] = v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ceph.TokenNamespace,
			},
		}
	}

	sliceNamesapces := []v1.Namespace{}
	for _, ns := range namespaces {
		annotateWithVersion(&ns)
		sliceNamesapces = append(sliceNamesapces, ns)
	}
	return sliceNamesapces
}

func GenerateConfigMap(config Config) v1.ConfigMap {
	configmap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: config.Namespace,
			Name:      config.ConfigMapName(),
		},
		Data: config.ConfigBind.ConfigMap,
	}

	annotateWithVersion(&configmap)
	return configmap
}

func GenerateSecret(config Config) []v1.Secret {
	secrets := map[string]v1.Secret{}

	for _, ceph := range config.CephBinds {
		lookup := fmt.Sprintf("%s/%s", ceph.TokenNamespace, ceph.TokenName)
		if sescret, exists := secrets[lookup]; exists {
			sescret.StringData[ceph.User] = ""
		} else {
			secrets[lookup] = v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ceph.TokenNamespace,
					Name:      ceph.TokenName,
				},
				StringData: map[string]string{
					ceph.User: "",
				},
			}
		}
	}

	sliceSecret := []v1.Secret{}
	for _, s := range secrets {
		annotateWithVersion(&s)
		sliceSecret = append(sliceSecret, s)
	}
	return sliceSecret
}
