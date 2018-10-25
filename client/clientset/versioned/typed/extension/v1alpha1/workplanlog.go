/*
Copyright 2018 The KubeCI Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	rest "k8s.io/client-go/rest"
)

// WorkplanLogsGetter has a method to return a WorkplanLogInterface.
// A group's client should implement this interface.
type WorkplanLogsGetter interface {
	WorkplanLogs(namespace string) WorkplanLogInterface
}

// WorkplanLogInterface has methods to work with WorkplanLog resources.
type WorkplanLogInterface interface {
	WorkplanLogExpansion
}

// workplanLogs implements WorkplanLogInterface
type workplanLogs struct {
	client rest.Interface
	ns     string
}

// newWorkplanLogs returns a WorkplanLogs
func newWorkplanLogs(c *ExtensionV1alpha1Client, namespace string) *workplanLogs {
	return &workplanLogs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}
