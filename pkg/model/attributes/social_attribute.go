package attributes

// SocialAttributeValue is the attribute value for social accounts
// @description Social account information
type SocialAttributeValue struct {
	SocialIDP string `json:"social_idp" example:"google"`
	SocialID  string `json:"social_id" example:"1234567890"`
}

// GetIndex returns the index of the social attribute value
func (s *SocialAttributeValue) GetIndex() string {
	return s.SocialIDP + "/" + s.SocialID
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (s *SocialAttributeValue) IndexIsSensitive() bool {
	return false // Social IDP/ID is not sensitive
}
