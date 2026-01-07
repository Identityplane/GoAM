package attributes

// EntitlementSetAttributeValue is a set of entitlements
// @description Entitlement set information
type EntitlementSetAttributeValue struct {
	Entitlements []Entitlement `json:"entitlements" example:"['admin', 'user']"`
}

// GetIndex returns the index of the entitlement attribute value
// Entitlements don't have an index for lookup, so return empty string
func (e *EntitlementSetAttributeValue) GetIndex() string {
	return ""
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (e *EntitlementSetAttributeValue) IndexIsSensitive() bool {
	return false // Entitlements don't have an index
}

// Entitlement represents a single entitlement
type Entitlement struct {
	Description string     `json:"description" example:"Admin"`
	Resource    string     `json:"resource" example:"arn:identityplane:acme:customers:users:123"`
	Action      string     `json:"action" example:"read"`
	Effect      EffectType `json:"effect" example:"allow"`
}

// EffectType represents the effect of an entitlement
type EffectType string

const (
	EffectTypeAllow EffectType = "allow"
	EffectTypeDeny  EffectType = "deny"
)
