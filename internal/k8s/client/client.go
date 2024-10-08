package client

import (
	"github.com/spf13/viper"
	"k8res/pkg/logger"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"

	clientSet "k8s.io/client-go/kubernetes"
	clientReset "k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type K8s struct {
	ClientSet     clientSet.Interface
	MetricsClient *metrics.Clientset
	RestConfig    *clientReset.Config
	namespace     string // current namespace
	outOfCluster  bool   // out of cluster config
}

// New creates a new k8s client
// cluster - used for get kubeconfig. refer getRestConfig
func New(cluster string) *K8s {
	var err error
	k := K8s{}

	k.RestConfig, err = k.getRestConfig(cluster)
	if err != nil {
		logger.Fatalf("get %s cluster config failed: %v", cluster, err)
		return nil
	}
	k.ClientSet, err = clientSet.NewForConfig(k.RestConfig)
	if err != nil {
		logger.Fatalf("can not create kubernetes clientSet: %v", err)
		return nil
	}

	k.MetricsClient, err = metrics.NewForConfig(k.RestConfig)
	if err != nil {
		logger.Fatalf("can not create kubernetes metric clientSet: %v", err)
		return nil
	}
	return &k
}

// GetVersion returns the version of the kubernetes cluster that is running
func (k *K8s) GetVersion() (string, error) {
	version, err := k.ClientSet.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func (k *K8s) SetNamespace(namespace string) {
	k.namespace = namespace
}

func (k *K8s) GetNamespace() string {
	if k.namespace == "" {
		logger.Warn("can not get current namespace, use 'default'")
		k.namespace = "default"
	}
	return k.namespace
}

// GetCurNamespace will return the current namespace for the running program
// Checks for the user passed ENV variable POD_NAMESPACE if not available
// pulls the namespace from pod, if not returns ""
func (k *K8s) GetCurNamespace() string {
	var namespace string
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	if k.outOfCluster {
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{})
		namespace, _, _ = kubeconfig.Namespace()
		return namespace
	}
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if namespace = strings.TrimSpace(string(data)); len(namespace) > 0 {
			return namespace
		}
	}
	if namespace == "" {
		logger.Warn("can not get current namespace, use 'default'")
		namespace = "default"
	}
	return namespace
}

// getRestConfig will return a rest config for the kubernetes cluster
func (k *K8s) getRestConfig(cluster string) (*clientReset.Config, error) {
	k.outOfCluster = true
	kubeconfigPath := viper.GetString(cluster + ".kube-config")
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		if kubeconfigPath != "" {
			logger.Infof("get %s cluster default config", cluster)
		}
	} else {
		logger.Infof("get %s cluster config", cluster)
	}
	if kubeconfigPath == "" {
		logger.Infof("use %s cluster internal config", cluster)
		k.outOfCluster = false
		return clientReset.InClusterConfig()
	}
	logger.Infof("use %s cluster out config %s", cluster, kubeconfigPath)
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}
