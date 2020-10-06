package container

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	BuiltinEnvContainerID   = "CONTAINER_ID"
	BuiltinEnvNamespace     = "CONTAINER_NAMESPACE"
	BuiltinEnvLabels        = "CONTAINER_LABELS"
	BuiltinEnvAnnotations   = "CONTAINER_ANNOTATIONS"
	BuiltinEnvConfigMapPath = "CONTAINER_CONFIGMAP"
	BuiltinEnvWorkDir       = "CONTAINER_WORK"
	BuiltinEnvHostIP        = "CONTAINER_HOST_IP"
	BuiltinEnvContainerIP   = "CONTAINER_IP"
)

type ContainerConfig struct {
	Name     string
	Image    string
	CPU      int
	MemoryMb int

	Entrypoint string
	WorkDir    string
}

func NewContaienr(config ContainerConfig) v1.Container {
	return v1.Container{
		Name:  config.Name,
		Image: config.Image,
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(int64(config.CPU), ""),
				v1.ResourceMemory: *resource.NewScaledQuantity(int64(config.MemoryMb), resource.Mega),
			},
		},
		Env:        builtinEnvs(config),
		WorkingDir: config.WorkDir,
		Command:    strings.Split(config.Entrypoint, " "),
	}
}

func builtinEnvs(config ContainerConfig) []v1.EnvVar {
	envs := []v1.EnvVar{
		{
			Name: BuiltinEnvContainerID,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: BuiltinEnvNamespace,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		/*
			{
				Name: BuiltinEnvAnnotations,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.annotations",
					},
				},
			},
			{
				Name: BuiltinEnvLabels,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "metadata.labels",
					},
				},
			},
		*/
		{
			Name: BuiltinEnvHostIP,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		},
		{
			Name: BuiltinEnvContainerIP,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	}

	return envs
}
