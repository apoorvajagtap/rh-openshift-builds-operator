package sharedresource

import (
	"path/filepath"
	"slices"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	osbv1alpha1 "github.com/redhat-openshift-builds/operator/api/v1alpha1"
	"github.com/redhat-openshift-builds/operator/internal/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// CreateSharedResources applies the required set of manifests
func (sr *SharedResource) CreateSharedResources(owner *osbv1alpha1.OpenShiftBuild) error {
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

	if err := sr.setupManifestival(); err != nil {
		return err
	}
	// TODO: Add deletion logic pertaining to the scenario when operator is being uninstalled.
	// i.e. if any CSI SharedResource object exists, the CRs shouldn't be deleted.

	// Rolling out the resources described on the manifests
	logger.Info("Applying manifests...")
	if err := manifest.Apply(); err != nil {
		logger.Error(err, "applying manifest")
		return err
	}

	return nil
}

// InjectFinalizer appends finalizer to the passed resources metadata.
func (sr *SharedResource) InjectFinalizer() manifestival.Transformer {
	return func(u *unstructured.Unstructured) error {
		finalizers := u.GetFinalizers()
		if !slices.Contains(finalizers, common.OpenShiftBuildFinalizerName) {
			finalizers = append(finalizers, common.OpenShiftBuildFinalizerName)
			u.SetFinalizers(finalizers)
		}

		return nil
	}
}

// setupManifestival instantiate manifestival with local controller attributes.
func (sr *SharedResource) setupManifestival() error {
	var err error
	sr.Manifest, err = common.SetupManifestival(sr.Client, filepath.Join("internal", "sharedresource", "manifests"), true, sr.Logger)
	if err != nil {
		return err
	}

	return nil
}
