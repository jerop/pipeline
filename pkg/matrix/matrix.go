/*
Copyright 2022 The Tekton Authors

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

package matrix

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelinerunresources "github.com/tektoncd/pipeline/pkg/reconciler/pipelinerun/resources"
	taskrunresources "github.com/tektoncd/pipeline/pkg/reconciler/taskrun/resources"
	"strconv"
)

func UpdateState(fannedOutRprts pipelinerunresources.PipelineRunState, facts *pipelinerunresources.PipelineRunFacts) pipelinerunresources.PipelineRunState {
	var newState pipelinerunresources.PipelineRunState
	fannedOutRprtsMap := fannedOutRprts.ToMap()
	for _, rprt := range facts.State {
		if _, exists := fannedOutRprtsMap[rprt.PipelineTask.Name]; !exists {
			newState = append(newState, rprt)
		}
	}
	return append(newState, fannedOutRprts...)
}

func FanOut(rprts pipelinerunresources.PipelineRunState) (pipelinerunresources.PipelineRunState, []*v1beta1.MatrixPermutations) {
	var allFannedOutRprts pipelinerunresources.PipelineRunState
	var matricesPermutations []*v1beta1.MatrixPermutations

	for _, rprt := range rprts {
		fannedOutRprts, matrixPermutations := fanOutRprt(rprt)
		allFannedOutRprts = append(allFannedOutRprts, fannedOutRprts...)
		if matrixPermutations != nil {
			matricesPermutations = append(matricesPermutations, matrixPermutations)
		}
	}

	return allFannedOutRprts, matricesPermutations
}

func fanOutRprt(rprt *pipelinerunresources.ResolvedPipelineRunTask) (pipelinerunresources.PipelineRunState, *v1beta1.MatrixPermutations) {
	if len(rprt.PipelineTask.Matrix) == 0 {
		return pipelinerunresources.PipelineRunState{rprt}, nil
	}

	var rprts pipelinerunresources.PipelineRunState
	permutations := generatePermutations(rprt.PipelineTask.Matrix)
	for _, permutation := range permutations {
		rprts = append(rprts, createRprt(rprt, permutation))
	}

	return rprts, &v1beta1.MatrixPermutations{
		PipelineTaskName: rprt.PipelineTask.Name,
		Permutations:     permutations,
	}
}

func createRprt(rprt *pipelinerunresources.ResolvedPipelineRunTask, permutation *v1beta1.MatrixPermutation) *pipelinerunresources.ResolvedPipelineRunTask {
	// CustomTask, RunName, Run not supported
	// ResolvedConditionChecks not supported
	var newRprt pipelinerunresources.ResolvedPipelineRunTask
	if rprt.PipelineTask != nil {
		newRprt.PipelineTask = rprt.PipelineTask.DeepCopy()
	}
	if rprt.ResolvedTaskResources != nil {
		// Inputs and Outputs not supported
		newRprt.ResolvedTaskResources = &taskrunresources.ResolvedTaskResources{}
		newRprt.ResolvedTaskResources.TaskName = rprt.ResolvedTaskResources.TaskName
		newRprt.ResolvedTaskResources.Kind = rprt.ResolvedTaskResources.Kind
		if rprt.ResolvedTaskResources.TaskSpec != nil {
			newRprt.ResolvedTaskResources.TaskSpec = rprt.ResolvedTaskResources.TaskSpec.DeepCopy()
		}
	}
	newRprt.TaskRunName = rprt.TaskRunName + "-" + permutation.PermutationId
	newRprt.PipelineTask.Params = append(newRprt.PipelineTask.Params, permutation.Params...)
	newRprt.PipelineTask.Matrix = nil
	return &newRprt
}

func generatePermutations(params []v1beta1.Param) []*v1beta1.MatrixPermutation {
	var perms []*v1beta1.MatrixPermutation
	for _, param := range params {
		perms = distributeParam(param, perms)
	}
	return perms
}

func distributeParam(param v1beta1.Param, existingPermutations []*v1beta1.MatrixPermutation) []*v1beta1.MatrixPermutation {
	var expandedPermutations []*v1beta1.MatrixPermutation

	name := param.Name
	values := param.Value.ArrayVal

	if len(existingPermutations) == 0 {
		for i, value := range values {
			expandedPermutations = append(expandedPermutations, createPermutation(i, name, value, []v1beta1.Param{}))
		}
	} else {
		count := 0
		for _, value := range values {
			for _, perm := range existingPermutations {
				expandedPermutations = append(expandedPermutations, createPermutation(count, name, value, perm.Params))
				count += 1
			}
		}
	}
	return expandedPermutations
}

func createPermutation(i int, name string, value string, params []v1beta1.Param) *v1beta1.MatrixPermutation {
	return &v1beta1.MatrixPermutation{
		PermutationId: strconv.Itoa(i),
		Params: append(params, v1beta1.Param{
			Name:  name,
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: value},
		}),
	}
}
