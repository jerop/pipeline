//go:build examples
// +build examples

/*
Copyright 2021 The Tekton Authors

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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type pathFilter func(string) bool

const (
	systemNamespaceEnvVar  = "SYSTEM_NAMESPACE"
	defaultSystemNamespace = "tekton-pipelines"
)

// getPathFilter returns a pathFilter that filters out examples
// unsuitable for the current feature-gate. For example,
// if the enable-api-fields feature flag is currently set
// to "alpha" then all stable and alpha examples would be
// allowed. When the flag is set to "stable", only stable
// examples are allowed.
func getPathFilter(t *testing.T) (pathFilter, error) {
	ns := os.Getenv(systemNamespaceEnvVar)
	if ns == "" {
		ns = defaultSystemNamespace
	}
	enabledFeatureGate, embeddedStatus, err := getFeatureGates(ns)
	if err != nil {
		return nil, fmt.Errorf("error reading enabled feature gate: %v", err)
	}
	var f pathFilter
	switch enabledFeatureGate {
	case "stable":
		f = stablePathFilter
	case "alpha":
		switch embeddedStatus {
		case "minimal":
			f = alphaMinimalPathFilter
		default:
			f = alphaPathFilter
		}
	}
	if f == nil {
		return nil, fmt.Errorf("unable to create path filter from feature gate %q", enabledFeatureGate)
	}
	t.Logf("Allowing only %q examples with %q status enabled", enabledFeatureGate, embeddedStatus)
	return f, nil
}

// Memoize value of enable-api-fields flag so we don't
// need to repeatedly query the feature flag configmap
var enableAPIFields = ""

var embeddedStatus = ""

// getFeatureGates queries the tekton pipelines namespace for the
// current value of the "enable-api-fields" feature gate.
func getFeatureGates(namespace string) (string, string, error) {
	if enableAPIFields == "" {
		cmd := exec.Command("kubectl", "get", "configmap", "feature-flags", "-n", namespace, "-o", `jsonpath="{.data['enable-api-fields']}"`)
		output, err := cmd.Output()
		if err != nil {
			return "", "", fmt.Errorf("error getting feature-flags configmap: %v", err)
		}
		output = bytes.TrimSpace(output)
		output = bytes.Trim(output, "\"")
		if len(output) == 0 {
			output = []byte("stable")
		}
		enableAPIFields = string(output)
	}
	if embeddedStatus == "" {
		cmd := exec.Command("kubectl", "get", "configmap", "feature-flags", "-n", namespace, "-o", `jsonpath="{.data['embedded-status']}"`)
		output, err := cmd.Output()
		if err != nil {
			return "", "", fmt.Errorf("error getting feature-flags configmap: %v", err)
		}
		output = bytes.TrimSpace(output)
		output = bytes.Trim(output, "\"")
		if len(output) == 0 {
			output = []byte("full")
		}
		embeddedStatus = string(output)
	}
	return enableAPIFields, embeddedStatus, nil
}

// stablePathFilter returns true for any example that should be allowed to run
// when "enable-api-fields" is "stable".
func stablePathFilter(p string) bool {
	return !(strings.Contains(p, "/alpha/") || strings.Contains(p, "/beta/"))
}

// alphaPathFilter returns true for any example that should be allowed to run
// when "enable-api-fields" is "alpha".
func alphaPathFilter(p string) bool {
	return strings.Contains(p, "/alpha/") || strings.Contains(p, "/beta/") || stablePathFilter(p)
}

// alphaPathFilter returns true for any example that should be allowed to run
// when "enable-api-fields" is "alpha".
func alphaMinimalPathFilter(p string) bool {
	return strings.Contains(p, "/minimal/") || alphaPathFilter(p)
}
