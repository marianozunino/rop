package app

import (
	"context"
	"fmt"

	"github.com/marianozunino/rop/internal/k8s"
	"github.com/marianozunino/rop/internal/ui"
	"github.com/rs/zerolog/log"
)

func (app *App) Run(ctx context.Context) error {
	if err := app.initialize(); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	if err := app.preparePodExecution(ctx); err != nil {
		return fmt.Errorf("pod preparation failed: %w", err)
	}

	if err := app.executeFile(ctx); err != nil {
		return fmt.Errorf("file execution failed: %w", err)
	}

	return nil
}

func (app *App) initialize() error {
	if err := app.validateInputFile(); err != nil {
		return fmt.Errorf("input file validation failed: %w", err)
	}

	client, err := k8s.NewClient(app.kubeContext, app.namespace)
	if err != nil {
		return fmt.Errorf("failed to create K8s client: %w", err)
	}
	app.client = client

	return nil
}

func (app *App) preparePodExecution(ctx context.Context) error {
	pod, err := app.client.FindPodByName(ctx, app.podName)
	if err != nil {
		return fmt.Errorf("failed to find pod: %w", err)
	}
	app.pod = pod
	log.Debug().Msgf("Found pod: %s", pod.Name)

	if err := app.PreparePodEnvironment(); err != nil {
		return fmt.Errorf("failed to prepare pod environment: %w", err)
	}

	if !app.noConfirm {
		if err := ui.ConfirmAction(app.filePath, pod.Name, app.container); err != nil {
			return fmt.Errorf("action not confirmed: %w", err)
		}
	}

	return nil
}

