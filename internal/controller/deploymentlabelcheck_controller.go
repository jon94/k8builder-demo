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
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	demov1 "operator-demo/api/v1"

	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentLabelCheckReconciler reconciles a DeploymentLabelCheck object
type DeploymentLabelCheckReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.demo.jonlimpw.io,resources=deploymentlabelchecks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.demo.jonlimpw.io,resources=deploymentlabelchecks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.demo.jonlimpw.io,resources=deploymentlabelchecks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DeploymentLabelCheck object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile

// Controller will only kick start if CRD object is created from the start. If no CRD object created, this will not log any error. That is actually handled by SetUpManagerComponent. Code is commented out
func (r *DeploymentLabelCheckReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Enter Reconcile", "req", req)

	// Fetch the DeploymentLabelCheck CRD
	dlc := &demov1.DeploymentLabelCheck{}
	err := r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.Name}, dlc)

	l.Info("Enter Reconcile", "spec", dlc.Spec, "status", dlc.Status)

	if err == nil {
		l.Info("DLC Found")
		// call func reconcileDeployment if DLC is found
		r.reconcileDeployment(ctx, dlc, req)
		return ctrl.Result{}, nil
	}

	if !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	l.Info("DLC Not found")
	return ctrl.Result{}, nil
}

// Fetch the Deployment resource using info from CRD
func (r *DeploymentLabelCheckReconciler) reconcileDeployment(ctx context.Context, dlc *demov1.DeploymentLabelCheck, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Enter DeploymentReconcile", "req", req)

	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Namespace: dlc.Spec.Namespace, Name: dlc.Spec.DeploymentName}, deployment)
	l.Info("Enter DeploymentReconcile", "spec", deployment.Spec, "status", deployment.Status)

	if err == nil {
		l.Info("Deployment Found", "name", deployment.Name, "namespace", deployment.Namespace)
		r.checklabel(ctx, dlc, deployment, req)
		return ctrl.Result{}, nil
	}

	if !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	l.Info("Deployment Not found")
	return ctrl.Result{}, nil
}

// check if label is present, if yes do nothing. If no, label the deployment.

func (r *DeploymentLabelCheckReconciler) checklabel(ctx context.Context, dlc *demov1.DeploymentLabelCheck, deployment *appsv1.Deployment, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Enter CheckLabelPresent", "req", req)
	// Check if the label is already present in the template metadata
	if deployment.Spec.Template.ObjectMeta.Labels != nil {
		if value, exists := deployment.Spec.Template.ObjectMeta.Labels["admission.datadoghq.com/enabled"]; exists && value == "true" {
			// Label is already present and set to true, do nothing
			l.Info("Label admission.datadoghq.com/enabled is already set to true")
			return ctrl.Result{}, nil
		}
	}
	if value, exists := deployment.Spec.Template.ObjectMeta.Labels["admission.datadoghq.com/enabled"]; exists && value == "false" {
		// Label is already present and set to false, do nothing
		l.Info("Label admission.datadoghq.com/enabled is already set to false")
		return ctrl.Result{}, nil
	} else {
		l.Info("Not there > Label admission.datadoghq.com/enabled")
		if deployment.Spec.Template.ObjectMeta.Labels == nil {
			deployment.Spec.Template.ObjectMeta.Labels = make(map[string]string)
		}
		deployment.Spec.Template.ObjectMeta.Labels["admission.datadoghq.com/enabled"] = "true"
		if err := r.Update(ctx, deployment); err != nil {
			l.Error(err, "Failed to update Deployment labels")
			return ctrl.Result{}, err
		}
		//rollout restart deployment
		if err := r.restartDeployment(ctx, deployment); err != nil {
			l.Error(err, "Failed to restart Deployment")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// Implement Function for rollout restart via incrementing restartAt annotation
func (r *DeploymentLabelCheckReconciler) restartDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	l := log.FromContext(ctx)

	// Retrieve the deployment's current annotations
	annotations := deployment.ObjectMeta.Annotations

	// Increment the "deployment.kubernetes.io/restartedAt" annotation to trigger a rollout restart
	if annotations == nil {
		annotations = make(map[string]string)
	}
	restartedAt, exists := annotations["deployment.kubernetes.io/restartedAt"]
	if !exists {
		restartedAt = "1"
	} else {
		restartedAtInt, err := strconv.Atoi(restartedAt)
		if err != nil {
			return err
		}
		restartedAt = strconv.Itoa(restartedAtInt + 1)
	}

	annotations["deployment.kubernetes.io/restartedAt"] = restartedAt
	deployment.ObjectMeta.Annotations = annotations
	l.Info("Deployment restarted using rollout restart strategy")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentLabelCheckReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// // Check if DeploymentLabelCheck exists before setting up the controller
	// dlc := &demov1.DeploymentLabelCheck{}
	// err := mgr.GetClient().Get(context.Background(), client.ObjectKey{Namespace: "your-namespace", Name: "your-name"}, dlc)

	// if errors.IsNotFound(err) {
	// 	log.Log.Info("DeploymentLabelCheck not found at the start of the controller setup")
	// } else if err != nil {
	// 	log.Log.Error(err, "Error checking for DeploymentLabelCheck existence")
	// 	return err
	// } else {
	// 	log.Log.Info("DeploymentLabelCheck found at the start of the controller setup")
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1.DeploymentLabelCheck{}).
		Complete(r)
}
