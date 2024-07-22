package common

import (
	"github.com/manifestival/manifestival"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// InjectFinalizer appends finalizer to the passed resources metadata.
func InjectFinalizer(finalizer string) manifestival.Transformer {
	return func(u *unstructured.Unstructured) error {
		finalizers := u.GetFinalizers()
		if !controllerutil.ContainsFinalizer(u, finalizer) {
			finalizers = append(finalizers, finalizer)
			u.SetFinalizers(finalizers)
		}
		return nil
	}
}

func InjectContainerImages(images map[string]string) manifestival.Transformer {
	return func(u *unstructured.Unstructured) error {
		var d runtime.Object
		switch u.GetKind() {
		case "Deployment":
			d = &appsv1.Deployment{}
		case "DaemonSet":
			d = &appsv1.DaemonSet{}
		default:
			return nil
		}

		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, d)
		if err != nil {
			return err
		}
		// fetch all the containers & replace container images
		containers := FetchContainerList(d)
		for i, container := range containers {
			if url, exist := images[container.Image]; exist {
				containers[i].Image = url
			}
		}
		unstrObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(d)
		if err != nil {
			return err
		}
		u.SetUnstructuredContent(unstrObj)
		return nil
	}
}
