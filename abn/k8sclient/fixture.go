package k8sclient

// Mostly copied from https://github.com/kubernetes/client-go/blob/e7cd4ba474b5efc2882e377362c9aa8b407428d9/testing/fixture.go
// Changes were made in Delete() method to handle finalizers; that is, to change delete to update if a finalizer exists
// Finalizers are not handled in original

import (
	"fmt"
	"sort"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	ktesting "k8s.io/client-go/testing"
)

type tracker struct {
	scheme  ktesting.ObjectScheme
	decoder runtime.Decoder
	lock    sync.RWMutex
	objects map[schema.GroupVersionResource]map[types.NamespacedName]runtime.Object
	// The value type of watchers is a map of which the key is either a namespace or
	// all/non namespace aka "" and its value is list of fake watchers.
	// Manipulations on resources will broadcast the notification events into the
	// watchers' channel. Note that too many unhandled events (currently 100,
	// see apimachinery/pkg/watch.DefaultChanSize) will cause a panic.
	watchers map[schema.GroupVersionResource]map[string][]*watch.RaceFreeFakeWatcher
}

var _ ktesting.ObjectTracker = &tracker{}

// NewObjectTracker returns an ObjectTracker that can be used to keep track
// of objects for the fake clientset. Mostly useful for unit tests.
func NewObjectTracker(scheme ktesting.ObjectScheme, decoder runtime.Decoder) ktesting.ObjectTracker {
	return &tracker{
		scheme:   scheme,
		decoder:  decoder,
		objects:  make(map[schema.GroupVersionResource]map[types.NamespacedName]runtime.Object),
		watchers: make(map[schema.GroupVersionResource]map[string][]*watch.RaceFreeFakeWatcher),
	}
}

func (t *tracker) List(gvr schema.GroupVersionResource, gvk schema.GroupVersionKind, ns string) (runtime.Object, error) {
	// Heuristic for list kind: original kind + List suffix. Might
	// not always be true but this tracker has a pretty limited
	// understanding of the actual API model.
	listGVK := gvk
	listGVK.Kind = listGVK.Kind + "List"
	// GVK does have the concept of "internal version". The scheme recognizes
	// the runtime.APIVersionInternal, but not the empty string.
	if listGVK.Version == "" {
		listGVK.Version = runtime.APIVersionInternal
	}

	list, err := t.scheme.New(listGVK)
	if err != nil {
		return nil, err
	}

	if !meta.IsListType(list) {
		return nil, fmt.Errorf("%q is not a list type", listGVK.Kind)
	}

	t.lock.RLock()
	defer t.lock.RUnlock()

	objs, ok := t.objects[gvr]
	if !ok {
		return list, nil
	}

	matchingObjs, err := filterByNamespace(objs, ns)
	if err != nil {
		return nil, err
	}
	if err := meta.SetList(list, matchingObjs); err != nil {
		return nil, err
	}
	return list.DeepCopyObject(), nil
}

func (t *tracker) Watch(gvr schema.GroupVersionResource, ns string) (watch.Interface, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	fakewatcher := watch.NewRaceFreeFake()

	if _, exists := t.watchers[gvr]; !exists {
		t.watchers[gvr] = make(map[string][]*watch.RaceFreeFakeWatcher)
	}
	t.watchers[gvr][ns] = append(t.watchers[gvr][ns], fakewatcher)
	return fakewatcher, nil
}

func (t *tracker) Get(gvr schema.GroupVersionResource, ns, name string) (runtime.Object, error) {
	errNotFound := errors.NewNotFound(gvr.GroupResource(), name)

	t.lock.RLock()
	defer t.lock.RUnlock()

	objs, ok := t.objects[gvr]
	if !ok {
		return nil, errNotFound
	}

	matchingObj, ok := objs[types.NamespacedName{Namespace: ns, Name: name}]
	if !ok {
		return nil, errNotFound
	}

	// Only one object should match in the tracker if it works
	// correctly, as Add/Update methods enforce kind/namespace/name
	// uniqueness.
	obj := matchingObj.DeepCopyObject()
	if status, ok := obj.(*metav1.Status); ok {
		if status.Status != metav1.StatusSuccess {
			return nil, &errors.StatusError{ErrStatus: *status}
		}
	}

	return obj, nil
}

func (t *tracker) Add(obj runtime.Object) error {
	if meta.IsListType(obj) {
		return t.addList(obj, false)
	}
	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	gvks, _, err := t.scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}

	if partial, ok := obj.(*metav1.PartialObjectMetadata); ok && len(partial.TypeMeta.APIVersion) > 0 {
		gvks = []schema.GroupVersionKind{partial.TypeMeta.GroupVersionKind()}
	}

	if len(gvks) == 0 {
		return fmt.Errorf("no registered kinds for %v", obj)
	}
	for _, gvk := range gvks {
		// NOTE: UnsafeGuessKindToResource is a heuristic and default match. The
		// actual registration in apiserver can specify arbitrary route for a
		// gvk. If a test uses such objects, it cannot preset the tracker with
		// objects via Add(). Instead, it should trigger the Create() function
		// of the tracker, where an arbitrary gvr can be specified.
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		// Resource doesn't have the concept of "__internal" version, just set it to "".
		if gvr.Version == runtime.APIVersionInternal {
			gvr.Version = ""
		}

		err := t.add(gvr, obj, objMeta.GetNamespace(), false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *tracker) Create(gvr schema.GroupVersionResource, obj runtime.Object, ns string) error {
	return t.add(gvr, obj, ns, false)
}

func (t *tracker) Update(gvr schema.GroupVersionResource, obj runtime.Object, ns string) error {
	return t.add(gvr, obj, ns, true)
}

func (t *tracker) getWatches(gvr schema.GroupVersionResource, ns string) []*watch.RaceFreeFakeWatcher {
	watches := []*watch.RaceFreeFakeWatcher{}
	if t.watchers[gvr] != nil {
		if w := t.watchers[gvr][ns]; w != nil {
			watches = append(watches, w...)
		}
		if ns != metav1.NamespaceAll {
			if w := t.watchers[gvr][metav1.NamespaceAll]; w != nil {
				watches = append(watches, w...)
			}
		}
	}
	return watches
}

func (t *tracker) add(gvr schema.GroupVersionResource, obj runtime.Object, ns string, replaceExisting bool) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	gr := gvr.GroupResource()

	// To avoid the object from being accidentally modified by caller
	// after it's been added to the tracker, we always store the deep
	// copy.
	obj = obj.DeepCopyObject()

	newMeta, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	// Propagate namespace to the new object if hasn't already been set.
	if len(newMeta.GetNamespace()) == 0 {
		newMeta.SetNamespace(ns)
	}

	if ns != newMeta.GetNamespace() {
		msg := fmt.Sprintf("request namespace does not match object namespace, request: %q object: %q", ns, newMeta.GetNamespace())
		return errors.NewBadRequest(msg)
	}

	_, ok := t.objects[gvr]
	if !ok {
		t.objects[gvr] = make(map[types.NamespacedName]runtime.Object)
	}

	namespacedName := types.NamespacedName{Namespace: newMeta.GetNamespace(), Name: newMeta.GetName()}
	if _, ok = t.objects[gvr][namespacedName]; ok {
		if replaceExisting {
			for _, w := range t.getWatches(gvr, ns) {
				// To avoid the object from being accidentally modified by watcher
				w.Modify(obj.DeepCopyObject())
			}
			t.objects[gvr][namespacedName] = obj
			return nil
		}
		return errors.NewAlreadyExists(gr, newMeta.GetName())
	}

	if replaceExisting {
		// Tried to update but no matching object was found.
		return errors.NewNotFound(gr, newMeta.GetName())
	}

	t.objects[gvr][namespacedName] = obj

	for _, w := range t.getWatches(gvr, ns) {
		// To avoid the object from being accidentally modified by watcher
		w.Add(obj.DeepCopyObject())
	}

	return nil
}

func (t *tracker) addList(obj runtime.Object, replaceExisting bool) error {
	list, err := meta.ExtractList(obj)
	if err != nil {
		return err
	}
	errs := runtime.DecodeList(list, t.decoder)
	if len(errs) > 0 {
		return errs[0]
	}
	for _, obj := range list {
		if err := t.Add(obj); err != nil {
			return err
		}
	}
	return nil
}

func (t *tracker) Delete(gvr schema.GroupVersionResource, ns, name string) error {
	t.lock.Lock()
	// defer t.lock.Unlock()

	objs, ok := t.objects[gvr]
	if !ok {
		t.lock.Unlock()
		return errors.NewNotFound(gvr.GroupResource(), name)
	}

	namespacedName := types.NamespacedName{Namespace: ns, Name: name}
	obj, ok := objs[namespacedName]
	if !ok {
		t.lock.Unlock()
		return errors.NewNotFound(gvr.GroupResource(), name)
	}

	// additions for finalizers
	// if finalizers, set DeletionTimestamp and convert to update, else delete
	uObj := obj.(*unstructured.Unstructured)
	if len(uObj.GetFinalizers()) > 0 {
		now := metav1.Now()
		uObj.SetDeletionTimestamp(&now)
		t.lock.Unlock()
		return t.Update(gvr, uObj, ns)
	}

	delete(objs, namespacedName)
	for _, w := range t.getWatches(gvr, ns) {
		w.Delete(obj.DeepCopyObject())
	}
	t.lock.Unlock()
	return nil
}

// filterByNamespace returns all objects in the collection that
// match provided namespace. Empty namespace matches
// non-namespaced objects.
func filterByNamespace(objs map[types.NamespacedName]runtime.Object, ns string) ([]runtime.Object, error) {
	var res []runtime.Object

	for _, obj := range objs {
		acc, err := meta.Accessor(obj)
		if err != nil {
			return nil, err
		}
		if ns != "" && acc.GetNamespace() != ns {
			continue
		}
		res = append(res, obj)
	}

	// Sort res to get deterministic order.
	sort.Slice(res, func(i, j int) bool {
		acc1, _ := meta.Accessor(res[i])
		acc2, _ := meta.Accessor(res[j])
		if acc1.GetNamespace() != acc2.GetNamespace() {
			return acc1.GetNamespace() < acc2.GetNamespace()
		}
		return acc1.GetName() < acc2.GetName()
	})
	return res, nil
}
