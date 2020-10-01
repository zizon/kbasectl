package container

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodConfig struct {
	Container ContainerConfig
	Namespace string
	EgressMb  int
	IngressMb int
	RunAS     int
	Labels    map[string]string
}

func NewPod(config PodConfig) v1.Pod {
	uid := int64(config.RunAS)
	if config.Labels == nil {
		config.Labels = map[string]string{}
	}
	config.Labels["kbase-pod"] = config.Container.Name

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Container.Name,
			Namespace: config.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/egress-bandwidth":  bandwidth(config.EgressMb).String(),
				"kubernetes.io/ingress-bandwidth": bandwidth(config.IngressMb).String(),
			},
			Labels: config.Labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				NewContaienr(config.Container),
			},
			HostNetwork: true,
			SecurityContext: &v1.PodSecurityContext{
				RunAsUser:  &uid,
				RunAsGroup: &uid,
			},
		},
	}
}

func bandwidth(quantity int) *resource.Quantity {
	if quantity <= 0 {
		return resource.NewScaledQuantity(100, resource.Mega)
	}

	return resource.NewScaledQuantity(int64(quantity), resource.Mega)
}
