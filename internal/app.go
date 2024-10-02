package app

import (
	"context"
)

func (app *App) Run(ctx context.Context) error {
	if err := app.validateFileExistence(); err != nil {
		return err
	}

	if err := app.initializeKubernetesClient(); err != nil {
		return err
	}

	if err := app.setupPodAndContainer(ctx); err != nil {
		return err
	}

	if !app.noConfirm {
		if err := app.confirmAction(); err != nil {
			return err
		}
	}

	return app.executeFile(ctx)
}

func (app *App) setupPodAndContainer(ctx context.Context) error {
	pod, err := app.getRunningPod(ctx)
	if err != nil {
		return err
	}
	app.pod = pod

	containerName, err := app.selectOrSetContainer()
	if err != nil {
		return err
	}
	app.containerName = containerName

	return nil
}
