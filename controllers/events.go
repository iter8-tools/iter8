package controllers

import (
	"github.com/iter8-tools/iter8/controllers/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	typedv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

// broadcastEvent broadcasts an event to the controller
func broadcastEvent(object runtime.Object, eventtype, reason, message string, client k8sclient.Interface) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// TODO: Do we want to reuse the event broadcaster?
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(4)
	eventBroadcaster.StartRecordingToSink(&typedv1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	eventRecorder := eventBroadcaster.NewRecorder(scheme, corev1.EventSource{})

	eventRecorder.Event(object, eventtype, reason, message)
	eventBroadcaster.Shutdown()
}
