package abn

import "github.com/iter8-tools/iter8/controllers"

type testroutemapsByName map[string]*testroutemap
type testroutemaps struct {
	nsRoutemap map[string]testroutemapsByName
}

func (s *testroutemaps) GetRoutemapFromNamespaceName(namespace string, name string) controllers.RoutemapInterface {
	rmByName, ok := s.nsRoutemap[namespace]
	if ok {
		return rmByName[name]
	}
	return nil
}

type testversion struct {
	signature *string
}

func (v *testversion) GetSignature() *string {
	return v.signature
}

type testroutemap struct {
	name              string
	namespace         string
	versions          []testversion
	normalizedWeights []uint32
}

func (s *testroutemap) RLock() {}

func (s *testroutemap) RUnlock() {}

func (s *testroutemap) GetNamespace() string {
	return s.namespace
}

func (s *testroutemap) GetName() string {
	return s.name
}

func (s *testroutemap) Weights() []uint32 {
	return s.normalizedWeights
}

func (s *testroutemap) GetVersions() []controllers.VersionInterface {
	result := make([]controllers.VersionInterface, len(s.versions))
	for i := range s.versions {
		v := s.versions[i]
		result[i] = controllers.VersionInterface(&v)
	}
	return result
}
