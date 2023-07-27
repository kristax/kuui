package kucli

import (
	"bufio"
	"context"
	"flag"
	"io"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"
)

type localCli struct {
	config    *rest.Config
	clientSet *kubernetes.Clientset
}

func NewLocalCli() KuCli {
	return &localCli{}
}

func (c *localCli) Init() error {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	c.config = config

	// create the clientset
	c.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func (c *localCli) GetNamespaces(ctx context.Context) ([]coreV1.Namespace, error) {
	list, err := c.clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *localCli) GetPods(ctx context.Context, namespace string) ([]coreV1.Pod, error) {
	podList, err := c.clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (c *localCli) GetPod(ctx context.Context, namespace, podName string) (*coreV1.Pod, error) {
	return c.clientSet.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
}

func (c *localCli) DeletePod(ctx context.Context, namespace, podName string) error {
	return c.clientSet.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}

func (c *localCli) TailLogs(ctx context.Context, namespace, podName string, tailLines int64) ([]string, error) {
	req := c.clientSet.CoreV1().Pods(namespace).GetLogs(podName, &coreV1.PodLogOptions{
		TailLines: &tailLines,
	})
	raw, err := req.DoRaw(ctx)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(raw), "\n"), nil
}

func (c *localCli) TailfLogs(ctx context.Context, namespace, podName string, tailLines int64) (chan string, error) {
	req := c.clientSet.CoreV1().Pods(namespace).GetLogs(podName, &coreV1.PodLogOptions{
		TypeMeta:                     metav1.TypeMeta{},
		Container:                    "",
		Follow:                       true,
		Previous:                     false,
		SinceSeconds:                 nil,
		SinceTime:                    nil,
		Timestamps:                   false,
		TailLines:                    &tailLines,
		LimitBytes:                   nil,
		InsecureSkipTLSVerifyBackend: false,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}
	ch := make(chan string, 10)
	go func() {
		defer close(ch)
		reader := bufio.NewReader(stream)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				str, err := reader.ReadString('\n')
				if err == io.EOF {
					return
				} else {
					ch <- str
				}
			}
		}
	}()
	return ch, nil
}

func (c *localCli) ExecPod(ctx context.Context, namespace, pod string, in io.Reader, out io.Writer) error {
	req := c.clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec").
		VersionedParams(&coreV1.PodExecOptions{
			TypeMeta:  metav1.TypeMeta{},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
			Container: "",
			Command:   []string{"bash"},
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             in,
		Stdout:            out,
		Stderr:            out,
		Tty:               true,
		TerminalSizeQueue: nil,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *localCli) GetDeployments(ctx context.Context, namespace string) ([]appV1.Deployment, error) {
	list, err := c.clientSet.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *localCli) GetDeployment(ctx context.Context, namespace, deployment string) (*appV1.Deployment, error) {
	return c.clientSet.AppsV1().Deployments(namespace).Get(ctx, deployment, metav1.GetOptions{})
}

func (c *localCli) UpdateDeployment(ctx context.Context, namespace string, deployment *appV1.Deployment) (*appV1.Deployment, error) {
	return c.clientSet.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

func (c *localCli) GetServices(ctx context.Context, namespace string) ([]coreV1.Service, error) {
	list, err := c.clientSet.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *localCli) GetService(ctx context.Context, namespace, deployment string) (*coreV1.Service, error) {
	return c.clientSet.CoreV1().Services(namespace).Get(ctx, deployment, metav1.GetOptions{})
}
