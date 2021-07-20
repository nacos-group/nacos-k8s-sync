package model

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Task struct {
	Handler func() error
}

type KubeOptions struct {
	KubeConfig string

	WatchedNamespace string
}

type KubeClient interface {
	// KubeInformer returns an informer factory for kube client
	InformerFactory() informers.SharedInformerFactory

	Run(<-chan struct{})
}

type kubeClient struct {
	informerFactory informers.SharedInformerFactory
}

func NewKubeClient(option KubeOptions) (KubeClient, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", option.KubeConfig)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)

	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, DefaultResyncInterval,
		informers.WithNamespace(option.WatchedNamespace))

	if err != nil {
		return nil, err
	}

	return &kubeClient{
		informerFactory: informerFactory,
	}, nil
}

func (k *kubeClient) InformerFactory() informers.SharedInformerFactory {
	return k.informerFactory
}

func (k *kubeClient) Run(stop <-chan struct{}) {
	go k.informerFactory.Start(stop)
}
