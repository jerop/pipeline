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
	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/reconciler/pipelinerun/resources"
	"testing"
)

func Test_generateCombinations(t *testing.T) {
	tests := []struct {
		name             string
		matrix           []v1beta1.Param
		wantCombinations []*Combination
	}{{
		name: "one array",
		matrix: []v1beta1.Param{{
			Name:  "platforms",
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
		}},
		wantCombinations: []*Combination{{
			MatrixId: "0",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}},
		}, {
			MatrixId: "1",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}},
		}, {
			MatrixId: "2",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}},
		}},
	}, {
		name: "two arrays",
		matrix: []v1beta1.Param{{
			Name:  "platforms",
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
		}, {
			Name:  "browsers",
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"chrome", "safari", "firefox"}},
		}},
		wantCombinations: []*Combination{{
			MatrixId: "0",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixId: "1",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixId: "2",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixId: "3",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			MatrixId: "4",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			MatrixId: "5",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			MatrixId: "6",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			MatrixId: "7",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			MatrixId: "8",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := generateCombinations(tt.matrix)
			if d := cmp.Diff(tt.wantCombinations, gotCombinations); d != "" {
				t.Errorf("failed: %s", d)
			}
		})
	}
}

func TestFanOut(t *testing.T) {
	tests := []struct {
		name                 string
		state                resources.PipelineRunState
		wantPipelineRunState resources.PipelineRunState
		//pipelineTaskCombinations []*PipelineTaskCombinations
	}{{
		name: "simple",
		state: resources.PipelineRunState{{
			TaskRunName: "a-taskRun",
			PipelineTask: &v1beta1.PipelineTask{
				Name:    "a-task",
				TaskRef: &v1beta1.TaskRef{Name: "a-taskRef"},
				Params: []v1beta1.Param{{
					Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "a-value"}},
				},
				Matrix: []v1beta1.Param{{
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"north", "east", "south", "west"}}},
				},
			},
		}},
		wantPipelineRunState: resources.PipelineRunState{{
			TaskRunName: "a-taskRun-0",
			MatrixId:    "0",
			PipelineTask: &v1beta1.PipelineTask{
				Name:    "a-task",
				TaskRef: &v1beta1.TaskRef{Name: "a-taskRef"},
				Params: []v1beta1.Param{{
					Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "a-value"},
				}, {
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "north"},
				}},
			},
		}, {
			TaskRunName: "a-taskRun-1",
			MatrixId:    "1",
			PipelineTask: &v1beta1.PipelineTask{
				Name:    "a-task",
				TaskRef: &v1beta1.TaskRef{Name: "a-taskRef"},
				Params: []v1beta1.Param{{
					Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "a-value"},
				}, {
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "east"},
				}},
			},
		}, {
			TaskRunName: "a-taskRun-2",
			MatrixId:    "2",
			PipelineTask: &v1beta1.PipelineTask{
				Name:    "a-task",
				TaskRef: &v1beta1.TaskRef{Name: "a-taskRef"},
				Params: []v1beta1.Param{{
					Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "a-value"},
				}, {
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "south"},
				}},
			},
		}, {
			TaskRunName: "a-taskRun-3",
			MatrixId:    "3",
			PipelineTask: &v1beta1.PipelineTask{
				Name:    "a-task",
				TaskRef: &v1beta1.TaskRef{Name: "a-taskRef"},
				Params: []v1beta1.Param{{
					Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "a-value"},
				}, {
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "west"},
				}},
			},
		}},
		//pipelineTaskCombinations: []*PipelineTaskCombinations{{
		//	PipelineTaskName: "a-task",
		//	Combinations: []*Combination{{
		//		MatrixId: "0",
		//		Params: []v1beta1.Param{{
		//			Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "north"},
		//		}},
		//	}, {
		//		MatrixId: "1",
		//		Params: []v1beta1.Param{{
		//			Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "east"},
		//		}},
		//	}, {
		//		MatrixId: "2",
		//		Params: []v1beta1.Param{{
		//			Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "south"},
		//		}},
		//	}, {
		//		MatrixId: "3",
		//		Params: []v1beta1.Param{{
		//			Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "west"},
		//		}},
		//	}},
		//}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipelineRunState, _ := FanOut(tt.state)
			if d := cmp.Diff(tt.wantPipelineRunState, gotPipelineRunState); d != "" {
				t.Errorf("Did not get expected fanned out ResolvedPipelineRunTasks: %s", d)
			}
			//if d := cmp.Diff(tt.pipelineTaskCombinations, gotPipelineTaskCombinations); d != "" {
			//	t.Errorf("Did not get expected combinations of matrices from ResolvedPipelineRunTasks: %s", d)
			//}
		})
	}
}
