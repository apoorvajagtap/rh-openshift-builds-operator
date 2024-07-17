package sharedresource

import (
	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	openshiftv1alpha1 "github.com/redhat-openshift-builds/operator/api/v1alpha1"
	"github.com/redhat-openshift-builds/operator/internal/common"
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

	if !owner.GetDeletionTimestamp().IsZero() {
		// removing only the finalizers as deletion will be taken care by the owner.
		logger.Info("Removing finalizers")
		sr.unsetFinalizer(&manifest)
	}

	if state == openshiftv1alpha1.Disabled {
		// Remove finalizer
		logger.Info("Removing finalizers")
		sr.unsetFinalizer(&manifest)
	}

	if state == openshiftv1alpha1.Enabled {
		// Rolling out the resources described on the manifests
		logger.Info("Applying manifests...")
		return manifest.Apply()
	}

	return nil
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

// unsetFinalizer remove all instances of finalizer string.
func (sr *SharedResource) unsetFinalizer(manifest *manifestival.Manifest) error {
	mfc := sr.Manifest.Client
	for _, res := range manifest.Resources() {
		obj, err := mfc.Get(&res)
		if err != nil {
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
