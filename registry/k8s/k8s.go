package k8s

import (
	"fmt"
	"github.com/eden/go-kratos-pkg/registry/k8s/k8s"
	"github.com/go-kratos/kratos/v2/registry"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func getDiscovery() (registry.Discovery, error) {
	clientSet, err := getK8sClientSet()
	if err != nil {
		return nil, fmt.Errorf("getK8sClientSet:%s", err.Error())
	}
	r := k8s.NewRegistry(clientSet)
	r.Start()
	return r, nil
}

func getRegistrar() (registry.Registrar, error) {
	clientSet, err := getK8sClientSet()
	if err != nil {
		return nil, fmt.Errorf("getK8sClientSet:%s", err.Error())
	}
	r := k8s.NewRegistry(clientSet)
	r.Start()
	return r, nil
}

func getK8sClientSet() (*kubernetes.Clientset, error) {
	restConfig, shouldReturn, returnValue, err := getKubeConfig()
	if shouldReturn {
		return returnValue, err
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("NewForConfig:%s", err.Error())
	}
	return clientSet, nil
}

func getKubeConfig() (*rest.Config, bool, *kubernetes.Clientset, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		home := homedir.HomeDir()
		kubeConfig := filepath.Join(home, ".kube", "config")
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, true, nil, fmt.Errorf("BuildConfigFromFlags:%s", err.Error())
		}
	}
	return restConfig, false, nil, nil
}
