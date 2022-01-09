package matrix

import (
	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/reconciler/pipelinerun/resources"
	"testing"
)

func Test_generatePermutations(t *testing.T) {
	tests := []struct {
		name      string
		matrix    []v1beta1.Param
		wantPerms []*v1beta1.MatrixPermutation
	}{{
		name: "one array",
		matrix: []v1beta1.Param{{
			Name:  "platforms",
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
		}},
		wantPerms: []*v1beta1.MatrixPermutation{{
			PermutationId: "0",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}},
		}, {
			PermutationId: "1",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}},
		}, {
			PermutationId: "2",
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
		wantPerms: []*v1beta1.MatrixPermutation{{
			PermutationId: "0",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			PermutationId: "1",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			PermutationId: "2",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			PermutationId: "3",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			PermutationId: "4",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			PermutationId: "5",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			PermutationId: "6",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			PermutationId: "7",
			Params: []v1beta1.Param{{
				Name:  "platforms",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browsers",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			PermutationId: "8",
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
			gotPerms := generatePermutations(tt.matrix)
			if d := cmp.Diff(tt.wantPerms, gotPerms); d != "" {
				t.Errorf("failed: %s", d)
			}
		})
	}
}

func TestFanOut(t *testing.T) {
	tests := []struct {
		name                     string
		state                    resources.PipelineRunState
		wantPipelineRunState     resources.PipelineRunState
		wantMatricesPermutations []*v1beta1.MatrixPermutations
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
		wantMatricesPermutations: []*v1beta1.MatrixPermutations{{
			PipelineTaskName: "a-task",
			Permutations: []*v1beta1.MatrixPermutation{{
				PermutationId: "0",
				Params: []v1beta1.Param{{
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "north"},
				}},
			}, {
				PermutationId: "1",
				Params: []v1beta1.Param{{
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "east"},
				}},
			}, {
				PermutationId: "2",
				Params: []v1beta1.Param{{
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "south"},
				}},
			}, {
				PermutationId: "3",
				Params: []v1beta1.Param{{
					Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "west"},
				}},
			}},
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipelineRunState, gotMatricesPermutations := FanOut(tt.state)
			if d := cmp.Diff(tt.wantPipelineRunState, gotPipelineRunState); d != "" {
				t.Errorf("Did not get expected fanned out ResolvedPipelineRunTasks: %s", d)
			}
			if d := cmp.Diff(tt.wantMatricesPermutations, gotMatricesPermutations); d != "" {
				t.Errorf("Did not get expected permutations of matrices from ResolvedPipelineRunTasks: %s", d)
			}
		})
	}
}
