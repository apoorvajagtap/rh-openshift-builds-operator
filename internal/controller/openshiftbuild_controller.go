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

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	osbv1alpha1 "github.com/redhat-openshift-builds/operator/api/v1alpha1"
	"github.com/redhat-openshift-builds/operator/internal/sharedresource"
)

// OpenShiftBuildReconciler reconciles a OpenShiftBuild object
type OpenShiftBuildReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger logr.Logger
}

//+kubebuilder:rbac:groups=operator.openshift.io,resources=openshiftbuilds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.openshift.io,resources=openshiftbuilds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=openshiftbuilds/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=secrets;configmaps;pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=sharedresource.openshift.io,resources=sharedconfigmaps;sharedsecrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create
//+kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=privileged,verbs=use
//+kubebuilder:rbac:groups=*,resources=services;endpoints;pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *OpenShiftBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("name", req.Name)
	logger.Info("Starting reconciliation")

	osb := &osbv1alpha1.OpenShiftBuild{}
	if err := r.Get(ctx, req.NamespacedName, osb); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Resource not found")
			return ctrl.Result{Requeue: false}, nil
		}
		logger.Error(err, "Retrieving object from cache")
		return ctrl.Result{}, nil
	}

	// TODO(user): Import upstream Shipwright Operator Reconcile method

	// Reconcile Shared Resources
	if err := r.ReconcileSharedResource(ctx, osb); err != nil {
		logger.Error(err, "failed to reconcile SharedResource")
		// apimeta.SetStatusCondition(&osb.Status.Conditions, metav1.Conditions{
		// 	Type: operatorv1alpha1.ConditionReady,
		// 	Status: metav1.ConditionFalse,
		// 	Reason: "Failed",
		// 	Message: fmt.Sprintf("Failed to reconcile OpenShiftBuild: %v", err),
		// })

		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

// ReconcileSharedResource creates and updates SharedResource objects
func (r *OpenShiftBuildReconciler) ReconcileSharedResource(ctx context.Context, osb *osbv1alpha1.OpenShiftBuild) error {
	logger := log.FromContext(ctx).WithValues("name", osb.ObjectMeta.Name)

	switch osb.Spec.SharedResource.State {
	case osbv1alpha1.Enabled:
		logger.Info("Starting SharedResource reconciliation...")
		sr := &sharedresource.SharedResource{}
		if err := sr.CreateSharedResources(osb); err != nil {
			logger.Error(err, "Failed reconciling SharedResources")
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenShiftBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&osbv1alpha1.OpenShiftBuild{}).
		Complete(r)
}
