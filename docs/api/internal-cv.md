# API: `internal/cv`

[Back to API Index](./)

<style>
  .api-type { color: #0f766e; font-weight: 700; }
  .api-func { color: #1d4ed8; font-weight: 700; }
  .api-method { color: #7c3aed; font-weight: 700; }
  .api-signature { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace; }
</style>

Legend: <span class="api-type">Type</span>, <span class="api-func">Function</span>, <span class="api-method">Method</span>

Package: `package cv // import "github.com/smysnk/sikuligo/internal/cv"`

## Symbol Index

### Types

- <span class="api-type">[`HybridMatcher`](#type-hybridmatcher)</span>
- <span class="api-type">[`MatcherEngine`](#type-matcherengine)</span>
- <span class="api-type">[`NCCMatcher`](#type-nccmatcher)</span>
- <span class="api-type">[`ORBMatcher`](#type-orbmatcher)</span>
- <span class="api-type">[`OpenCVMatcher`](#type-opencvmatcher)</span>
- <span class="api-type">[`SADMatcher`](#type-sadmatcher)</span>

### Functions

- <span class="api-func">[`NewDefaultMatcher`](#func-newdefaultmatcher)</span>
- <span class="api-func">[`NewMatcherForEngine`](#func-newmatcherforengine)</span>
- <span class="api-func">[`NewHybridMatcher`](#func-newhybridmatcher)</span>
- <span class="api-func">[`ParseMatcherEngine`](#func-parsematcherengine)</span>
- <span class="api-func">[`NewNCCMatcher`](#func-newnccmatcher)</span>
- <span class="api-func">[`NewORBMatcher`](#func-neworbmatcher)</span>
- <span class="api-func">[`NewOpenCVMatcher`](#func-newopencvmatcher)</span>
- <span class="api-func">[`NewSADMatcher`](#func-newsadmatcher)</span>

### Methods

- <span class="api-method">[`HybridMatcher.Find`](#method-hybridmatcher-find)</span>
- <span class="api-method">[`NCCMatcher.Find`](#method-nccmatcher-find)</span>
- <span class="api-method">[`ORBMatcher.Find`](#method-orbmatcher-find)</span>
- <span class="api-method">[`OpenCVMatcher.Find`](#method-opencvmatcher-find)</span>
- <span class="api-method">[`SADMatcher.Find`](#method-sadmatcher-find)</span>

## Declarations

### Types

#### <a id="type-hybridmatcher"></a><span class="api-type">Type</span> `HybridMatcher`

- Signature: <span class="api-signature">`type HybridMatcher struct {`</span>

#### <a id="type-matcherengine"></a><span class="api-type">Type</span> `MatcherEngine`

- Signature: <span class="api-signature">`type MatcherEngine string`</span>

#### <a id="type-nccmatcher"></a><span class="api-type">Type</span> `NCCMatcher`

- Signature: <span class="api-signature">`type NCCMatcher struct{}`</span>

#### <a id="type-orbmatcher"></a><span class="api-type">Type</span> `ORBMatcher`

- Signature: <span class="api-signature">`type ORBMatcher struct{}`</span>

#### <a id="type-opencvmatcher"></a><span class="api-type">Type</span> `OpenCVMatcher`

- Signature: <span class="api-signature">`type OpenCVMatcher struct{}`</span>

#### <a id="type-sadmatcher"></a><span class="api-type">Type</span> `SADMatcher`

- Signature: <span class="api-signature">`type SADMatcher struct{}`</span>

### Functions

#### <a id="func-newdefaultmatcher"></a><span class="api-func">Function</span> `NewDefaultMatcher`

- Signature: <span class="api-signature">`func NewDefaultMatcher() core.Matcher`</span>
- Notes: NewDefaultMatcher returns the matcher backend used by default in Sikuli flows.

#### <a id="func-newmatcherforengine"></a><span class="api-func">Function</span> `NewMatcherForEngine`

- Signature: <span class="api-signature">`func NewMatcherForEngine(engine MatcherEngine) (core.Matcher, error)`</span>
- Uses: [`MatcherEngine`](#type-matcherengine)

#### <a id="func-newhybridmatcher"></a><span class="api-func">Function</span> `NewHybridMatcher`

- Signature: <span class="api-signature">`func NewHybridMatcher(primary, fallback core.Matcher) *HybridMatcher`</span>
- Uses: [`HybridMatcher`](#type-hybridmatcher)

#### <a id="func-parsematcherengine"></a><span class="api-func">Function</span> `ParseMatcherEngine`

- Signature: <span class="api-signature">`func ParseMatcherEngine(raw string) (MatcherEngine, error)`</span>
- Uses: [`MatcherEngine`](#type-matcherengine)

#### <a id="func-newnccmatcher"></a><span class="api-func">Function</span> `NewNCCMatcher`

- Signature: <span class="api-signature">`func NewNCCMatcher() *NCCMatcher`</span>
- Uses: [`NCCMatcher`](#type-nccmatcher)

#### <a id="func-neworbmatcher"></a><span class="api-func">Function</span> `NewORBMatcher`

- Signature: <span class="api-signature">`func NewORBMatcher() *ORBMatcher`</span>
- Uses: [`ORBMatcher`](#type-orbmatcher)

#### <a id="func-newopencvmatcher"></a><span class="api-func">Function</span> `NewOpenCVMatcher`

- Signature: <span class="api-signature">`func NewOpenCVMatcher() *OpenCVMatcher`</span>
- Uses: [`OpenCVMatcher`](#type-opencvmatcher)

#### <a id="func-newsadmatcher"></a><span class="api-func">Function</span> `NewSADMatcher`

- Signature: <span class="api-signature">`func NewSADMatcher() *SADMatcher`</span>
- Uses: [`SADMatcher`](#type-sadmatcher)

### Methods

#### <a id="method-hybridmatcher-find"></a><span class="api-method">Method</span> `HybridMatcher.Find`

- Signature: <span class="api-signature">`func (m *HybridMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)`</span>

#### <a id="method-nccmatcher-find"></a><span class="api-method">Method</span> `NCCMatcher.Find`

- Signature: <span class="api-signature">`func (m *NCCMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)`</span>

#### <a id="method-orbmatcher-find"></a><span class="api-method">Method</span> `ORBMatcher.Find`

- Signature: <span class="api-signature">`func (m *ORBMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)`</span>

#### <a id="method-opencvmatcher-find"></a><span class="api-method">Method</span> `OpenCVMatcher.Find`

- Signature: <span class="api-signature">`func (m *OpenCVMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)`</span>

#### <a id="method-sadmatcher-find"></a><span class="api-method">Method</span> `SADMatcher.Find`

- Signature: <span class="api-signature">`func (m *SADMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)`</span>

## Raw Package Doc

```text
package cv // import "github.com/smysnk/sikuligo/internal/cv"


FUNCTIONS

func NewDefaultMatcher() core.Matcher
    NewDefaultMatcher returns the matcher backend used by default in Sikuli
    flows.

func NewMatcherForEngine(engine MatcherEngine) (core.Matcher, error)

TYPES

type HybridMatcher struct {
	// Has unexported fields.
}

func NewHybridMatcher(primary, fallback core.Matcher) *HybridMatcher

func (m *HybridMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)

type MatcherEngine string

const (
	MatcherEngineTemplate MatcherEngine = "template"
	MatcherEngineORB      MatcherEngine = "orb"
	MatcherEngineHybrid   MatcherEngine = "hybrid"
)
func ParseMatcherEngine(raw string) (MatcherEngine, error)

type NCCMatcher struct{}

func NewNCCMatcher() *NCCMatcher

func (m *NCCMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)

type ORBMatcher struct{}

func NewORBMatcher() *ORBMatcher

func (m *ORBMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)

type OpenCVMatcher struct{}

func NewOpenCVMatcher() *OpenCVMatcher

func (m *OpenCVMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)

type SADMatcher struct{}

func NewSADMatcher() *SADMatcher

func (m *SADMatcher) Find(req core.SearchRequest) ([]core.MatchCandidate, error)

```
