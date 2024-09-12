package generic

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	finalizerKey = "wrangler.cattle.io/"
)

type Updater func(runtime.Object) (runtime.Object, error)

type objectLifecycleAdapter struct {
	name    string
	handler HandlerContext
	updater Updater
}

func NewRemoveHandler(name string, updater Updater, handler Handler) Handler {
	handlerCtx := func(_ context.Context, key string, obj runtime.Object) (runtime.Object, error) {
		return handler(key, obj)
	}
	o := objectLifecycleAdapter{
		name:    name,
		handler: handlerCtx,
		updater: updater,
	}
	return o.sync
}

func NewRemoveHandlerContext(name string, updater Updater, handler HandlerContext) HandlerContext {
	o := objectLifecycleAdapter{
		name:    name,
		handler: handler,
		updater: updater,
	}
	return o.syncContext
}

func (o *objectLifecycleAdapter) sync(key string, obj runtime.Object) (runtime.Object, error) {
	return o.syncContext(context.Background(), key, obj)
}

func (o *objectLifecycleAdapter) syncContext(ctx context.Context, key string, obj runtime.Object) (runtime.Object, error) {
	if obj == nil {
		return nil, nil
	}

	metadata, err := meta.Accessor(obj)
	if err != nil {
		return obj, err
	}

	if metadata.GetDeletionTimestamp() == nil {
		return o.addFinalizer(obj)
	}

	if !o.hasFinalizer(obj) {
		return obj, nil
	}

	newObj, err := o.handler(ctx, key, obj)
	if err != nil {
		return newObj, err
	}

	if newObj != nil {
		obj = newObj
	}

	return o.removeFinalizer(obj)
}

func (o *objectLifecycleAdapter) constructFinalizerKey() string {
	return finalizerKey + o.name
}

func (o *objectLifecycleAdapter) hasFinalizer(obj runtime.Object) bool {
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return false
	}

	finalizerKey := o.constructFinalizerKey()
	finalizers := metadata.GetFinalizers()
	for _, finalizer := range finalizers {
		if finalizer == finalizerKey {
			return true
		}
	}

	return false
}

func (o *objectLifecycleAdapter) removeFinalizer(obj runtime.Object) (runtime.Object, error) {
	if !o.hasFinalizer(obj) {
		return obj, nil
	}

	obj = obj.DeepCopyObject()
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return obj, err
	}

	finalizerKey := o.constructFinalizerKey()
	finalizers := metadata.GetFinalizers()

	var newFinalizers []string
	for k, v := range finalizers {
		if v != finalizerKey {
			continue
		}
		newFinalizers = append(finalizers[:k], finalizers[k+1:]...)
	}

	metadata.SetFinalizers(newFinalizers)
	return o.updater(obj)
}

func (o *objectLifecycleAdapter) addFinalizer(obj runtime.Object) (runtime.Object, error) {
	if o.hasFinalizer(obj) {
		return obj, nil
	}

	obj = obj.DeepCopyObject()
	metadata, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	metadata.SetFinalizers(append(metadata.GetFinalizers(), o.constructFinalizerKey()))
	return o.updater(obj)
}
