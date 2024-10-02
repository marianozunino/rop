package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func (app *App) buildConfigWithContextAndNamespaceFromFlags(kubeconfigPath string) (*rest.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	overrides := &clientcmd.ConfigOverrides{
		CurrentContext: app.kubeContext,
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		overrides,
	)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// If namespace is empty, use the one from the context
	if app.namespace == "" {
		log.Debug().Msgf("Using namespace from context: %s", app.kubeContext)
		app.namespace, _, err = clientConfig.Namespace()
		if err != nil {
			return nil, err
		}
	}

	log.Debug().Msgf("Using namespace: %s", app.namespace)

	// We don't set the namespace in the config, as rest.Config doesn't have a Namespace field
	// Instead, we'll use app.namespace when creating clients or making API calls
	return config, nil
}

func (app *App) initializeKubernetesClient() error {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	log.Debug().Msgf("Using kubeconfig: %s", kubeconfig)

	// config, err := buildConfigWithContextFromFlags(app.kubeContext, kubeconfig)
	config, err := app.buildConfigWithContextAndNamespaceFromFlags(kubeconfig)
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

func (app *App) getRunningPod(ctx context.Context) (*corev1.Pod, error) {
	log.Debug().Msgf("Getting pod %s from namespace %s", app.podName, app.namespace)
	pods, err := app.clientset.CoreV1().Pods(app.namespace).List(ctx, metav1.ListOptions{
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

func GetAvailableContexts() ([]string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	var contexts []string
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	return contexts, nil
}

func GetAvailableNamespaces(ctx string) ([]string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: ctx,
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	namespaceList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var namespaces []string
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}
