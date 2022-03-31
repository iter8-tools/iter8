package base

import (
	"context"
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// derived from https://github.com/kubernetes/client-go/blob/master/dynamic/simple.go

type fakeDynamicClient struct {
	client kubernetes.Interface
}

var _ dynamic.Interface = &fakeDynamicClient{}

// NewForClient
func NewForClient(client kubernetes.Interface) dynamic.Interface {
	return &fakeDynamicClient{client: client}
}

type fakeDynamicResourceClient struct {
	client    *fakeDynamicClient
	namespace string
	resource  schema.GroupVersionResource
}

func (c *fakeDynamicClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &fakeDynamicResourceClient{client: c, resource: resource}
}

func (c *fakeDynamicResourceClient) Namespace(ns string) dynamic.ResourceInterface {
	ret := *c
	ret.namespace = ns
	return &ret
}

func (c *fakeDynamicResourceClient) Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDynamicResourceClient) Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDynamicResourceClient) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDynamicResourceClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions, subresources ...string) error {
	return errors.New("not implemented")
}

func (c *fakeDynamicResourceClient) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return errors.New("not implemented")
}

func (c *fakeDynamicResourceClient) Get(ctx context.Context, name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	if strings.EqualFold("pods", c.resource.Resource) &&
		strings.EqualFold("", c.resource.Group) &&
		strings.EqualFold("v1", c.resource.Version) {
		pod, err := kd.Clientset.CoreV1().Pods(c.namespace).Get(ctx, name, opts)
		if err != nil {
			return nil, err
		}
		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
		if err != nil {
			return nil, err
		}
		return &unstructured.Unstructured{Object: obj}, nil

	}
	return nil, fmt.Errorf("resource %s not supported", c.resource.Resource)
}

func (c *fakeDynamicResourceClient) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDynamicResourceClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDynamicResourceClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, errors.New("not implemented")
}
