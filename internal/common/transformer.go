package common

import (
	"github.com/manifestival/manifestival"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
