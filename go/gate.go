package aclic

import (
	"encoding/json"
	"sync"
)

// The refusal an unlicensed or unentitled request gets. Not 401 — no credential would change
// the answer.
const Forbidden = 403

// The codes a refused request carries beyond the ones a judgement produces.
const (
	FeatureNotLicensed  = "LICENSE_FEATURE_NOT_LICENSED"
	FeatureNotActivated = "LICENSE_FEATURE_NOT_ACTIVATED"
	QuotaExceeded       = "LICENSE_QUOTA_EXCEEDED"
)

// Denial is a message key with the values that fill its placeholders.
//
// The core carries keys and arguments, never translated text: every language the SDK binds has
// its own internationalization, and a core holding wording would grow a phrase book each time a
// language is added.
type Denial struct {
	MessageKey string `json:"messageKey"`
	Arguments  []any  `json:"arguments"`
}

// Evaluation is everything one judgement returns.
type Evaluation struct {
	Status            string           `json:"status"`
	Usable            bool             `json:"usable"`
	ErrorCode         string           `json:"errorCode"`
	Denial            *Denial          `json:"denial,omitempty"`
	NextWatermark     string           `json:"nextWatermark,omitempty"`
	AvailableFeatures []string         `json:"availableFeatures"`
	Limits            map[string]int64 `json:"limits"`
	GraceExpiresAt    string           `json:"graceExpiresAt,omitempty"`
	HeartbeatDeadline string           `json:"heartbeatDeadline,omitempty"`
	Payload           *Payload         `json:"payload,omitempty"`
}

// Payload is what one signed license says.
type Payload struct {
	SchemaVersion          int               `json:"schemaVersion"`
	Alg                    string            `json:"alg"`
	Kid                    string            `json:"kid"`
	LicenseID              string            `json:"licenseId"`
	ActivationID           string            `json:"activationId,omitempty"`
	CustomerID             string            `json:"customerId,omitempty"`
	Customer               string            `json:"customer,omitempty"`
	Channel                string            `json:"channel"`
	ProductCode            string            `json:"productCode"`
	MaxRelease             string            `json:"maxRelease,omitempty"`
	TierLabel              string            `json:"tierLabel,omitempty"`
	IssuedAt               string            `json:"issuedAt"`
	ActivatedAt            string            `json:"activatedAt,omitempty"`
	ExpiresAt              string            `json:"expiresAt"`
	GracePeriodDays        int               `json:"gracePeriodDays"`
	MachineFingerprints    []string          `json:"machineFingerprints"`
	Nonce                  string            `json:"nonce,omitempty"`
	Features               []string          `json:"features"`
	Limits                 map[string]int64  `json:"limits"`
	HeartbeatRequired      bool              `json:"heartbeatRequired"`
	HeartbeatIntervalDays  int               `json:"heartbeatIntervalDays"`
	HeartbeatToleranceDays int               `json:"heartbeatToleranceDays"`
	Metadata               map[string]string `json:"metadata"`
}

// LicenseDenied is what a gate returns when it refuses a request.
type LicenseDenied struct {
	Status     string
	ErrorCode  string
	Denial     Denial
	HTTPStatus int
}

func (d *LicenseDenied) Error() string {
	return d.ErrorCode + ": " + d.Denial.MessageKey
}

// Body is the one document every adapter writes.
func (d *LicenseDenied) Body() []byte {
	arguments := d.Denial.Arguments
	if arguments == nil {
		arguments = []any{}
	}
	body, err := json.Marshal(map[string]any{
		"errorCode":  d.ErrorCode,
		"status":     d.Status,
		"messageKey": d.Denial.MessageKey,
		"arguments":  arguments,
	})
	if err != nil {
		// The document is a map of strings; a failure here cannot happen from this input, and
		// answering with nothing would hide the refusal.
		return []byte(`{"errorCode":"LICENSE_NOT_ACTIVATED","status":"NOT_ACTIVATED",` +
			`"messageKey":"error.license.notValid","arguments":[]}`)
	}
	return body
}

// Gate is the neutral gate every request-time check goes through.
//
// Four operations, the same four in every language the SDK binds:
//
//	gate.State()                     the cached judgement
//	gate.RequireFeature(key)         refuse when the license does not grant it
//	gate.CheckQuota(code, current)   refuse when the ceiling is passed
//	gate.Refresh(evaluation)         replace the cache with a fresh judgement
//
// The binding holds the judgement; the host decides when to refresh it. The gate that runs on
// every request reads this cache and never calls the core.
type Gate struct {
	mutex             sync.RWMutex
	state             Evaluation
	activationChecker func(feature string) bool
}

// NewGate builds a gate that has judged nothing yet.
func NewGate() *Gate {
	return &Gate{state: unknown()}
}

// NewGateWithActivation builds a gate that also consults whether an administrator switched a
// feature on.
func NewGateWithActivation(checker func(feature string) bool) *Gate {
	return &Gate{state: unknown(), activationChecker: checker}
}

// unknown is the state a deployment starts in, before anything has been judged.
func unknown() Evaluation {
	return Evaluation{
		Status:    "NOT_ACTIVATED",
		Usable:    false,
		ErrorCode: "LICENSE_NOT_ACTIVATED",
		Denial: &Denial{
			MessageKey: "error.license.notActivated",
			Arguments:  []any{},
		},
		AvailableFeatures: []string{},
		Limits:            map[string]int64{},
	}
}

// State is the cached judgement.
func (g *Gate) State() Evaluation {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.state
}

// Refresh replaces the cache with a fresh judgement.
func (g *Gate) Refresh(evaluation *Evaluation) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if evaluation == nil {
		g.state = unknown()
		return
	}
	g.state = *evaluation
}

// Status is what the judgement concluded.
func (g *Gate) Status() string {
	return g.State().Status
}

// IsUsable reports whether the deployment keeps operating.
func (g *Gate) IsUsable() bool {
	return g.State().Usable
}

// RequireUsable refuses when the deployment is not licensed to serve at all.
func (g *Gate) RequireUsable() error {
	state := g.State()
	if state.Usable {
		return nil
	}
	denial := Denial{MessageKey: "error.license.notValid", Arguments: []any{}}
	if state.Denial != nil {
		denial = *state.Denial
	}
	return &LicenseDenied{
		Status:     state.Status,
		ErrorCode:  state.ErrorCode,
		Denial:     denial,
		HTTPStatus: Forbidden,
	}
}

// RequireFeature refuses when the license does not grant a feature, or it is switched off.
func (g *Gate) RequireFeature(feature string) error {
	if err := g.RequireUsable(); err != nil {
		return err
	}

	state := g.State()
	if !contains(state.AvailableFeatures, feature) {
		return &LicenseDenied{
			Status:    state.Status,
			ErrorCode: FeatureNotLicensed,
			Denial: Denial{
				MessageKey: "error.license.featureNotLicensed",
				Arguments:  []any{feature},
			},
			HTTPStatus: Forbidden,
		}
	}
	if !g.IsActivated(feature) {
		return &LicenseDenied{
			Status:    state.Status,
			ErrorCode: FeatureNotActivated,
			Denial: Denial{
				MessageKey: "error.license.featureNotActivated",
				Arguments:  []any{feature},
			},
			HTTPStatus: Forbidden,
		}
	}
	return nil
}

// CheckQuota refuses when taking one more of something would pass its ceiling.
//
// The comparison is against the count after the addition, so a deployment holding exactly its
// ceiling is refused the next one rather than the one it already has.
func (g *Gate) CheckQuota(quotaCode string, current int64) error {
	if err := g.RequireUsable(); err != nil {
		return err
	}

	state := g.State()
	limit, named := state.Limits[quotaCode]
	if !named {
		return nil
	}
	if current+1 <= limit {
		return nil
	}
	return &LicenseDenied{
		Status:    state.Status,
		ErrorCode: QuotaExceeded,
		Denial: Denial{
			MessageKey: "error.license.quotaExceeded",
			Arguments:  []any{quotaCode, limit},
		},
		HTTPStatus: Forbidden,
	}
}

// IsLicensed reports whether the license grants a feature, ignoring activation.
func (g *Gate) IsLicensed(feature string) bool {
	return contains(g.State().AvailableFeatures, feature)
}

// IsActivated reports whether an administrator has switched a feature on.
func (g *Gate) IsActivated(feature string) bool {
	if g.activationChecker == nil {
		return true
	}
	return g.activationChecker(feature)
}

// IsAllowed reports whether the license grants a feature and it is switched on.
func (g *Gate) IsAllowed(feature string) bool {
	return g.IsLicensed(feature) && g.IsActivated(feature)
}

// EffectiveFeatures are the features a client may act on: granted by the license and switched
// on. A client gates its navigation on this list, so it has to agree with the server-side
// gates — which it does, because both read it from here.
func (g *Gate) EffectiveFeatures() []string {
	effective := []string{}
	for _, feature := range g.State().AvailableFeatures {
		if g.IsActivated(feature) {
			effective = append(effective, feature)
		}
	}
	return effective
}

// contains reports whether a list holds a value.
func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
