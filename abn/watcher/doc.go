// Package watcher provides methods to watch resources and build a runtime model of applications

package watcher

// Configured resource types are watched in a set of namespaces. Each version of an application
// of interest should be include EXACTLY one watched resource with the following metadata:
//  - label 'iter8.tools/abn' set to true indicating the resource should be inspected further
//  - label 'app.kubernetes.io/name' identifying the name of the application to which the resource belongs
//  - label 'app.kubernetes.io/version' identifying the name of the version  to which the resource belongs
//
// It may also include the following metadata:
//  - annotation 'iter8.tools/ready' indicating the version is ready to receive traffic (default: false)
//  - annotation 'iter8.tools/track' defining the track label that should be associated with the version (no track when not present)

// A track is a user assigned identifier. When the Iter8 A/N(/n) service is used to lookup versions,
// the track identifier is returned (instead of a version name). The caller can use the the track
// identifier to route requests to the appropriate version. To do this, the set of track identifiers
// should be a (small) fixed set, such as "current" and "candidate", that can be mapped to a set of
// static routes.
