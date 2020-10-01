package container

import (
	"github.com/zizon/kbasectl/pkg/endpoint"
	v1 "k8s.io/api/apps/v1"
)

func Deploy(client endpoint.Client, deployment v1.Deployment) {
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

		return current
	})
}
