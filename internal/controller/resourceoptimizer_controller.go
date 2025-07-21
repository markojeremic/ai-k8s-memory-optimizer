package controller

import (
    "context"
    "fmt"
    "time"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    "github.com/markojeremic/ai-k8s-memory-optimizer/metrics"
    "github.com/yourname/ai-k8s-memory-optimizer/pr"
)

type PodWatcherReconciler struct {
    client.Client
}

func (r *PodWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    var pod corev1.Pod
    if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    if pod.Status.Phase != corev1.PodRunning {
        return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
    }

    usage, err := metrics.GetMemoryUsage(pod.Namespace, pod.Name)
    if err != nil {
        logger.Error(err, "Failed to get memory usage")
        return ctrl.Result{}, err
    }

    suggestions := map[string]int64{}
    for _, container := range pod.Spec.Containers {
        reqMem := container.Resources.Requests.Memory().Value()
        usedMem := usage[container.Name]

        if float64(usedMem) < 0.3*float64(reqMem) {
            // Example: suggest 50% of request if usage is low
            suggested := int64(0.5 * float64(reqMem))
            suggestions[container.Name] = suggested
        }
    }

    if len(suggestions) > 0 {
        err := pr.CreateOptimizationPR(pod, suggestions)
        if err != nil {
            logger.Error(err, "Failed to create PR")
        }
    }

    return ctrl.Result{RequeueAfter: 6 * time.Hour}, nil
}

func (r *PodWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&corev1.Pod{}).
        Complete(r)
}
