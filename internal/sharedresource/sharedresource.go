package sharedresource

import (
	"path/filepath"
	"slices"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	"github.com/redhat-openshift-builds/operator/internal/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// FinalizerAnnotation annotation string appended to finalizer slice.
	FinalizerAnnotation = "finalizer.operator.sharedresource.io"
)

// SharedResource type defines methods to Get, Create v1alpha1.SharedResource resource
type SharedResource struct {
	Client   client.Client
	Logger   logr.Logger
	Manifest manifestival.Manifest
}

// New creates new instance of SharedResource type
func New(client client.Client) *SharedResource {
	return &SharedResource{
		Client: client,
	}
}

// InjectFinalizer appends finalizer to the passed resources metadata.
func InjectFinalizer() manifestival.Transformer {
	return func(u *unstructured.Unstructured) error {
		finalizers := u.GetFinalizers()
		if !slices.Contains(finalizers, FinalizerAnnotation) {
			finalizers = append(finalizers, FinalizerAnnotation)
			u.SetFinalizers(finalizers)
		}

		return nil
	}
}

// setupManifestival instantiate manifestival with local controller attributes.
func (sr *SharedResource) setupManifestival() error {
	var err error
	sr.Manifest, err = common.SetupManifestival(sr.Client, filepath.Join("config", "sharedresource"), true, sr.Logger)
	if err != nil {
		return err
	}

	return nil
}

// RequeueWithError triggers a object requeue because the informed error happend.
func RequeueWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, err
}

// NoRequeue all done, the object does not need reconciliation anymore.
func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: false}, nil
}
