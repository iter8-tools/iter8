package controllers

// RoutemapsInterface defines behavior for a set of routemaps
type RoutemapsInterface interface {
	// GetRoutemapFromNamespaceName returns a route map with the given namespace and name
	GetRoutemapFromNamespaceName(string, string) RoutemapInterface
}

// RoutemapInterface defines behavior of a routemap
type RoutemapInterface interface {
	// RLock locks the object for reading
	RLock()
	// RUnlock unlocks an object locked for reading
	RUnlock()
	// GetName returns the name of the object
	GetName() string
	// GetNamespace returns the namespace of the object
	GetNamespace() string
	// Weights provides the relative weights from traffic routing between versions
	Weights() []uint32
	// GetVersions returns a list of versions
	GetVersions() []VersionInterface
}

// VersionInterface defines behavior for a version
type VersionInterface interface {
	// GetSignature returns a signature of a version
	GetSignature() *string
}
