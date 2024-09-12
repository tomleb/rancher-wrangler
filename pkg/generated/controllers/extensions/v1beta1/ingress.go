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

package v1beta1

import (
	"context"
	"sync"
	"time"

	"github.com/rancher/wrangler/v3/pkg/apply"
	"github.com/rancher/wrangler/v3/pkg/condition"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/rancher/wrangler/v3/pkg/kv"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IngressController interface for managing Ingress resources.
type IngressController interface {
	generic.ControllerInterface[*v1beta1.Ingress, *v1beta1.IngressList]
}

type IngressControllerContext interface {
	generic.ControllerInterfaceContext[*v1beta1.Ingress, *v1beta1.IngressList]
}

// IngressClient interface for managing Ingress resources in Kubernetes.
type IngressClient interface {
	generic.ClientInterface[*v1beta1.Ingress, *v1beta1.IngressList]
}

// IngressCache interface for retrieving Ingress resources in memory.
type IngressCache interface {
	generic.CacheInterface[*v1beta1.Ingress]
}

// IngressStatusHandler is executed for every added or modified Ingress. Should return the new status to be updated
type IngressStatusHandler func(obj *v1beta1.Ingress, status v1beta1.IngressStatus) (v1beta1.IngressStatus, error)

// IngressGeneratingHandler is the top-level handler that is executed for every Ingress event. It extends IngressStatusHandler by a returning a slice of child objects to be passed to apply.Apply
type IngressGeneratingHandler func(obj *v1beta1.Ingress, status v1beta1.IngressStatus) ([]runtime.Object, v1beta1.IngressStatus, error)

// RegisterIngressStatusHandler configures a IngressController to execute a IngressStatusHandler for every events observed.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterIngressStatusHandler(ctx context.Context, controller IngressController, condition condition.Cond, name string, handler IngressStatusHandler) {
	statusHandler := &ingressStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, generic.FromObjectHandlerToHandler(statusHandler.sync))
}

// RegisterIngressGeneratingHandler configures a IngressController to execute a IngressGeneratingHandler for every events observed, passing the returned objects to the provided apply.Apply.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterIngressGeneratingHandler(ctx context.Context, controller IngressController, apply apply.Apply,
	condition condition.Cond, name string, handler IngressGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &ingressGeneratingHandler{
		IngressGeneratingHandler: handler,
		apply:                    apply,
		name:                     name,
		gvk:                      controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterIngressStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type ingressStatusHandler struct {
	client    IngressClient
	condition condition.Cond
	handler   IngressStatusHandler
}

// sync is executed on every resource addition or modification. Executes the configured handlers and sends the updated status to the Kubernetes API
func (a *ingressStatusHandler) sync(key string, obj *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type ingressGeneratingHandler struct {
	IngressGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
	seen  sync.Map
}

// Remove handles the observed deletion of a resource, cascade deleting every associated resource previously applied
func (a *ingressGeneratingHandler) Remove(key string, obj *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1beta1.Ingress{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	if a.opts.UniqueApplyForResourceVersion {
		a.seen.Delete(key)
	}

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

// Handle executes the configured IngressGeneratingHandler and pass the resulting objects to apply.Apply, finally returning the new status of the resource
func (a *ingressGeneratingHandler) Handle(obj *v1beta1.Ingress, status v1beta1.IngressStatus) (v1beta1.IngressStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.IngressGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}
	if !a.isNewResourceVersion(obj) {
		return newStatus, nil
	}

	err = generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
	if err != nil {
		return newStatus, err
	}
	a.storeResourceVersion(obj)
	return newStatus, nil
}

// isNewResourceVersion detects if a specific resource version was already successfully processed.
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *ingressGeneratingHandler) isNewResourceVersion(obj *v1beta1.Ingress) bool {
	if !a.opts.UniqueApplyForResourceVersion {
		return true
	}

	// Apply once per resource version
	key := obj.Namespace + "/" + obj.Name
	previous, ok := a.seen.Load(key)
	return !ok || previous != obj.ResourceVersion
}

// storeResourceVersion keeps track of the latest resource version of an object for which Apply was executed
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *ingressGeneratingHandler) storeResourceVersion(obj *v1beta1.Ingress) {
	if !a.opts.UniqueApplyForResourceVersion {
		return
	}

	key := obj.Namespace + "/" + obj.Name
	a.seen.Store(key, obj.ResourceVersion)
}
