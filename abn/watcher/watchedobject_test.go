package watcher

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHasName(t *testing.T) {
	objName := "myobj"
	objNamespace := "mynamespace"
	appName := "myapp"
	appVersion := "myversion"
	appTrack := "myTrack"

	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: objNamespace,
			Labels: map[string]string{
				nameLabel:    appName,
				versionLabel: appVersion,
			},
			Annotations: map[string]string{
				trackAnnotation: appTrack,
			},
		},
		Spec: corev1.PodSpec{},
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&p)
	assert.NoError(t, err)
	wo := watchedObject{Obj: &unstructured.Unstructured{Object: obj}}

	n, ok := wo.getName()
	assert.True(t, ok)
	assert.Equal(t, appName, n)

	assert.Equal(t, objNamespace, wo.getNamespace())

	nn, ok := wo.getNamespacedName()
	assert.True(t, ok)
	assert.Equal(t, objNamespace+"/"+appName, nn)

	v, ok := wo.getVersion()
	assert.True(t, ok)
	assert.Equal(t, appVersion, v)

	assert.Equal(t, appTrack, wo.getTrack())

	assert.False(t, wo.isReady())
	assert.False(t, wo.isReady())
}

func TestHasNoName(t *testing.T) {
	objName := "myobj"
	objNamespace := "mynamespace"

	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: objNamespace,
		},
		Spec: corev1.PodSpec{},
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&p)
	assert.NoError(t, err)
	wo := watchedObject{Obj: &unstructured.Unstructured{Object: obj}}

	n, ok := wo.getName()
	assert.False(t, ok)
	assert.Empty(t, n)
}

func TestHasNoTrack(t *testing.T) {
	objName := "myobj"
	objNamespace := "mynamespace"
	appName := "myapp"
	appVersion := "myversion"

	p := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objName,
			Namespace: objNamespace,
			Labels: map[string]string{
				nameLabel:    appName,
				versionLabel: appVersion,
			},
			Annotations: map[string]string{
				readyAnnotation: "true",
			},
		},
		Spec: corev1.PodSpec{},
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&p)
	assert.NoError(t, err)
	wo := watchedObject{Obj: &unstructured.Unstructured{Object: obj}}

	assert.Equal(t, "", wo.getTrack())

	assert.True(t, wo.isReady())
	assert.True(t, wo.isReady())
}
