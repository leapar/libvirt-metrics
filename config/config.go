package config

import (
        "github.com/leapar/libvirt-metrics/libvirt"
        "github.com/leapar/libvirt-metrics/backend"
)

// Configuration
type Configuration struct {
	Kvms []*libvirt.KVM
	Interval int
        Backend  backend.Backend
}
