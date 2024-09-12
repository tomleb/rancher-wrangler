/*
Copyright The Kubernetes Authors.

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

// Code generated by main. DO NOT EDIT.

package v1

import (
	"github.com/rancher/wrangler/v3/pkg/generic"
	v1 "k8s.io/api/networking/v1"
)

// NetworkPolicyController interface for managing NetworkPolicy resources.
type NetworkPolicyController interface {
	generic.ControllerInterface[*v1.NetworkPolicy, *v1.NetworkPolicyList]
}

type NetworkPolicyControllerContext interface {
	generic.ControllerInterfaceContext[*v1.NetworkPolicy, *v1.NetworkPolicyList]
}

// NetworkPolicyClient interface for managing NetworkPolicy resources in Kubernetes.
type NetworkPolicyClient interface {
	generic.ClientInterface[*v1.NetworkPolicy, *v1.NetworkPolicyList]
}

// NetworkPolicyCache interface for retrieving NetworkPolicy resources in memory.
type NetworkPolicyCache interface {
	generic.CacheInterface[*v1.NetworkPolicy]
}
