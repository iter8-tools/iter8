package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	typedv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

// broadcastEvent broadcasts an event to the controller
func broadcastEvent(eventtype, reason, message string, client k8sclient.Interface) {
	ns, ok := os.LookupEnv(podNamespaceEnvVariable)
	if !ok {
		log.Logger.Errorf("could not get pod namespace from environment variable %s", podNamespaceEnvVariable)
		return
	}

	name, ok := os.LookupEnv(podNameEnvVariable)
	if !ok {
		log.Logger.Errorf("could not get pod name from environment variable %s", podNameEnvVariable)
		return
	}

	log.Logger.Trace(fmt.Sprintf("name: %s, namespace: %s", name, ns))

	p := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec:   v1.PodSpec{},
		Status: v1.PodStatus{},
	}

	jsonP, _ := json.Marshal(p)
	log.Logger.Trace(fmt.Sprintf("p: %s", string(jsonP)))

	pod, err := client.CoreV1().Pods(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Logger.Errorf("could not get pod with name %s in namespace %s", name, ns)
		return
	}

	jsonPod, _ := json.Marshal(pod)
	log.Logger.Trace(fmt.Sprintf("Pod: %s", string(jsonPod)))

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// TODO: Do we want to reuse the event broadcaster?
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(4)
	eventBroadcaster.StartRecordingToSink(&typedv1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	eventRecorder := eventBroadcaster.NewRecorder(scheme, v1.EventSource{})

	eventRecorder.Event(&p, eventtype, reason, message)
	eventBroadcaster.Shutdown()
}
