package aclic

import (
	"strings"
	"time"
)

// LicenseManager ties the host's interfaces to the core.
//
// It reads the store, asks the machine what it is, calls the core, and hands the answer to the
// gate. It is the only type in the binding that touches both sides.
type LicenseManager struct {
	core                *Core
	gate                *Gate
	store               Store
	identity            ProductIdentity
	keys                []KeyMaterial
	catalogue           FeatureCatalogue
	counters            []QuotaCounter
	audit               AuditRecorder
	directory           ActivationServerDirectory
	configuredServerURL string
	instanceID          string
}

// KeyMaterial is one verification key a deployment carries.
type KeyMaterial struct {
	// Kid is the name a payload's kid matches.
	Kid string `json:"kid"`
	// Alg is what the holder says the key is. Advisory: the algorithm that decides anything
	// comes from the key material itself, and a declaration that disagrees with the key makes
	// the key unusable rather than changing how it is read.
	Alg string `json:"alg,omitempty"`
	// PublicKey is the key, as an X.509 PEM document.
	PublicKey string `json:"publicKey"`
}

// CollectedIdentifiers is what a machine reports about itself.
type CollectedIdentifiers struct {
	Fingerprints       []string `json:"fingerprints"`
	Sources            []string `json:"sources"`
	UnavailableSources []string `json:"unavailableSources"`
}

// PreparedRequest is a request ready for the host to send.
type PreparedRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	Nonce   string            `json:"nonce"`
}

// ActivationOutcome is the result of an activation, heartbeat, or release attempt.
type ActivationOutcome struct {
	Success    bool               `json:"success"`
	Error      string             `json:"error,omitempty"`
	MessageKey string             `json:"messageKey,omitempty"`
	Result     *ActivationSuccess `json:"result,omitempty"`
	Detail     string             `json:"detail"`
}

// ActivationSuccess is what a successful attempt returned.
type ActivationSuccess struct {
	LicenseToken     string   `json:"licenseToken,omitempty"`
	ActivationID     string   `json:"activationId,omitempty"`
	ActivationSecret string   `json:"activationSecret,omitempty"`
	Payload          *Payload `json:"payload,omitempty"`
}

// ManagerOption customizes a manager past its required interfaces.
type ManagerOption func(*LicenseManager)

// WithCore names the core to call, for a test that links its own.
func WithCore(core *Core) ManagerOption {
	return func(m *LicenseManager) { m.core = core }
}

// WithGate names the gate every enforcement point reads.
func WithGate(gate *Gate) ManagerOption {
	return func(m *LicenseManager) { m.gate = gate }
}

// WithCatalogue registers the catalogue of feature keys this deployment gates on.
func WithCatalogue(catalogue FeatureCatalogue) ManagerOption {
	return func(m *LicenseManager) { m.catalogue = catalogue }
}

// WithCounter registers a counter for one area's quota codes.
func WithCounter(counter QuotaCounter) ManagerOption {
	return func(m *LicenseManager) { m.counters = append(m.counters, counter) }
}

// WithAudit registers where license events are recorded.
func WithAudit(audit AuditRecorder) ManagerOption {
	return func(m *LicenseManager) { m.audit = audit }
}

// WithDirectory registers where the activation server address is kept.
func WithDirectory(directory ActivationServerDirectory) ManagerOption {
	return func(m *LicenseManager) { m.directory = directory }
}

// WithConfiguredServerURL sets the address used when nothing is registered.
func WithConfiguredServerURL(url string) ManagerOption {
	return func(m *LicenseManager) { m.configuredServerURL = normalizeServerURL(url) }
}

// WithInstanceID sets this deployment's stable identifier.
func WithInstanceID(instanceID string) ManagerOption {
	return func(m *LicenseManager) { m.instanceID = instanceID }
}

// NewLicenseManager assembles a manager over the interfaces a product contributes.
//
// The verification keys come from the caller because they belong in the binary rather than in
// configuration: a deployment that could point at another key file could verify a license it
// minted itself, and at that moment the signature check defends nothing.
func NewLicenseManager(
	store Store,
	identity ProductIdentity,
	keys []KeyMaterial,
	options ...ManagerOption,
) *LicenseManager {
	manager := &LicenseManager{
		core:     Shared,
		gate:     NewGate(),
		store:    store,
		identity: identity,
		keys:     keys,
	}
	for _, option := range options {
		option(manager)
	}
	return manager
}

// Gate is the gate every request-time check goes through.
func (m *LicenseManager) Gate() *Gate {
	return m.gate
}

// Verify judges the stored registration and publishes the result to the gate.
//
// now and integrityOK come from the host: the core reads no clock, and what to hash for an
// integrity check differs per language.
func (m *LicenseManager) Verify(now time.Time, integrityOK bool) (*Evaluation, error) {
	previous := m.gate.Status()

	collected, err := m.Fingerprint()
	if err != nil {
		return nil, err
	}

	request := map[string]any{
		"keys":                  m.keys,
		"now":                   rfc3339(now),
		"integrityOk":           integrityOK,
		"productCode":           m.identity.ProductCode(),
		"productRelease":        m.identity.Release(),
		"collectedFingerprints": collected.Fingerprints,
	}
	if record := m.store.Load(); record != nil {
		request["storedRecord"] = record.ToStoredRecord()
	}

	var evaluation Evaluation
	if err := m.core.Evaluate(request, &evaluation); err != nil {
		return nil, err
	}

	m.persistWatermark(&evaluation)
	m.gate.Refresh(&evaluation)
	m.recordTransition(previous, &evaluation)
	return &evaluation, nil
}

// Apply replaces the stored registration and judges it.
func (m *LicenseManager) Apply(
	record *RegistrationRecord,
	now time.Time,
	integrityOK bool,
) (*Evaluation, error) {
	m.store.Save(record)
	return m.Verify(now, integrityOK)
}

// MarkRevoked records that the issuing server revoked this activation and judges again.
//
// The stamp is persisted so the denial outlives a restart.
func (m *LicenseManager) MarkRevoked(now time.Time, integrityOK bool) (*Evaluation, error) {
	record := m.store.Load()
	if record == nil {
		record = &RegistrationRecord{}
	}
	record.RevokedAt = rfc3339(now)
	m.store.Save(record)
	if m.audit != nil {
		m.audit.Lifecycle(Revoked, "", "")
	}
	return m.Verify(now, integrityOK)
}

// ClearRegistration discards the registration, returning the deployment to unregistered.
func (m *LicenseManager) ClearRegistration(now time.Time, integrityOK bool) (*Evaluation, error) {
	m.store.Clear()
	return m.Verify(now, integrityOK)
}

// Fingerprint reads what this machine reports about itself.
func (m *LicenseManager) Fingerprint() (*CollectedIdentifiers, error) {
	var collected CollectedIdentifiers
	if err := m.core.Fingerprint(&collected); err != nil {
		return nil, err
	}
	return &collected, nil
}

// ActivationRequest writes the activation request the host sends.
func (m *LicenseManager) ActivationRequest(
	productKey string,
	now time.Time,
) (*PreparedRequest, error) {
	collected, err := m.Fingerprint()
	if err != nil {
		return nil, err
	}

	var prepared PreparedRequest
	err = m.core.ActivationRequest(map[string]any{
		"operation":             "ACTIVATE",
		"productKey":            productKey,
		"productCode":           m.identity.ProductCode(),
		"productRelease":        m.identity.Release(),
		"machineFingerprints":   collected.Fingerprints,
		"instanceId":            m.instanceID,
		"requestedAt":           rfc3339(now),
		"knownFeatures":         m.GatedFeatures(),
		"knownQuotas":           m.CountedQuotas(),
		"knownVerificationKeys": m.CarriedKeyNames(),
		"usage":                 map[string]int64{},
	}, &prepared)
	if err != nil {
		return nil, err
	}
	return &prepared, nil
}

// HeartbeatRequest writes the heartbeat request the host sends.
func (m *LicenseManager) HeartbeatRequest(now time.Time) (*PreparedRequest, error) {
	record := m.store.Load()
	if record == nil {
		record = &RegistrationRecord{}
	}
	collected, err := m.Fingerprint()
	if err != nil {
		return nil, err
	}

	instanceID := record.InstanceID
	if instanceID == "" {
		instanceID = m.instanceID
	}

	var prepared PreparedRequest
	err = m.core.ActivationRequest(map[string]any{
		"operation":             "HEARTBEAT",
		"activationId":          record.ActivationID,
		"activationSecret":      record.ActivationSecret,
		"productCode":           m.identity.ProductCode(),
		"productRelease":        m.identity.Release(),
		"machineFingerprints":   collected.Fingerprints,
		"instanceId":            instanceID,
		"requestedAt":           rfc3339(now),
		"knownFeatures":         m.GatedFeatures(),
		"knownQuotas":           m.CountedQuotas(),
		"knownVerificationKeys": m.CarriedKeyNames(),
		"usage":                 m.Usage(),
	}, &prepared)
	if err != nil {
		return nil, err
	}
	return &prepared, nil
}

// DeactivationRequest writes the release request the host sends, handing the seat back.
func (m *LicenseManager) DeactivationRequest(now time.Time) (*PreparedRequest, error) {
	record := m.store.Load()
	if record == nil {
		record = &RegistrationRecord{}
	}
	collected, err := m.Fingerprint()
	if err != nil {
		return nil, err
	}

	var prepared PreparedRequest
	err = m.core.ActivationRequest(map[string]any{
		"operation":           "DEACTIVATE",
		"activationId":        record.ActivationID,
		"activationSecret":    record.ActivationSecret,
		"machineFingerprints": collected.Fingerprints,
		"requestedAt":         rfc3339(now),
	}, &prepared)
	if err != nil {
		return nil, err
	}
	return &prepared, nil
}

// Interpret reads the answer the server gave.
//
// The token is checked against this deployment's key, this deployment's machine, and the nonce
// it just sent, before the host is told to store anything.
func (m *LicenseManager) Interpret(
	operation string,
	status int,
	body string,
	nonce string,
	reachable bool,
) (*ActivationOutcome, error) {
	collected, err := m.Fingerprint()
	if err != nil {
		return nil, err
	}

	keys := make([]map[string]string, 0, len(m.keys))
	for _, key := range m.keys {
		keys = append(keys, map[string]string{"kid": key.Kid, "publicKey": key.PublicKey})
	}

	var outcome ActivationOutcome
	err = m.core.ActivationResponse(map[string]any{
		"operation":             operation,
		"status":                status,
		"body":                  body,
		"nonce":                 nonce,
		"keys":                  keys,
		"collectedFingerprints": collected.Fingerprints,
		"reachable":             reachable,
	}, &outcome)
	if err != nil {
		return nil, err
	}
	return &outcome, nil
}

// ServerURL is the address to call, read per request rather than fixed at start.
//
// An operator who registers a server during installation expects the next activation to go
// there, not after a restart they have no way to perform.
func (m *LicenseManager) ServerURL() string {
	if m.directory == nil {
		return m.configuredServerURL
	}
	registered := normalizeServerURL(m.directory.ServerURL())
	if registered == "" {
		return m.configuredServerURL
	}
	return registered
}

// GatedFeatures are the feature keys this deployment gates on.
func (m *LicenseManager) GatedFeatures() []string {
	if m.catalogue == nil {
		return []string{}
	}
	return m.catalogue.GatedFeatures()
}

// CarriedKeyNames are the names of the verification keys this build carries.
//
// Taken from the keys themselves rather than from the host, because a name written down
// separately could name a key that is not here — and the issuer would then sign under a key this
// deployment cannot check.
func (m *LicenseManager) CarriedKeyNames() []string {
	names := make([]string, 0, len(m.keys))
	for _, key := range m.keys {
		names = append(names, key.Kid)
	}
	return names
}

// CountedQuotas are the quota codes this deployment can count.
func (m *LicenseManager) CountedQuotas() []string {
	codes := []string{}
	for _, counter := range m.counters {
		codes = append(codes, counter.CountedQuotaCodes()...)
	}
	return codes
}

// Usage is what this deployment currently holds, per quota code.
func (m *LicenseManager) Usage() map[string]int64 {
	usage := map[string]int64{}
	for _, counter := range m.counters {
		for _, code := range counter.CountedQuotaCodes() {
			usage[code] = counter.Count(code)
		}
	}
	return usage
}

// CheckQuota checks an addition against the ceiling, counting through the host's counters.
func (m *LicenseManager) CheckQuota(quotaCode string) error {
	var current int64
	for _, counter := range m.counters {
		if contains(counter.CountedQuotaCodes(), quotaCode) {
			current = counter.Count(quotaCode)
			break
		}
	}

	err := m.gate.CheckQuota(quotaCode, current)
	if err != nil && m.audit != nil {
		m.audit.LimitReached(quotaCode, current, m.gate.State().Limits[quotaCode])
	}
	return err
}

// persistWatermark stores the watermark the judgement produced.
func (m *LicenseManager) persistWatermark(evaluation *Evaluation) {
	if evaluation.NextWatermark == "" {
		return
	}
	record := m.store.Load()
	if record == nil {
		return
	}
	if record.VerificationWatermark == evaluation.NextWatermark {
		return
	}
	record.VerificationWatermark = evaluation.NextWatermark
	m.store.Save(record)
}

// recordTransition records a status change, and only a change.
func (m *LicenseManager) recordTransition(previous string, evaluation *Evaluation) {
	if m.audit == nil || previous == evaluation.Status {
		return
	}
	licenseID := ""
	if evaluation.Payload != nil {
		licenseID = evaluation.Payload.LicenseID
	}
	m.audit.StatusChanged(previous, evaluation.Status, licenseID)
}

// rfc3339 renders an instant the way the payload writes one.
func rfc3339(moment time.Time) string {
	return moment.UTC().Format(time.RFC3339)
}

// normalizeServerURL strips trailing slashes so two forms of one address compare equal.
func normalizeServerURL(url string) string {
	return strings.TrimRight(strings.TrimSpace(url), "/")
}
