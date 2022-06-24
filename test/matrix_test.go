//go:build e2e
// +build e2e

/*
Copyright 2019 The Tekton Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/tektoncd/pipeline/test/parse"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knativetest "knative.dev/pkg/test"
	"knative.dev/pkg/test/helpers"
)

var requireAlphaMinimalFeatureFlags = requireAnyGate(map[string]string{
	"enable-api-fields": "alpha",
	"embedded-status":   "minimal",
})

func TestPipelineWithSingleMatrixedPipelineTask(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c, namespace := setup(ctx, t, requireAlphaMinimalFeatureFlags)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, namespace) }, t.Logf)
	defer tearDown(ctx, t, c, namespace)

	task := parse.MustParseTask(t, fmt.Sprintf(`
metadata:
  name: mytask
  namespace: %s
spec:
  params:
    - name: platform
      type: string
    - name: browser
      type: string
  steps:
    - name: echo
      image: alpine
      script: |
        echo "$(params.platform) and $(params.browser)"
`, namespace))
	if _, err := c.TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create dag Task: %s", err)
	}

	pipeline := parse.MustParsePipeline(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  tasks:
  - name: pipelinetask
    matrix:
    - name: platform
      value:
      - linux
      - mac
      - windows
    - name: browser
      value:
      - chrome
      - safari
      - firefox
    taskRef:
      name: %s
`, helpers.ObjectNameForTest(t), namespace, task.Name))
	if _, err := c.PipelineClient.Create(ctx, pipeline, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pipeline: %s", err)
	}

	pipelineRun := getPipelineRun(t, namespace, pipeline.Name)
	if _, err := c.PipelineRunClient.Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pipeline Run `%s`: %s", pipelineRun.Name, err)
	}

	if err := WaitForPipelineRunState(ctx, c, pipelineRun.Name, timeout, PipelineRunSucceed(pipelineRun.Name), "PipelineRunCompleted"); err != nil {
		t.Fatalf("Waiting for PipelineRun %s to succeed: %v", pipelineRun.Name, err)
	}

	pr, err := c.PipelineRunClient.Get(ctx, pipelineRun.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get PipelineRun %q: %v", pipelineRun.Name, err)
	}

	if len(pr.Status.ChildReferences) != 9 {
		t.Fatalf("Expected 9 child references, got %d", len(pr.Status.ChildReferences))
	}

	taskrunList, err := c.TaskRunClient.List(ctx, metav1.ListOptions{LabelSelector: "tekton.dev/pipelineRun=" + pipelineRun.Name})
	if err != nil {
		t.Fatalf("Error listing TaskRuns for PipelineRun %s: %s", pipelineRun.Name, err)
	}
	if len(taskrunList.Items) != 9 {
		t.Fatalf("Expected 9 TaskRuns for PipelineRun %s, got %d", pipelineRun.Name, len(taskrunList.Items))
	}
}

func TestPipelineWithMultipleMatrixedPipelineTasks(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c, namespace := setup(ctx, t, requireAlphaMinimalFeatureFlags)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, namespace) }, t.Logf)
	defer tearDown(ctx, t, c, namespace)

	task := parse.MustParseTask(t, fmt.Sprintf(`
metadata:
  name: mytask
  namespace: %s
spec:
  params:
    - name: platform
      type: string
    - name: browser
      type: string
  steps:
    - name: echo
      image: alpine
      script: |
        echo "$(params.platform) and $(params.browser)"
`, namespace))
	if _, err := c.TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create dag Task: %s", err)
	}

	pipeline := parse.MustParsePipeline(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  tasks:
  - name: pipelinetask1
    matrix:
    - name: platform
      value:
      - linux
      - mac
    - name: browser
      value:
      - chrome
    taskRef:
      name: %s
  - name: pipelinetask2
    matrix:
    - name: platform
      value:
      - windows
    - name: browser
      value:
      - safari
      - firefox
    taskRef:
      name: %s
`, helpers.ObjectNameForTest(t), namespace, task.Name, task.Name))
	if _, err := c.PipelineClient.Create(ctx, pipeline, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pipeline: %s", err)
	}

	pipelineRun := getPipelineRun(t, namespace, pipeline.Name)
	if _, err := c.PipelineRunClient.Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pipeline Run `%s`: %s", pipelineRun.Name, err)
	}

	if err := WaitForPipelineRunState(ctx, c, pipelineRun.Name, timeout, PipelineRunSucceed(pipelineRun.Name), "PipelineRunCompleted"); err != nil {
		t.Fatalf("Waiting for PipelineRun %s to succeed: %v", pipelineRun.Name, err)
	}

	pr, err := c.PipelineRunClient.Get(ctx, pipelineRun.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get PipelineRun %q: %v", pipelineRun.Name, err)
	}

	if len(pr.Status.ChildReferences) != 4 {
		t.Fatalf("Expected 4 child references, got %d", len(pr.Status.ChildReferences))
	}

	taskrunList, err := c.TaskRunClient.List(ctx, metav1.ListOptions{LabelSelector: "tekton.dev/pipelineRun=" + pipelineRun.Name})
	if err != nil {
		t.Fatalf("Error listing TaskRuns for PipelineRun %s: %s", pipelineRun.Name, err)
	}
	if len(taskrunList.Items) != 4 {
		t.Fatalf("Expected 4 TaskRuns for PipelineRun %s, got %d", pipelineRun.Name, len(taskrunList.Items))
	}
}

func TestPipelineWithMatrixedPipelineTasksInFinallyToo(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c, namespace := setup(ctx, t, requireAlphaMinimalFeatureFlags)
	knativetest.CleanupOnInterrupt(func() { tearDown(ctx, t, c, namespace) }, t.Logf)
	defer tearDown(ctx, t, c, namespace)

	task := parse.MustParseTask(t, fmt.Sprintf(`
metadata:
  name: mytask
  namespace: %s
spec:
  params:
    - name: platform
      type: string
    - name: browser
      type: string
  steps:
    - name: echo
      image: alpine
      script: |
        echo "$(params.platform) and $(params.browser)"
`, namespace))
	if _, err := c.TaskClient.Create(ctx, task, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create dag Task: %s", err)
	}

	pipeline := parse.MustParsePipeline(t, fmt.Sprintf(`
metadata:
  name: %s
  namespace: %s
spec:
  tasks:
  - name: pipelinetask1
    matrix:
    - name: platform
      value:
      - linux
      - mac
    - name: browser
      value:
      - chrome
    taskRef:
      name: %s
  finally:
  - name: pipelinetask2
    matrix:
    - name: platform
      value:
      - windows
    - name: browser
      value:
      - safari
      - firefox
    taskRef:
      name: %s
`, helpers.ObjectNameForTest(t), namespace, task.Name, task.Name))
	if _, err := c.PipelineClient.Create(ctx, pipeline, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pipeline: %s", err)
	}

	pipelineRun := getPipelineRun(t, namespace, pipeline.Name)
	if _, err := c.PipelineRunClient.Create(ctx, pipelineRun, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Failed to create Pipeline Run `%s`: %s", pipelineRun.Name, err)
	}

	if err := WaitForPipelineRunState(ctx, c, pipelineRun.Name, timeout, PipelineRunSucceed(pipelineRun.Name), "PipelineRunCompleted"); err != nil {
		t.Fatalf("Waiting for PipelineRun %s to succeed: %v", pipelineRun.Name, err)
	}

	pr, err := c.PipelineRunClient.Get(ctx, pipelineRun.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get PipelineRun %q: %v", pipelineRun.Name, err)
	}

	if len(pr.Status.ChildReferences) != 4 {
		t.Fatalf("Expected 4 child references, got %d", len(pr.Status.ChildReferences))
	}

	taskrunList, err := c.TaskRunClient.List(ctx, metav1.ListOptions{LabelSelector: "tekton.dev/pipelineRun=" + pipelineRun.Name})
	if err != nil {
		t.Fatalf("Error listing TaskRuns for PipelineRun %s: %s", pipelineRun.Name, err)
	}
	if len(taskrunList.Items) != 4 {
		t.Fatalf("Expected 4 TaskRuns for PipelineRun %s, got %d", pipelineRun.Name, len(taskrunList.Items))
	}
}
