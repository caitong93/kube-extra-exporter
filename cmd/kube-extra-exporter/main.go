package main

import (
	"context"
	"fmt"
	"os"

	"github.com/caitong93/kube-extra-exporter/pkg/apis"
	"github.com/caitong93/kube-extra-exporter/pkg/apis/filters"
	"github.com/caitong93/kube-extra-exporter/pkg/apis/modifiers"
	"github.com/caitong93/kube-extra-exporter/pkg/manager"
	"github.com/caitong93/kube-extra-exporter/pkg/metrics"
	"github.com/caitong93/kube-extra-exporter/pkg/pod"
	"github.com/caitong93/kube-extra-exporter/pkg/version"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/logger"
	metricsplugin "github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/reqlog"
	pversion "github.com/caicloud/nirvana/plugins/version"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Print nirvana banner.
	fmt.Println(nirvana.Logo, nirvana.Banner)

	// Create nirvana command.
	cmd := config.NewNamedNirvanaCommand("server", config.NewDefaultOption())

	nodeName := mustGetNodeName()
	log.Infoln("Node name", nodeName)

	restCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal(err)
	}
	podLister := pod.NewLister(context.Background(), kubernetes.NewForConfigOrDie(restCfg), nodeName)

	// Init manager and prometheus collector.
	manager, err := manager.New(podLister)
	if err != nil {
		log.Fatalln("Err create manager:", err)
	}
	go func() {
		// FIXME: graceful terminate
		if err := manager.Run(context.Background()); err != nil {
			log.Fatal("Err run manager:", err)
		}
	}()
	prometheus.MustRegister(metrics.NewPrometheusCollector(manager))

	// Create plugin options.
	metricsOption := metricsplugin.NewDefaultOption() // Metrics plugin.
	loggerOption := logger.NewDefaultOption()         // Logger plugin.
	reqlogOption := reqlog.NewDefaultOption()         // Request log plugin.
	versionOption := pversion.NewOption(              // Version plugin.
		"kube-extra-exporter",
		version.Version,
		version.Commit,
		version.Package,
	)

	// Enable plugins.
	cmd.EnablePlugin(metricsOption, loggerOption, reqlogOption, versionOption)

	// Create server config.
	serverConfig := nirvana.NewConfig()

	// Configure APIs. These configurations may be changed by plugins.
	serverConfig.Configure(
		nirvana.Logger(log.DefaultLogger()), // Will be changed by logger plugin.
		nirvana.Filter(filters.Filters()...),
		nirvana.Modifier(modifiers.Modifiers()...),
		nirvana.Descriptor(apis.Descriptor()),
	)

	// Set nirvana command hooks.
	cmd.SetHook(&config.NirvanaCommandHookFunc{
		PreServeFunc: func(config *nirvana.Config, server nirvana.Server) error {
			// Output project information.
			config.Logger().Infof("Package:%s Version:%s Commit:%s", version.Package, version.Version, version.Commit)
			return nil
		},
	})

	// Start with server config.
	if err := cmd.ExecuteWithConfig(serverConfig); err != nil {
		serverConfig.Logger().Fatal(err)
	}
}

func mustGetNodeName() string {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		log.Fatal("Node name not found")
	}

	return nodeName
}
