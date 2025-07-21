package metrics

import (
    "context"

    "k8s.io/apimachinery/pkg/types"
    "k8s.io/metrics/pkg/client/clientset/versioned"
    "k8s.io/client-go/rest"
)

func GetMemoryUsage(namespace, podName string) (map[string]int64, error) {
    config, err := rest.InClusterConfig()
    if err != nil {
        return nil, err
    }

    metricsClient, err := versioned.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), podName, types.GetOptions{})
    if err != nil {
        return nil, err
    }

    usage := make(map[string]int64)
    for _, c := range podMetrics.Containers {
        if memQuantity, ok := c.Usage["memory"]; ok {
            usage[c.Name] = memQuantity.Value()
        }
    }

    return usage, nil
}
