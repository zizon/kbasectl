package container

import (
	"testing"

	"github.com/zizon/kbasectl/pkg/endpoint"
)

func TestContaienr(t *testing.T) {
	endpoint.SetLogLevel(10)
	config := endpoint.NewDefaultConfig()
	client := endpoint.NewAPIClient(config)

	pod := NewPod(PodConfig{
		Namespace: endpoint.Namespace,
		Container: ContainerConfig{
			Name:  "test-shell",
			Image: "golang:1.15.2",
		},
	})

	endpoint.MaybeUpdate(client, func(current endpoint.NamesapcedObject) endpoint.NamesapcedObject {
		// fill infos
		current.API = "api"
		current.Version = "v1"
		current.Kind = "pods"
		current.Name = pod.Name
		current.Namespace = pod.Namespace
		if current.Object == nil {
			current.Object = pod
		}

		return current
	})
}
