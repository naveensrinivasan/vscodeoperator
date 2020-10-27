/*


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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	vscodev1alpha1 "github.com/naveensrinivasan/api/v1alpha1"
)

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vscode.naveensrinivasan.dev,resources=servers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vscode.naveensrinivasan.dev,resources=servers/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;watch;get;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=list;watch;get;patch
func (r *ServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("codeserver", req.NamespacedName)

	log.Info("reconciling codeserver")

	var code vscodev1alpha1.Server
	if err := r.Get(ctx, req.NamespacedName, &code); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	deployment, err := r.desiredDeployment(code)
	if err != nil {
		return ctrl.Result{}, err
	}
	svc, err := r.desiredService(code)
	if err != nil {
		return ctrl.Result{}, err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("codeserver-controller")}

	err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Patch(ctx, &svc, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconciled codeserver")
	return ctrl.Result{}, nil
}

func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vscodev1alpha1.Server{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
