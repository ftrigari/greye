package ports

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// ImportProcessApplication defines application-level behaviors.
type ImportProcessApplication interface {
	GetKubernetesMonitoringObject() watch.Interface
}

// ImportProcessRepository defines repository-level behaviors for Kubernetes.
type ImportProcessRepository interface {
	GetConfig() (*kubernetes.Clientset, error)
}
