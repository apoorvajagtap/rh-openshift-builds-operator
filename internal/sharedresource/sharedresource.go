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
func (sr *SharedResource) CreateSharedResources(owner *openshiftv1alpha1.OpenShiftBuild) error {
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

	// Rolling out the resources described on the manifests
	logger.Info("Applying manifests...")
	return manifest.Apply()
}

// TODO: Add deletion logic pertaining to the scenario when operator is being uninstalled.
// i.e. if any CSI SharedResource object exists, the CRs shouldn't be deleted.

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
