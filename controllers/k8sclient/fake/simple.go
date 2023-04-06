// Package fake provides fake Kuberntes clients for testing
package fake

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	fakek8s "k8s.io/client-go/kubernetes/fake"
)

// Client provides structured and dynamic fake clients
type Client struct {
	*fakek8s.Clientset
	*fakedynamic.FakeDynamicClient
}

/*
Patch applies a patch for a resource.
Important: fake clients should not be used for server-side apply or strategic-merge patches.
https://github.com/kubernetes/kubernetes/pull/78630#issuecomment-500424163

Hence, we are mocking the Patch call in this fake client so that,
instead of server-side apply as in the real client, we perform of merge patch instead.
*/
func (cl *Client) Patch(gvr schema.GroupVersionResource, objNamespace string, objName string, jsonBytes []byte) (*unstructured.Unstructured, error) {
	return cl.FakeDynamicClient.Resource(gvr).Namespace(objNamespace).Patch(context.TODO(), objName, types.MergePatchType, jsonBytes, metav1.PatchOptions{})
}

// New returns a new fake Kubernetes client populated with runtime objects
func New(sObjs []runtime.Object, unsObjs []runtime.Object) *Client {
	s := runtime.NewScheme()
	return &Client{
		fakek8s.NewSimpleClientset(sObjs...),
		fakedynamic.NewSimpleDynamicClientWithCustomListKinds(s, map[schema.GroupVersionResource]string{
			{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}: "DeploymentList",
			{
				Group:    "",
				Version:  "v1",
				Resource: "configmaps",
			}: "ConfigMapList",
			{
				Group:    "networking.istio.io",
				Version:  "v1beta1",
				Resource: "virtualservices",
			}: "VirtualServiceList",
			{
				Group:    "",
				Version:  "v1",
				Resource: "services",
			}: "ServiceList",
			{
				Group:    "serving.kserve.io",
				Version:  "v1beta1",
				Resource: "inferenceservices",
			}: "InferenceServiceList",
		}, unsObjs...),
	}
}
