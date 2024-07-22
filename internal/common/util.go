package common

import (
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// imagesFromEnv fetches & returns the key-value pair of images from operator's env.
func ImagesFromEnv(prefix string) map[string]string {
	images := map[string]string{}
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		keyValue := strings.Split(env, "=")
		images[keyValue[0]] = keyValue[1]
	}
	return images
}

func FetchContainerList(obj runtime.Object) []corev1.Container {
	switch dep := obj.(type) {
	case *appsv1.DaemonSet:
		return *getDaemonSetContainersList(dep)
	case *appsv1.Deployment:
		return *getDeploymentContainersList(dep)
	}
	return nil
}

func getDaemonSetContainersList(ds *appsv1.DaemonSet) *[]corev1.Container {
	return &ds.Spec.Template.Spec.Containers
}

func getDeploymentContainersList(dep *appsv1.Deployment) *[]corev1.Container {
	return &dep.Spec.Template.Spec.Containers
}
