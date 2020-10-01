package container

import (
	"fmt"

	"github.com/zizon/kbasectl/pkg/endpoint"
	"github.com/zizon/kbasectl/pkg/panichain"
	v1 "k8s.io/api/apps/v1"
)

func Deploy(client endpoint.Client, deployment v1.Deployment) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if len(container.Image) == 0 {
			panichain.Propogate(fmt.Errorf("container specified no image:%v", container))
			return
		}
	}

	endpoint.MaybeUpdate(client, func(current endpoint.NamesapcedObject) endpoint.NamesapcedObject {
		// fill infos
		current.API = "apis"
		current.Group = "apps"
		current.Version = "v1"
		current.Kind = "deployments"
		current.Name = deployment.Name
		current.Namespace = deployment.Namespace
		if current.Object == nil {
			current.Object = deployment
			return current
		}

		current.Object = deployment
		return endpoint.NamesapcedObject{
			API:       "apis",
			Group:     "apps",
			Version:   "v1",
			Kind:      "deployments",
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Object:    deployment,
		}
	})
}
