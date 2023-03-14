/*
 * Parseable Server (C) 2022 - 2023 Parseable, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package parseabletenantcontroller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/parseablehq/parseable-operator/api/v1beta1"
	parseableiov1beta1 "github.com/parseablehq/parseable-operator/api/v1beta1"
)

// ParseableTenantReconciler reconciles a ParseableTenant object
type ParseableTenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=parseable.io,resources=parseabletenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=parseable.io,resources=parseabletenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=parseable.io,resources=parseabletenants/finalizers,verbs=update

func (r *ParseableTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	parseableCR := &v1beta1.ParseableTenant{}
	err := r.Get(context.TODO(), req.NamespacedName, parseableCR)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := reconcileParseable(r.Client, parseableCR); err != nil {
		return ctrl.Result{}, err
	} else {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ParseableTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&parseableiov1beta1.ParseableTenant{}).
		Complete(r)
}
