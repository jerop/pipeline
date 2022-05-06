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

func UpdatePipelineTasksList(fannedOutPipelineTasksMap map[string]v1beta1.PipelineTaskList, pipelineTaskList v1beta1.PipelineTaskList) v1beta1.PipelineTaskList {
	var newPipelineTasks v1beta1.PipelineTaskList
	for _, pipelineTask := range pipelineTaskList {
		if fannedOutPipelineTasks, exists := fannedOutPipelineTasksMap[pipelineTask.Name]; !exists {
			newPipelineTasks = append(newPipelineTasks, pipelineTask)
		} else {
			newPipelineTasks = append(newPipelineTasks, fannedOutPipelineTasks...)
		}
	}
	return newPipelineTasks
}

func UpdateState(fannedOutRprts pipelinerunresources.PipelineRunState, state pipelinerunresources.PipelineRunState) pipelinerunresources.PipelineRunState {
	var newState pipelinerunresources.PipelineRunState
	fannedOutRprtsMap := fannedOutRprts.ToMap()
	for _, rprt := range state {
		if _, exists := fannedOutRprtsMap[rprt.PipelineTask.Name]; !exists {
			newState = append(newState, rprt)
		}
	}
	return append(newState, fannedOutRprts...)
}

func FanOut(rprts pipelinerunresources.PipelineRunState) (pipelinerunresources.PipelineRunState, map[string]v1beta1.PipelineTaskList) {
	var allFannedOutRprts pipelinerunresources.PipelineRunState
	allFannedOutPipelineTasks := map[string]v1beta1.PipelineTaskList{}

	for _, rprt := range rprts {
		fannedOutRprts, combinations := fanOutRprt(rprt)
		allFannedOutRprts = append(allFannedOutRprts, fannedOutRprts...)

		fannedOutPts := fanOutPipelineTask(rprt.PipelineTask, combinations.Combinations)
		allFannedOutPipelineTasks[rprt.PipelineTask.Name] = fannedOutPts
	}

	return allFannedOutRprts, allFannedOutPipelineTasks
}

func fanOutRprt(rprt *pipelinerunresources.ResolvedPipelineRunTask) (pipelinerunresources.PipelineRunState, *PipelineTaskCombinations) {
	if len(rprt.PipelineTask.Matrix) == 0 {
		return pipelinerunresources.PipelineRunState{rprt}, nil
	}

	var rprts pipelinerunresources.PipelineRunState
	combinations := generateCombinations(rprt.PipelineTask.Matrix)
	for _, combination := range combinations {
		rprts = append(rprts, createRprt(rprt, combination))
	}

	return rprts, &PipelineTaskCombinations{
		PipelineTaskName: rprt.PipelineTask.Name,
		Combinations:     combinations,
	}
}

func createRprt(rprt *pipelinerunresources.ResolvedPipelineRunTask, combination *Combination) *pipelinerunresources.ResolvedPipelineRunTask {
	// Note: conditions, inputs and output fields are not supported because they are deprecated
	var newRprt pipelinerunresources.ResolvedPipelineRunTask
	if rprt.PipelineTask != nil {
		newRprt.PipelineTask = rprt.PipelineTask.DeepCopy()
		// newRprt.PipelineTask.Name = createName(newRprt.PipelineTask.Name, combination.MatrixId)
	}
	if rprt.ResolvedTaskResources != nil {
		newRprt.ResolvedTaskResources = &taskrunresources.ResolvedTaskResources{}
		newRprt.ResolvedTaskResources.TaskName = rprt.ResolvedTaskResources.TaskName
		newRprt.ResolvedTaskResources.Kind = rprt.ResolvedTaskResources.Kind
		if rprt.ResolvedTaskResources.TaskSpec != nil {
			newRprt.ResolvedTaskResources.TaskSpec = rprt.ResolvedTaskResources.TaskSpec.DeepCopy()
		}
	}
	if rprt.CustomTask {
		newRprt.RunName = createName(rprt.RunName, combination.MatrixId)
	} else {
		newRprt.TaskRunName = createName(rprt.TaskRunName, combination.MatrixId)
	}
	newRprt.PipelineTask.Params = append(newRprt.PipelineTask.Params, combination.Params...)
	newRprt.PipelineTask.Matrix = nil
	newRprt.MatrixId = combination.MatrixId
	return &newRprt
}

func generateCombinations(params []v1beta1.Param) []*Combination {
	var combinations []*Combination
	for _, parameter := range params {
		combinations = distributeParameter(parameter, combinations)
	}
	return combinations
}

func distributeParameter(param v1beta1.Param, existingCombinations []*Combination) []*Combination {
	var expandedCombinations []*Combination
	if len(existingCombinations) == 0 {
		for i, value := range param.Value.ArrayVal {
			expandedCombinations = append(expandedCombinations, createCombination(i, param.Name, value, []v1beta1.Param{}))
		}
	} else {
		count := 0
		for _, value := range param.Value.ArrayVal {
			for _, perm := range existingCombinations {
				expandedCombinations = append(expandedCombinations, createCombination(count, param.Name, value, perm.Params))
				count += 1
			}
		}
	}
	return expandedCombinations
}

func createCombination(i int, name string, value string, parameters []v1beta1.Param) *Combination {
	return &Combination{
		MatrixId: strconv.Itoa(i),
		Params: append(parameters, v1beta1.Param{
			Name:  name,
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: value},
		}),
	}
}

func createName(name string, matrixId string) string {
	return name + "-" + matrixId
}

func createPipelineTask(pt *v1beta1.PipelineTask, combination *Combination) *v1beta1.PipelineTask {
	newPipelineTask := pt.DeepCopy()
	newPipelineTask.Params = append(newPipelineTask.Params, combination.Params...)
	newPipelineTask.Matrix = nil
	newPipelineTask.Name = createName(pt.Name, combination.MatrixId)
	return newPipelineTask
}

func fanOutPipelineTask(pt *v1beta1.PipelineTask, combinations []*Combination) v1beta1.PipelineTaskList {
	if len(pt.Matrix) == 0 {
		return v1beta1.PipelineTaskList{*pt}
	}

	var ptl v1beta1.PipelineTaskList
	for _, combination := range combinations {
		ptl = append(ptl, *createPipelineTask(pt, combination))
	}
	return ptl
}
