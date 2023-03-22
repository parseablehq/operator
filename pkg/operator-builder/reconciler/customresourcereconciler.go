package reconciler

import (
	"os"
	"time"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CustomResourceReconciler struct {
	Recorder      record.EventRecorder
	ReconcileWait time.Duration
}

func NewCustomResourceReconciler(mgr ctrl.Manager, customResourceName string) *CustomResourceReconciler {
	return &CustomResourceReconciler{
		ReconcileWait: LookupReconcileTime(),
		Recorder:      mgr.GetEventRecorderFor(customResourceName),
	}
}

func LookupReconcileTime() time.Duration {
	val, exists := os.LookupEnv("RECONCILE_WAIT")
	if !exists {
		return time.Second * 10
	} else {
		v, err := time.ParseDuration(val)
		if err != nil {
			// Exit Program if not valid
			os.Exit(1)
		}
		return v
	}
}
