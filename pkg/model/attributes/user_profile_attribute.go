package attributes

// UserProfileAttributeValue is the attribute value for user profiles
// @description User profile information
type UserProfileAttributeValue struct {
	DisplayName string `json:"display_name" example:"John Doe"`
	GivenName   string `json:"given_name" example:"John"`
	FamilyName  string `json:"family_name" example:"Doe"`
	Locale      string `json:"locale" example:"en-US"`
	PictureUri  string `json:"picture_uri" example:"https://example.com/profile.jpg"`
}

// GetIndex returns the index of the user profile attribute value
// User profiles don't have an index for lookup, so return empty string
func (u *UserProfileAttributeValue) GetIndex() string {
	return ""
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (u *UserProfileAttributeValue) IndexIsSensitive() bool {
	return false // User profiles don't have an index
}
