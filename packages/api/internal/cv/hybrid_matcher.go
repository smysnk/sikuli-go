package cv

import (
	"errors"
	"fmt"

	"github.com/smysnk/sikuligo/internal/core"
)

type HybridMatcher struct {
	primary  core.Matcher
	fallback core.Matcher
}

const hybridORBThresholdCeiling = 0.10

func NewHybridMatcher(primary, fallback core.Matcher) *HybridMatcher {
	return &HybridMatcher{
		primary:  primary,
		fallback: fallback,
	}
}

func (m *HybridMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error) {
	if m == nil {
		return nil, fmt.Errorf("matcher is nil")
	}
	if m.primary == nil && m.fallback == nil {
		return nil, fmt.Errorf("no matcher backends configured")
	}

	primaryResults, primaryErr := runMatcher(m.primary, req)
	if primaryErr == nil && len(primaryResults) > 0 {
		return primaryResults, nil
	}

	fallbackReq := adjustedRequestForMatcher(m.fallback, req)
	fallbackResults, fallbackErr := runMatcher(m.fallback, fallbackReq)
	if fallbackErr == nil && len(fallbackResults) > 0 {
		return fallbackResults, nil
	}

	if primaryErr == nil && len(primaryResults) == 0 {
		if fallbackErr == nil {
			return fallbackResults, nil
		}
		if errors.Is(fallbackErr, core.ErrMatcherUnsupported) {
			return nil, nil
		}
		return nil, fallbackErr
	}

	if primaryErr != nil && fallbackErr != nil {
		return nil, errors.Join(primaryErr, fallbackErr)
	}
	if primaryErr != nil {
		return nil, primaryErr
	}
	if fallbackErr != nil {
		return nil, fallbackErr
	}
	return nil, nil
}

func runMatcher(m core.Matcher, req core.SearchRequest) ([]core.MatchCandidate, error) {
	if m == nil {
		return nil, nil
	}
	return m.Find(req)
}

func adjustedRequestForMatcher(m core.Matcher, req core.SearchRequest) core.SearchRequest {
	if m == nil {
		return req
	}
	adjusted := req
	switch m.(type) {
	case *ORBMatcher:
		if adjusted.Threshold > hybridORBThresholdCeiling {
			adjusted.Threshold = hybridORBThresholdCeiling
		}
	}
	return adjusted
}
