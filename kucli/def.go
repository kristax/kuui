package kucli

import (
	"context"
	"io"
	v1 "k8s.io/api/core/v1"
)

type KuCli interface {
	GetNamespaces(ctx context.Context) ([]v1.Namespace, error)
	GetPods(ctx context.Context, namespace string) ([]v1.Pod, error)
	GetPod(ctx context.Context, namespace, podName string) (*v1.Pod, error)
	DeletePod(ctx context.Context, namespace, podName string) error
	TailLogs(ctx context.Context, namespace, podName string, tailLines int64) ([]string, error)
	TailfLogs(ctx context.Context, namespace, podName string, tailLines int64) (chan string, error)
	ExecPod(ctx context.Context, namespace, pod string, in io.Reader, out io.Writer) error
}
