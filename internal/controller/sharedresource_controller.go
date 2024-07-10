/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/manifestival/manifestival"
	openshiftv1alpha1 "github.com/redhat-openshift-builds/operator/api/v1alpha1"
	"github.com/redhat-openshift-builds/operator/internal/common"
	"github.com/redhat-openshift-builds/operator/internal/sharedresource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SharedResourceReconciler reconciles a SharedResource object
type SharedResourceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Manifest manifestival.Manifest
	Logger   logr.Logger
}

//+kubebuilder:rbac:groups=operator.openshift.io,resources=sharedresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.openshift.io,resources=sharedresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=sharedresources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SharedResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *SharedResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	logger := r.Logger.WithValues("name", req.Name)
	logger.Info("Starting resource reconciliation...")

	// Applying transformers
	transformerfuncs := []manifestival.Transformer{}
	transformerfuncs = append(transformerfuncs, sharedresource.InjectFinalizer())
	transformerfuncs = append(transformerfuncs, manifestival.InjectOwner(&openshiftv1alpha1.OpenShiftBuild{}))
	transformerfuncs = append(transformerfuncs, manifestival.InjectNamespace(common.OpenShiftBuildNamespaceName))

	manifest, err := r.Manifest.Transform(transformerfuncs...)
	if err != nil {
		logger.Error(err, "transforming manifest")
		return sharedresource.RequeueWithError(err)
	}

	// TODO: Add deletion logic pertaining to the scenario when operator is being uninstalled.
	// i.e. if any CSI SharedResource object exists, the CRs shouldn't be deleted.

	// Rolling out the resources described on the manifests
	logger.Info("Applying manifests...")
	if err := manifest.Apply(); err != nil {
		logger.Error(err, "applying manifest")
		return sharedresource.RequeueWithError(err)
	}

	return sharedresource.NoRequeue()
	// return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SharedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&openshiftv1alpha1.OpenShiftBuild{}).
		Complete(r)
}
