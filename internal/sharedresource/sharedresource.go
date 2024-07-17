package sharedresource

import (
	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	openshiftv1alpha1 "github.com/redhat-openshift-builds/operator/api/v1alpha1"
	"github.com/redhat-openshift-builds/operator/internal/common"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// SharedResource type defines methods to Get, Create v1alpha1.SharedResource resource
type SharedResource struct {
	Logger   logr.Logger
	Manifest manifestival.Manifest
	State    openshiftv1alpha1.State
}

// New creates new instance of SharedResource type
func New(manifest manifestival.Manifest) *SharedResource {
	return &SharedResource{
		Manifest: manifest,
	}
}

// CreateSharedResources applies the required set of manifests
func (sr *SharedResource) ApplySharedResources(owner *openshiftv1alpha1.OpenShiftBuild, state openshiftv1alpha1.State) error {
	logger := sr.Logger.WithValues("name", owner.Name)
	sr.State = state

	// Applying transformers
	transformerfuncs := []manifestival.Transformer{}
	transformerfuncs = append(transformerfuncs, sr.InjectFinalizer())
	transformerfuncs = append(transformerfuncs, manifestival.InjectOwner(owner))
	transformerfuncs = append(transformerfuncs, manifestival.InjectNamespace(common.OpenShiftBuildNamespaceName))

	manifest, err := sr.Manifest.Transform(transformerfuncs...)
	if err != nil {
		logger.Error(err, "transforming manifest")
		return err
	}

	if !owner.GetDeletionTimestamp().IsZero() || state == openshiftv1alpha1.Disabled {
		// Remove the finalizers if owner is set for deletion, or SharedResource.State is disabled
		logger.Info("Removing finalizers")
		return sr.deleteManifests(&manifest)
	}

	logger.Info("Applying manifests...")
	return manifest.Apply()
}

// InjectFinalizer appends finalizer to the passed resources metadata.
func (sr *SharedResource) InjectFinalizer() manifestival.Transformer {
	return func(u *unstructured.Unstructured) error {
		finalizers := u.GetFinalizers()
		if !controllerutil.ContainsFinalizer(u, common.OpenShiftBuildFinalizerName) {
			finalizers = append(finalizers, common.OpenShiftBuildFinalizerName)
			u.SetFinalizers(finalizers)
		}
		return nil
	}
}

// deleteManifests removes the applied finalizer string from all manifest.Resources &
// if SharedResource.State is disabled continues with the deletion of all resources.
func (sr *SharedResource) deleteManifests(manifest *manifestival.Manifest) error {
	mfc := sr.Manifest.Client
	for _, res := range manifest.Resources() {
		obj, err := mfc.Get(&res)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if len(obj.GetFinalizers()) > 0 {
			obj.SetFinalizers([]string{})
			if err := mfc.Update(obj); err != nil {
				return err
			}
		}

		// TODO: uncomment following after confirming the deletion behavior.
		// if sr.State == openshiftv1alpha1.Disabled {
		// 	sr.Logger.Info("Deleting SharedResources")
		// 	mfc.Delete(&res)
		// }
	}
	return nil
}
