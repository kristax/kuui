package kucli

import (
	"context"
	"io"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

type KuCli interface {
	GetNamespaces(ctx context.Context) ([]coreV1.Namespace, error)

	GetPods(ctx context.Context, namespace string) ([]coreV1.Pod, error)
	GetPod(ctx context.Context, namespace, podName string) (*coreV1.Pod, error)
	DeletePod(ctx context.Context, namespace, podName string) error
	TailLogs(ctx context.Context, namespace, podName string, tailLines int64) ([]string, error)
	TailfLogs(ctx context.Context, namespace, podName string, tailLines int64) (chan string, error)
	ExecPod(ctx context.Context, namespace, pod string, in io.Reader, out io.Writer) error

	GetDeployments(ctx context.Context, namespace string) ([]appV1.Deployment, error)
	GetDeployment(ctx context.Context, namespace, deployment string) (*appV1.Deployment, error)
	UpdateDeployment(ctx context.Context, namespace string, deployment *appV1.Deployment) (*appV1.Deployment, error)

	GetServices(ctx context.Context, namespace string) ([]coreV1.Service, error)
	GetService(ctx context.Context, namespace, deployment string) (*coreV1.Service, error)
}
