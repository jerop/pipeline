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

package taskrun

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/list"
	"github.com/tektoncd/pipeline/pkg/reconciler/taskrun/resources"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ValidateResolvedTask validates that all parameters declared in the TaskSpec are present in the taskrun
// It also validates that all parameters have values, parameter types match the specified type and
// object params have all the keys required
func ValidateResolvedTask(params v1beta1.Params, matrixParams v1beta1.Params, rtr *resources.ResolvedTask) error {
	if rtr != nil && rtr.TaskSpec != nil {
		if err := rtr.TaskSpec.Params.ValidateParams(params, matrixParams); err != nil {
			return fmt.Errorf("invalid input params for task %s: %w", rtr.TaskName, err)
		}
	}
	return nil
}

func validateTaskSpecRequestResources(taskSpec *v1beta1.TaskSpec) error {
	if taskSpec != nil {
		for _, step := range taskSpec.Steps {
			for k, request := range step.Resources.Requests {
				// First validate the limit in step
				if limit, ok := step.Resources.Limits[k]; ok {
					if (&limit).Cmp(request) == -1 {
						return fmt.Errorf("Invalid request resource value: %v must be less or equal to limit %v", request.String(), limit.String())
					}
				} else if taskSpec.StepTemplate != nil {
					// If step doesn't configure the limit, validate the limit in stepTemplate
					if limit, ok := taskSpec.StepTemplate.Resources.Limits[k]; ok {
						if (&limit).Cmp(request) == -1 {
							return fmt.Errorf("Invalid request resource value: %v must be less or equal to limit %v", request.String(), limit.String())
						}
					}
				}
			}
		}
	}

	return nil
}

// validateOverrides validates that all stepOverrides map to valid steps, and likewise for sidecarOverrides
func validateOverrides(ts *v1beta1.TaskSpec, trs *v1beta1.TaskRunSpec) error {
	stepErr := validateStepOverrides(ts, trs)
	sidecarErr := validateSidecarOverrides(ts, trs)
	return multierror.Append(stepErr, sidecarErr).ErrorOrNil()
}

func validateStepOverrides(ts *v1beta1.TaskSpec, trs *v1beta1.TaskRunSpec) error {
	var err error
	stepNames := sets.NewString()
	for _, step := range ts.Steps {
		stepNames.Insert(step.Name)
	}
	for _, stepOverride := range trs.StepOverrides {
		if !stepNames.Has(stepOverride.Name) {
			err = multierror.Append(err, fmt.Errorf("invalid StepOverride: No Step named %s", stepOverride.Name))
		}
	}
	return err
}

func validateSidecarOverrides(ts *v1beta1.TaskSpec, trs *v1beta1.TaskRunSpec) error {
	var err error
	sidecarNames := sets.NewString()
	for _, sidecar := range ts.Sidecars {
		sidecarNames.Insert(sidecar.Name)
	}
	for _, sidecarOverride := range trs.SidecarOverrides {
		if !sidecarNames.Has(sidecarOverride.Name) {
			err = multierror.Append(err, fmt.Errorf("invalid SidecarOverride: No Sidecar named %s", sidecarOverride.Name))
		}
	}
	return err
}

// validateResults checks the emitted results type and object properties against the ones defined in spec.
func validateTaskRunResults(tr *v1beta1.TaskRun, resolvedTaskSpec *v1beta1.TaskSpec) error {
	specResults := []v1beta1.TaskResult{}
	if tr.Spec.TaskSpec != nil {
		specResults = append(specResults, tr.Spec.TaskSpec.Results...)
	}

	if resolvedTaskSpec != nil {
		specResults = append(specResults, resolvedTaskSpec.Results...)
	}

	// When get the results, check if the type of result is the expected one
	if missmatchedTypes := mismatchedTypesResults(tr, specResults); len(missmatchedTypes) != 0 {
		var s []string
		for k, v := range missmatchedTypes {
			s = append(s, fmt.Sprintf(" \"%v\": %v", k, v))
		}
		sort.Strings(s)
		return fmt.Errorf("Provided results don't match declared results; may be invalid JSON or missing result declaration: %v", strings.Join(s, ","))
	}

	// When get the results, for object value need to check if they have missing keys.
	if missingKeysObjectNames := missingKeysofObjectResults(tr, specResults); len(missingKeysObjectNames) != 0 {
		return fmt.Errorf("missing keys for these results which are required in TaskResult's properties %v", missingKeysObjectNames)
	}
	return nil
}

// mismatchedTypesResults checks and returns all the mismatched types of emitted results against specified results.
func mismatchedTypesResults(tr *v1beta1.TaskRun, specResults []v1beta1.TaskResult) map[string]string {
	neededTypes := make(map[string]string)
	mismatchedTypes := make(map[string]string)
	var filteredResults []v1beta1.TaskRunResult
	// collect needed types for results
	for _, r := range specResults {
		neededTypes[r.Name] = string(r.Type)
	}

	// collect mismatched types for results, and correct results in filteredResults
	// TODO(#6097): Validate if the emitted results are defined in taskspec
	for _, trr := range tr.Status.TaskRunResults {
		needed, ok := neededTypes[trr.Name]
		if ok && needed != string(trr.Type) {
			mismatchedTypes[trr.Name] = fmt.Sprintf("task result is expected to be \"%v\" type but was initialized to a different type \"%v\"", needed, trr.Type)
		} else {
			filteredResults = append(filteredResults, trr)
		}
	}
	// remove the mismatched results
	tr.Status.TaskRunResults = filteredResults
	return mismatchedTypes
}

// missingKeysofObjectResults checks and returns the missing keys of object results.
func missingKeysofObjectResults(tr *v1beta1.TaskRun, specResults []v1beta1.TaskResult) map[string][]string {
	neededKeys := make(map[string][]string)
	providedKeys := make(map[string][]string)
	// collect needed keys for object results
	for _, r := range specResults {
		if string(r.Type) == string(v1beta1.ParamTypeObject) {
			for key := range r.Properties {
				neededKeys[r.Name] = append(neededKeys[r.Name], key)
			}
		}
	}

	// collect provided keys for object results
	for _, trr := range tr.Status.TaskRunResults {
		if trr.Value.Type == v1beta1.ParamTypeObject {
			for key := range trr.Value.ObjectVal {
				providedKeys[trr.Name] = append(providedKeys[trr.Name], key)
			}
		}
	}
	return findMissingKeys(neededKeys, providedKeys)
}

// findMissingKeys checks if objects have missing keys in its providers (taskrun value and default)
func findMissingKeys(neededKeys, providedKeys map[string][]string) map[string][]string {
	missings := map[string][]string{}
	for p, keys := range providedKeys {
		if _, ok := neededKeys[p]; !ok {
			// Ignore any missing objects - this happens when object param is provided with default
			continue
		}
		missedKeys := list.DiffLeft(neededKeys[p], keys)
		if len(missedKeys) != 0 {
			missings[p] = missedKeys
		}
	}

	return missings
}
