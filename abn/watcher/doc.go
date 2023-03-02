// Package watcher provides methods to watch resources and build a runtime model of applications
package watcher

// A configuration file defines which resources should be watched. Resources for a given version
// are expected to be named using the track label. These labels are:
//  <name>
//  <name>-candidate-<index>
// where index is 1, 2, 3, ...

// Iter8 maintains a mapping of track labels to ready versions. A version is ready if all of the
// resources that comprise it are present, at least one has the label `app.kubernetes.io/version`
// (if more than one does, they must all have the same value), and the resources are "ready".
// "readiness" is determined by the resource type.

// The mapping of track to version changes over time as versions change.
