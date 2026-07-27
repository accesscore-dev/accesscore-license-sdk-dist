package aclic

// What the host supplies.
//
// The core reads no file, opens no socket, and counts nothing. Everything it needs that lives
// outside a pure function arrives through one of these, contributed by the product.

// RegistrationRecord is everything a deployment persists about its registration.
//
// Wider than what a judgement reads. The signed token carries the granted rights; the rest is
// local state the activation client needs across restarts — the credentials that authenticate a
// heartbeat or a release, the monotonic watermark that makes a rolled-back clock detectable,
// and the revocation stamp that has to outlive a restart to mean anything.
type RegistrationRecord struct {
	ProductKey            string `json:"productKey,omitempty"`
	LicenseToken          string `json:"licenseToken,omitempty"`
	ActivationID          string `json:"activationId,omitempty"`
	ActivationSecret      string `json:"activationSecret,omitempty"`
	InstanceID            string `json:"instanceId,omitempty"`
	VerificationWatermark string `json:"verificationWatermark,omitempty"`
	LastHeartbeatAt       string `json:"lastHeartbeatAt,omitempty"`
	RevokedAt             string `json:"revokedAt,omitempty"`
}

// StoredRecord is the part of a registration one judgement reads.
type StoredRecord struct {
	LicenseToken          string `json:"licenseToken,omitempty"`
	VerificationWatermark string `json:"verificationWatermark,omitempty"`
	LastHeartbeatAt       string `json:"lastHeartbeatAt,omitempty"`
	RevokedAt             string `json:"revokedAt,omitempty"`
}

// ToStoredRecord narrows a registration to what a judgement needs.
func (r *RegistrationRecord) ToStoredRecord() StoredRecord {
	return StoredRecord{
		LicenseToken:          r.LicenseToken,
		VerificationWatermark: r.VerificationWatermark,
		LastHeartbeatAt:       r.LastHeartbeatAt,
		RevokedAt:             r.RevokedAt,
	}
}

// HasToken reports whether a signed token is present.
func (r *RegistrationRecord) HasToken() bool {
	return r.LicenseToken != ""
}

// HasActivationCredentials reports whether heartbeat and release requests can be authenticated.
func (r *RegistrationRecord) HasActivationCredentials() bool {
	return r.ActivationID != "" && r.ActivationSecret != ""
}

// Store reads and writes this deployment's registration.
//
// The SDK does not know how a product persists anything, so the product contributes the store.
// A database row that survives container replacement and a file on a mounted volume are both
// valid answers, and the SDK has no opinion between them.
type Store interface {
	// Load returns the persisted registration, or nil when nothing is registered.
	Load() *RegistrationRecord
	// Save persists a registration, replacing whatever was stored before.
	Save(record *RegistrationRecord)
	// Clear removes the registration, returning the deployment to unregistered.
	Clear()
}

// QuotaCounter supplies the current count for a licensed quota.
//
// The SDK knows the ceilings but not how the things being counted are modelled, so the product
// contributes counters — one per area, each answering only for the codes it owns. A quota
// nothing answers for is not enforced, and that is said once at start rather than let through
// silently.
type QuotaCounter interface {
	// CountedQuotaCodes are the codes this counter answers for.
	CountedQuotaCodes() []string
	// Count is the current count for a code this counter supports.
	Count(quotaCode string) int64
}

// FeatureCatalogue lists every feature key this deployment actually gates on.
//
// The keys are scattered across the product's gates, so nothing in the SDK can enumerate them.
// What the list is for is reconciliation: the issuer sells modules by key and cannot know which
// of them a given deployment can act on, and a key that was sold but is gated nowhere earns a
// customer nothing.
type FeatureCatalogue interface {
	GatedFeatures() []string
}

// ProductIdentity says which product this deployment is, and which release runs here.
//
// This comes from code, never from configuration. A deployment that could edit its own product
// code could accept a license issued for another product, and the issuer signs every product
// with one key.
type ProductIdentity interface {
	ProductCode() string
	Release() string
}

// FeatureActivationChecker decides whether a licensed feature is effectively active.
//
// The license grants permission; an administrator grants activation. A billable add-on is
// effectively active only when both say so.
type FeatureActivationChecker interface {
	IsEffectivelyActive(feature string) bool
}

// LifecycleEvent names a registration event worth a record.
type LifecycleEvent string

// The registration lifecycle events worth a record.
const (
	Activated   LifecycleEvent = "ACTIVATED"
	Deactivated LifecycleEvent = "DEACTIVATED"
	Heartbeat   LifecycleEvent = "HEARTBEAT"
	Revoked     LifecycleEvent = "REVOKED"
)

// AuditRecorder records license events into the product's audit trail.
//
// An implementation must not let a recording failure reach its caller. Judgement runs at start
// and on a schedule, and a failed audit write may never change what a deployment concludes
// about its license.
type AuditRecorder interface {
	// StatusChanged records a change in status. Only a transition is worth a row.
	StatusChanged(previous, current, licenseID string)
	// Lifecycle records a registration lifecycle event.
	Lifecycle(event LifecycleEvent, licenseID, detail string)
	// LimitReached records an operation refused because a licensed ceiling was reached.
	LimitReached(quotaCode string, current, limit int64)
}

// ActivationServerDirectory says where this deployment reaches its activation server.
//
// Configuration alone cannot answer this. A deployment is installed before anyone decides which
// server it belongs to — often by someone holding the address on a printed page and with no way
// to edit a file on the host — so the address has to be settable while the product runs and has
// to survive a restart.
type ActivationServerDirectory interface {
	// ServerURL is the address to use, or an empty string to fall back to the configured one.
	ServerURL() string
}
