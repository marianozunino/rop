/*
Copyright © 2024 Mariano Zunino <marianoz@posteo.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"context"
	"fmt"
	"os"

	app "github.com/marianozunino/rop/internal"
	"github.com/marianozunino/rop/internal/logger"
	"github.com/spf13/cobra"
)

type config struct {
	kubeContext   string
	filePath      string
	podName       string
	containerName string
	noConfirm     bool
	fileType      string
	fileArgs      []string
	destPath      string
	runner        string
	namespace     string
	verbose       bool
}

var logo = `
 ______     ______     ______
/\  == \   /\  __ \   /\  == \
\ \  __<   \ \ \/\ \  \ \  _-/
 \ \_\ \_\  \ \_____\  \ \_\
  \/_/ /_/   \/_____/   \/_/     ` + VersionFromBuild()

func NewRootCmd() *cobra.Command {
	cfg := &config{}

	rootCmd := &cobra.Command{
		Use:   "rop",
		Short: "Run a script or binary on a Kubernetes pod",
		Long: logo + `
Run on Pod (ROP) is a tool to execute scripts or binaries on Kubernetes pods.
It simplifies the process of running files directly in your Kubernetes environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.ConfigureLogger(cfg.verbose)
			runRop(cmd.Context(), cfg)
		},
	}

	addFlags(rootCmd, cfg)

	return rootCmd
}

func addFlags(cmd *cobra.Command, cfg *config) {
	cmd.Flags().StringVarP(&cfg.kubeContext, "context", "c", "", "Kubernetes context")
	cmd.Flags().StringVarP(&cfg.namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().StringVarP(&cfg.podName, "pod", "p", "", "The target pod name")
	cmd.Flags().StringVar(&cfg.containerName, "container", "", "The container name (optional for single-container pods)")
	cmd.Flags().StringVarP(&cfg.filePath, "file", "f", "", "The file path to execute")
	cmd.Flags().StringArrayVarP(&cfg.fileArgs, "args", "a", []string{}, "File arguments")
	cmd.Flags().StringVarP(&cfg.destPath, "dest-path", "d", "/tmp", "Destination path for the script or binary")
	cmd.Flags().StringVarP(&cfg.runner, "runner", "r", "", "Custom runner for the script (e.g., 'python', 'node')")
	cmd.Flags().StringVarP(&cfg.fileType, "type", "t", "auto", "File type: 'script', 'binary', or 'auto'")
	cmd.Flags().BoolVar(&cfg.noConfirm, "no-confirm", false, "Skip confirmation prompt")
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", false, "Verbose output")

	cmd.MarkFlagRequired("context")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("pod")

	cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "script", "binary"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.RegisterFlagCompletionFunc("context", contextCompletion)
	cmd.RegisterFlagCompletionFunc("namespace", namespaceCompletion)

	cmd.Flags().SortFlags = false
}

func contextCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	contexts, err := app.GetAvailableContexts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting contexts: %v\n", err)
		return nil, cobra.ShellCompDirectiveError
	}
	return contexts, cobra.ShellCompDirectiveNoFileComp
}

func namespaceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	context, err := cmd.Flags().GetString("context")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting context flag: %v\n", err)
		return nil, cobra.ShellCompDirectiveError
	}

	namespaces, err := app.GetAvailableNamespaces(context)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting namespaces: %v\n", err)
		return nil, cobra.ShellCompDirectiveError
	}
	return namespaces, cobra.ShellCompDirectiveNoFileComp
}

func runRop(ctx context.Context, cfg *config) {
	if err := validateConfig(cfg); err != nil {
		fmt.Errorf("invalid config: %w", err)
		os.Exit(1)
	}

	appInstance := app.NewApp(
		app.WithKubeContext(cfg.kubeContext),
		app.WithNamespace(cfg.namespace),
		app.WithFilePath(cfg.filePath),
		app.WithPodName(cfg.podName),
		app.WithContainerName(cfg.containerName),
		app.WithNoConfirm(cfg.noConfirm),
		app.WithFileType(cfg.fileType),
		app.WithArgs(cfg.fileArgs),
		app.WithDestPath(cfg.destPath),
		app.WithRunner(cfg.runner),
	)

	if err := appInstance.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func validateConfig(cfg *config) error {
	if cfg.kubeContext == "" {
		return fmt.Errorf("kubernetes context is required")
	}
	if cfg.filePath == "" {
		return fmt.Errorf("file path is required")
	}
	if cfg.podName == "" {
		return fmt.Errorf("pod name is required")
	}
	if cfg.fileType != "auto" && cfg.fileType != "script" && cfg.fileType != "binary" {
		return fmt.Errorf("invalid file type: %s. Must be 'auto', 'script', or 'binary'", cfg.fileType)
	}
	return nil
}

var rootCmd = NewRootCmd()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
