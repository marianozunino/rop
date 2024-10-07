package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	Clientset    *kubernetes.Clientset
	ClientConfig clientcmd.ClientConfig
	Config       *rest.Config
	Namespace    string
	Context      string
}

func NewClient(kubeContext, namespace string) (*Client, error) {
	client := &Client{
		Context:   kubeContext,
		Namespace: namespace,
	}

	if err := client.initializeKubernetesClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}

	return client, nil
}

func (c *Client) buildConfigWithContextAndNamespace(kubeconfigPath string) (*rest.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	overrides := &clientcmd.ConfigOverrides{CurrentContext: c.Context}

	c.ClientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	config, err := c.ClientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config: %w", err)
	}

	if c.Namespace == "" {
		c.Namespace, _, err = c.ClientConfig.Namespace()
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace from context: %w", err)
		}
		log.Debug().Msgf("Using namespace from context: %s", c.Namespace)
	}

	log.Debug().Msgf("Using namespace: %s", c.Namespace)
	return config, nil
}

func (c *Client) initializeKubernetesClient() error {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		kubeconfig = envPath
	}
	log.Debug().Msgf("Using kubeconfig: %s", kubeconfig)

	config, err := c.buildConfigWithContextAndNamespace(kubeconfig)
	if err != nil {
		return err
	}
	c.Config = config

	c.Clientset, err = kubernetes.NewForConfig(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return nil
}

