package attributes

// UsernameAttributeValue is the attribute value for usernames
// @description Username information
type UsernameAttributeValue struct {
	PreferredUsername string `json:"preferred_username" example:"john.doe"`
	Website           string `json:"website" example:"https://example.com"`
	Zoneinfo          string `json:"zoneinfo" example:"Europe/Berlin"`
	Birthdate         string `json:"birthdate" example:"1990-01-01"`
	Gender            string `json:"gender" example:"male"`
	Profile           string `json:"profile" example:"https://example.com/profile"`
	GivenName         string `json:"given_name" example:"John"`
	MiddleName        string `json:"middle_name" example:"Doe"`
	Locale            string `json:"locale" example:"en-US"`
	Picture           string `json:"picture" example:"https://example.com/picture.jpg"`
	UpdatedAt         string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Name              string `json:"name" example:"John Doe"`
	Nickname          string `json:"nickname" example:"john.doe"`
	FamilyName        string `json:"family_name" example:"Doe"`
}

// GetIndex returns the index of the username attribute value
func (u *UsernameAttributeValue) GetIndex() string {
	return u.PreferredUsername
}

// IndexIsSensitive returns whether the index should be omitted from JSON API responses
func (u *UsernameAttributeValue) IndexIsSensitive() bool {
	return false // Usernames are not sensitive
}
