// Package main provides the entrypoint for the pod-image-swap-webhook
// executable.
package main

import (
	"os"

	"github.com/bonial-oss/pod-image-swap-webhook/pkg/admission"
	"github.com/bonial-oss/pod-image-swap-webhook/pkg/config"

	clientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var logger = log.Log.WithName("pod-image-swap-webhook")

func init() {
	log.SetLogger(zap.New())
}

func main() {
	configPath := os.Getenv("PISW_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	webhookConfig, err := config.Load(configPath)
	if err != nil {
		logger.Error(err, "failed to load webhook config")
		os.Exit(1)
	}

	logger.Info("loaded config", "config", webhookConfig)

	webhookServer := webhook.NewServer(webhook.Options{
		CertDir: os.Getenv("PISW_CERT_DIR"),
	})

	metricsOptions := metricsserver.Options{BindAddress: ":8080"}

	mgrOptions := manager.Options{
		HealthProbeBindAddress: ":8081",
		WebhookServer:          webhookServer,
		Metrics:                metricsOptions,
	}

	logger.Info("setting up manager")
	mgr, err := manager.New(clientconfig.GetConfigOrDie(), mgrOptions)
	if err != nil {
		logger.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	logger.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	handler := admission.NewPodImageHandler(webhookConfig, mgr.GetScheme())

	logger.Info("registering webhooks to the webhook server")
	hookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: handler})

	err = mgr.AddReadyzCheck("webhook", hookServer.StartedChecker())
	if err != nil {
		logger.Error(err, "failed to set up readiness probe")
		os.Exit(1)
	}

	err = mgr.AddHealthzCheck("webhook", hookServer.StartedChecker())
	if err != nil {
		logger.Error(err, "failed to set up liveness probe")
		os.Exit(1)
	}

	logger.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
