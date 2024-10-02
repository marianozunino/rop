package app

import (
	"context"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func (app *App) initializeKubernetesClient() error {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s\n", kubeconfig)

	config, err := buildConfigWithContextFromFlags(app.kubeContext, kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes client: %w", err)
	}
	app.config = config

	clientset, err := kubernetes.NewForConfig(app.config)
	if err != nil {
		return fmt.Errorf("error creating Kubernetes clientset: %w", err)
	}
	app.clientset = clientset

	return nil
}

func buildConfigWithContextFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func (app *App) getRunningPod(ctx context.Context) (*corev1.Pod, error) {
	pods, err := app.clientset.CoreV1().Pods("oc-app").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Running",
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", app.podName),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting pod: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no running pods found for %s", app.podName)
	}
	return &pods.Items[0], nil
}
