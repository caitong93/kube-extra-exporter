package pod

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type Lister interface {
	List() ([]*v1.Pod, error)
}

type wrappedPodLister struct {
	listers.PodLister
}

func (l wrappedPodLister) List() ([]*v1.Pod, error) {
	return l.PodLister.List(labels.Everything())
}

// NewLister creates a lister to list pods on local node.
func NewLister(ctx context.Context, kubeClient kubernetes.Interface, node string) Lister {
	lw := createPodListWatch(kubeClient, node)
	indexer, reflector := cache.NewNamespaceKeyedIndexerAndReflector(lw, &v1.Pod{}, 5*time.Minute)

	go reflector.Run(ctx.Done())

	return wrappedPodLister{listers.NewPodLister(indexer)}
}

func fieldsSelector(nodeName string) fields.Selector {
	sel, err := fields.ParseSelector("spec.nodeName=" + nodeName)
	if err != nil {
		panic(fmt.Sprintf("error parse selector: %v", err))
	}
	return sel
}

func createPodListWatch(kubeClient kubernetes.Interface, node string) cache.ListerWatcher {
	const ns = ""
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			opts.FieldSelector = fieldsSelector(node).String()
			return kubeClient.CoreV1().Pods(ns).List(opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			opts.FieldSelector = fieldsSelector(node).String()
			return kubeClient.CoreV1().Pods(ns).Watch(opts)
		},
	}
}
