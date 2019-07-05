package info

import (
	"github.com/caitong93/kube-extra-exporter/pkg/network"
)

type Stats struct {
	PodName string
	Namespace string
	Network *network.Stats
}