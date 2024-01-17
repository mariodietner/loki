package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
	"github.com/grafana/loki/operator/controllers/loki/internal/lokistack"
)

// RulerConfigReconciler reconciles a RulerConfig object
type RulerConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=loki.grafana.com,resources=rulerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=loki.grafana.com,resources=rulerconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=loki.grafana.com,resources=rulerconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *RulerConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.Log.Info("Reconciliation for Request: ", "Name", req.Name, "Namespace", req.Namespace)

	var rc lokiv1.RulerConfig
	key := client.ObjectKey{Name: req.Name, Namespace: req.Namespace}
	if err := r.Get(ctx, key, &rc); err != nil {

		r.Log.Info("Reconciliation error getting resource with key: ", "Name", req.Name, "Namespace", req.Namespace, "key", key)

		if errors.IsNotFound(err) {

			r.Log.Info("Reconciliation error getting ObjectKey object not found", "error", err, "key", key)

			// RulerConfig not found, remove annotation from LokiStack.
			err = lokistack.RemoveRulerConfigAnnotation(ctx, r.Client, req.Name, req.Namespace)
			if err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	r.Log.Info("Reconciliation Annotating: ", "Name", req.Name, "Namespace", req.Namespace)

	err := lokistack.AnnotateForRulerConfig(ctx, r.Client, req.Name, req.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RulerConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {

	r.Log.Info("SetupWithManager started with ruler config for ", "config-object", &lokiv1.RulerConfig{})

	return ctrl.NewControllerManagedBy(mgr).
		For(&lokiv1.RulerConfig{}).
		Complete(r)
}
