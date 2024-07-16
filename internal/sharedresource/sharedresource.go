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
}

// New creates new instance of SharedResource type
func New(manifest manifestival.Manifest) *SharedResource {
	return &SharedResource{
		Manifest: manifest,
	}
}

// CreateSharedResources applies the required set of manifests
func (sr *SharedResource) CreateSharedResources(owner *openshiftv1alpha1.OpenShiftBuild, delete bool) error {
	logger := sr.Logger.WithValues("name", owner.Name)

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
		// remove finalizers
		logger.Info("Removing finalizers")
		sr.unsetFinalizer(*owner)
	}

	// if delete {
	// 	// Remove finalizer
	// 	sr.Logger.Info("Removing finalizers")
	// 	sr.unsetFinalizer(*owner)

	// 	// Delete manifests
	// 	sr.Logger.Info("Deleting SharedResources...")
	// 	return sr.Manifest.Delete()
	// }

	// Rolling out the resources described on the manifests
	logger.Info("Applying manifests...")
	return manifest.Apply()
}

// TODO: Add deletion logic pertaining to the scenario when operator is being uninstalled.
// i.e. if any CSI SharedResource object exists, the CRs shouldn't be deleted.
// func (sr *SharedResource) DeleteSharedResources(owner *openshiftv1alpha1.OpenShiftBuild) error {
// 	// Remove finalizer
// 	sr.Logger.Info("Removing finalizers")
// 	sr.unsetFinalizer(*owner)

// 	// Delete manifests
// 	sr.Logger.Info("Deleting SharedResources...")
// 	return sr.Manifest.Delete()
// 	// Delete the applied resources
// 	return nil
// }

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
func (sr *SharedResource) unsetFinalizer(owner openshiftv1alpha1.OpenShiftBuild) {
	finalizers := []string{}
	for _, f := range owner.GetFinalizers() {
		if f != common.OpenShiftBuildFinalizerName {
			finalizers = append(finalizers, f)
		}
	}
	owner.SetFinalizers(finalizers)
}
