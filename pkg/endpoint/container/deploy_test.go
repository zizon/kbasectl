package container

import (
	"testing"

	"github.com/zizon/kbasectl/pkg/endpoint"
)

func TestDeploy(t *testing.T) {
	endpoint.SetLogLevel(10)
	config := endpoint.NewDefaultConfig()
	client := endpoint.NewAPIClient(config)

	deployment := NewDeployment(PodConfig{
		Namespace: endpoint.Namespace,
		Container: ContainerConfig{
			Name:  "test-shell",
			Image: "golang:1.15.2",
		},
	}, 2)

	Deploy(client, deployment)
}
