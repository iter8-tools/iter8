// Package testing provides a fake dynamic client that supports finalizers.
package testing

// The default fake dynamic client does not do so. To add the appropriate behavior (if a finalizer is present,
// then add a deletion timestamp and update instead). To do so, it is necessary to reimplement the underlying
// object tracker (fixture.go).
