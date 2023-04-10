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
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	"github.com/parseablehq/parseable-operator/api/v1beta1"
	parseableiov1beta1 "github.com/parseablehq/parseable-operator/api/v1beta1"
)

// ParseableTenantReconciler reconciles a ParseableTenant object
type ParseableTenantReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// reconcile time duration, defaults to 10s
	ReconcileWait time.Duration
	Recorder      record.EventRecorder
}

func NewParseableTenantReconciler(mgr ctrl.Manager) *ParseableTenantReconciler {
	initLogger := ctrl.Log.WithName("controllers").WithName("parseable-tenant")
	return &ParseableTenantReconciler{
		Client:        mgr.GetClient(),
		Log:           initLogger,
		Scheme:        mgr.GetScheme(),
		ReconcileWait: LookupReconcileTime(initLogger),
		Recorder:      mgr.GetEventRecorderFor("parseable-operator"),
	}
}

//+kubebuilder:rbac:groups=parseable.io,resources=parseabletenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=parseable.io,resources=parseabletenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=parseable.io,resources=parseabletenants/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *ParseableTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)

	parseableCR := &v1beta1.ParseableTenant{}
	err := r.Get(context.TODO(), req.NamespacedName, parseableCR)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.do(ctx, parseableCR, logr); err != nil {
		logr.Error(err, err.Error())
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

func LookupReconcileTime(log logr.Logger) time.Duration {
	val, exists := os.LookupEnv("RECONCILE_WAIT")
	if !exists {
		return time.Second * 10
	} else {
		v, err := time.ParseDuration(val)
		if err != nil {
			log.Error(err, err.Error())
			// Exit Program if not valid
			os.Exit(1)
		}
		return v
	}
}
