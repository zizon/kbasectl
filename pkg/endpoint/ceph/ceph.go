package ceph

import "github.com/zizon/kbasectl/pkg/endpoint"

type CephEndpoint interface {
}

type cephEndpoint struct {
	client endpoint.Client
}

type Mount struct {
}

func (ceph cephEndpoint) CreateMountable(mount Mount) Volume {

	return nil
}
