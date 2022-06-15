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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/test/diff"
)

func Test_FanOut(t *testing.T) {
	tests := []struct {
		name             string
		matrix           []v1beta1.Param
		wantCombinations Combinations
	}{{
		name: "single array in matrix",
		matrix: []v1beta1.Param{{
			Name:  "platform",
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
		}},
		wantCombinations: Combinations{{
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}},
		}, {
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}},
		}},
	}, {
		name: "multiple arrays in matrix",
		matrix: []v1beta1.Param{{
			Name:  "platform",
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"linux", "mac", "windows"}},
		}, {
			Name:  "browser",
			Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"chrome", "safari", "firefox"}},
		}},
		wantCombinations: Combinations{{
			MatrixID: "0",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixID: "1",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixID: "2",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "chrome"},
			}},
		}, {
			MatrixID: "3",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			MatrixID: "4",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			MatrixID: "5",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "safari"},
			}},
		}, {
			MatrixID: "6",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "linux"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			MatrixID: "7",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "mac"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}, {
			MatrixID: "8",
			Params: []v1beta1.Param{{
				Name:  "platform",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "windows"},
			}, {
				Name:  "browser",
				Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeString, StringVal: "firefox"},
			}},
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCombinations := FanOut(tt.matrix)
			if d := cmp.Diff(tt.wantCombinations, gotCombinations); d != "" {
				t.Errorf("Combinations of Parameters did not match the expected Combinations: %s", d)
			}
		})
	}
}

func TestPipelineTask_CountCombinations(t *testing.T) {
	tests := []struct {
		name                    string
		matrix                  []v1beta1.Param
		matrixCombinationsCount int
	}{{
		name:                    "combinations count is zero",
		matrix:                  []v1beta1.Param{},
		matrixCombinationsCount: 0,
	}, {
		name: "combinations count is one from one parameter",
		matrix: []v1beta1.Param{{
			Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"foo"}},
		}},
		matrixCombinationsCount: 1,
	}, {
		name: "combinations count is one from two parameters",
		matrix: []v1beta1.Param{{
			Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"foo"}},
		}, {
			Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"bar"}},
		}},
		matrixCombinationsCount: 1,
	}, {
		name: "combinations count is two from one parameter",
		matrix: []v1beta1.Param{{
			Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"foo", "bar"}},
		}},
		matrixCombinationsCount: 2,
	}, {
		name: "combinations count is nine",
		matrix: []v1beta1.Param{{
			Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"f", "o", "o"}},
		}, {
			Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"b", "a", "r"}},
		}},
		matrixCombinationsCount: 9,
	}, {
		name: "combinations count is large",
		matrix: []v1beta1.Param{{
			Name: "foo", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"f", "o", "o"}},
		}, {
			Name: "bar", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"b", "a", "r"}},
		}, {
			Name: "quz", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"q", "u", "z"}},
		}, {
			Name: "xyzzy", Value: v1beta1.ArrayOrString{Type: v1beta1.ParamTypeArray, ArrayVal: []string{"x", "y", "z", "z", "y"}},
		}},
		matrixCombinationsCount: 135,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if d := cmp.Diff(tt.matrixCombinationsCount, CountCombinations(tt.matrix)); d != "" {
				t.Errorf("CountCombinations() errors diff %s", diff.PrintWantGot(d))
			}
		})
	}
}
