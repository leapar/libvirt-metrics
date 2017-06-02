package config

import (
        "../libvirt"
        "../backend"
)

// Configuration
type Configuration struct {
	Kvms []*libvirt.KVM
	Interval int
        Backend  backend.Backend
}
