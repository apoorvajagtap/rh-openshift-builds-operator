package sharedresource

import (
	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	openshiftv1alpha1 "github.com/redhat-openshift-builds/operator/api/v1alpha1"
	"github.com/redhat-openshift-builds/operator/internal/common"
	"k8s.io/apimachinery/pkg/api/errors"
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

// UpdateSharedResource transforms the manifests, and applies or deletes them based on SharedResource.State.
func (sr *SharedResource) UpdateSharedResources(owner *openshiftv1alpha1.OpenShiftBuild) error {
	logger := sr.Logger.WithValues("name", owner.Name)
	sr.State = owner.Spec.SharedResource.State
	images := common.ImagesFromEnv(common.SharedResourceImagePrefix)

	// Applying transformers
	transformerfuncs := []manifestival.Transformer{}
	transformerfuncs = append(transformerfuncs, manifestival.InjectOwner(owner))
	transformerfuncs = append(transformerfuncs, manifestival.InjectNamespace(common.OpenShiftBuildNamespaceName))
	transformerfuncs = append(transformerfuncs, common.InjectContainerImages(images))
	if sr.State == openshiftv1alpha1.Enabled && owner.DeletionTimestamp.IsZero() {
		transformerfuncs = append(transformerfuncs, common.InjectFinalizer(common.OpenShiftBuildFinalizerName))
	}

	manifest, err := sr.Manifest.Transform(transformerfuncs...)
	if err != nil {
		logger.Error(err, "transforming manifest")
		return err
	}

	// Remove the finalizers if owner is set for deletion or SharedResource is disabled
	// & perform the deletion if SharedResource is disabled.
	if !owner.DeletionTimestamp.IsZero() || sr.State == openshiftv1alpha1.Disabled {
		return sr.deleteManifests(&manifest)
	}

	logger.Info("Applying manifests...")
	return manifest.Apply()
}

// deleteManifests removes the applied finalizer from all manifest.Resources &
// deletes the resources if SharedResource.State is disabled.
func (sr *SharedResource) deleteManifests(manifest *manifestival.Manifest) error {
	mfc := sr.Manifest.Client
	for _, res := range manifest.Resources() {
		obj, err := mfc.Get(&res)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		// removes finalizers
		if len(obj.GetFinalizers()) > 0 {
			obj.SetFinalizers([]string{})
			if err := mfc.Update(obj); err != nil {
				return err
			}
		}

		if sr.State == openshiftv1alpha1.Disabled {
			sr.Logger.Info("Deleting SharedResources")
			mfc.Delete(&res)
		}
	}
	return nil
}
