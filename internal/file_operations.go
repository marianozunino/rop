package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (app *App) validateFileExistence() error {
	if _, err := os.Stat(app.filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", app.filePath)
	}
	return nil
}

func (app *App) copyFileToPod(ctx context.Context, file *os.File, destPath string) error {
	command := []string{"cp", "/dev/stdin", destPath}

	req := app.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(app.pod.Name).
		Namespace(app.pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: app.containerName,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(app.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("error creating SPDY executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  file,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return fmt.Errorf("error copying file to pod: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

func (app *App) deleteFileFromPod(ctx context.Context, filePath string) error {
	cmd := []string{"rm", "-f", filePath}
	return app.runCommandInPod(ctx, cmd)
}

func (app *App) executeFile(ctx context.Context) error {
	file, err := os.Open(app.filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	isExecutable := fileInfo.Mode()&0o111 != 0
	if app.fileType == "auto" {
		app.fileType = "script"
		if isExecutable {
			app.fileType = "binary"
		}
	}

	if app.destPath == "" {
		app.destPath = "/tmp"
	}

	tempPath := fmt.Sprintf("%s/%s", app.destPath, filepath.Base(app.filePath))

	// Ensure cleanup happens
	defer func() {
		if err := app.deleteFileFromPod(ctx, tempPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to delete file %s from pod: %v\n", tempPath, err)
		} else {
			fmt.Fprintf(os.Stdout, "Deleted file %s from pod\n", tempPath)
		}
	}()

	if err := app.copyFileToPod(ctx, file, tempPath); err != nil {
		return fmt.Errorf("failed to copy file to pod: %w", err)
	}

	if app.fileType == "script" {
		return app.executeScript(ctx, tempPath)
	}
	return app.executeBinary(ctx, tempPath)
}

func (app *App) executeScript(ctx context.Context, filePath string) error {
	var command []string

	if app.runner != "" {
		command = []string{app.runner, filePath}
	} else {
		log.Debug().Msgf("No runner specified. Infering from file extension...")
		ext := filepath.Ext(filePath)
		switch ext {
		case ".js":
			log.Debug().Msgf("Using node as runner...")
			command = []string{"node", filePath}
		case ".py":
			log.Debug().Msgf("Using python as runner...")
			command = []string{"python", filePath}
		case ".rb":
			fmt.Println("Using ruby as runner...")
			command = []string{"ruby", filePath}
		case ".sh":
			fmt.Println("Using sh as runner...")
			command = []string{"sh", filePath}
		default:
			return fmt.Errorf("Wasn't able to infer runner from file extension: %s", ext)
		}
	}

	if len(app.args) > 0 {
		command = append(command, app.args...)
	}

	log.Debug().Msgf("Running command: %s\n", strings.Join(command, " "))

	return app.runCommandInPod(ctx, command)
}

func (app *App) executeBinary(ctx context.Context, filePath string) error {
	command := []string{filePath}
	if len(app.args) > 0 {
		command = append(command, app.args...)
	}
	return app.runCommandInPod(ctx, command)
}
