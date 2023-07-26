package kucli

import (
	"context"
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewLocalCli(t *testing.T) {
	cli := NewLocalCli()
	ioc.RunTest(t, app.SetComponents(cli))
	ctx := context.Background()

	t.Run("Test_localCli_GetNamespaces", func(t *testing.T) {
		namespaceList, err := cli.GetNamespaces(ctx)
		assert.NoError(t, err)

		fmt.Println("count of namespaces: ", len(namespaceList))
	})

	t.Run("Test_localCli_GetPods", func(t *testing.T) {
		pods, err := cli.GetPods(ctx, "")
		assert.NoError(t, err)
		fmt.Println("count of pods: ", len(pods))
	})

	t.Run("Test_localCli_GetLogs", func(t *testing.T) {
		namespace := "gi2-saas-platform-dev"
		pods, err := cli.GetPods(ctx, namespace)
		assert.NoError(t, err)
		fmt.Println(pods)
		logCh, err := cli.TailfLogs(ctx, namespace, pods[0].Name, 100)
		assert.NoError(t, err)
		for ch := range logCh {
			fmt.Println(ch)
		}
	})

	t.Run("Test_localCli_TailLog", func(t *testing.T) {

	})
}

func TestExec(t *testing.T) {
	cli := NewLocalCli()
	ioc.RunTest(t, app.SetComponents(cli))
	ctx := context.Background()

	err := cli.ExecPod(ctx, "api-adaptor-dev", "api-adaptor-5c8c958f9b-v4skt", os.Stdin, os.Stdout)
	assert.NoError(t, err)
}

func TestTailLog(t *testing.T) {
	cli := NewLocalCli()
	ioc.RunTest(t, app.SetComponents(cli))
	ctx := context.Background()
	logs, err := cli.TailLogs(ctx, "api-adaptor-dev", "api-adaptor-5c8c958f9b-v4skt")
	assert.NoError(t, err)
	for _, log := range logs {

		fmt.Println(log)
	}
}
