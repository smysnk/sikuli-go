//go:build !opencv

package cv

import (
	"fmt"

	"github.com/smysnk/sikuligo/internal/core"
)

type ORBMatcher struct{}
type AKAZEMatcher struct{}
type BRISKMatcher struct{}
type KAZEMatcher struct{}
type SIFTMatcher struct{}

func NewORBMatcher() *ORBMatcher {
	return &ORBMatcher{}
}

func NewAKAZEMatcher() *AKAZEMatcher {
	return &AKAZEMatcher{}
}

func NewBRISKMatcher() *BRISKMatcher {
	return &BRISKMatcher{}
}

func NewKAZEMatcher() *KAZEMatcher {
	return &KAZEMatcher{}
}

func NewSIFTMatcher() *SIFTMatcher {
	return &SIFTMatcher{}
}

func (m *ORBMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error) {
	return unsupportedFeatureFind(req, "ORB")
}

func (m *AKAZEMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error) {
	return unsupportedFeatureFind(req, "AKAZE")
}

func (m *BRISKMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error) {
	return unsupportedFeatureFind(req, "BRISK")
}

func (m *KAZEMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error) {
	return unsupportedFeatureFind(req, "KAZE")
}

func (m *SIFTMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error) {
	return unsupportedFeatureFind(req, "SIFT")
}

func unsupportedFeatureFind(req core.SearchRequest, name string) ([]core.MatchCandidate, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("%w: %s matcher requires build with -tags opencv", core.ErrMatcherUnsupported, name)
}
