package oauth2

type ScopeMapping struct {
	Scope  string   `json:"scope"`  // e.g. email
	Claims []string `json:"claims"` // e.g. ["email", "email_verified"]
}

var DefaultScopeMapping = []ScopeMapping{
	{
		Scope:  "email",
		Claims: []string{"email", "email_verified"},
	},
	{
		Scope: "profile",
		Claims: []string{"website",
			"zoneinfo",
			"birthdate",
			"gender",
			"profile",
			"preferred_username",
			"given_name",
			"middle_name",
			"locale",
			"picture",
			"updated_at",
			"name",
			"nickname",
			"family_name"},
	},
	{
		Scope:  "address",
		Claims: []string{"address"},
	},
	{
		Scope:  "phone",
		Claims: []string{"phone_number", "phone_number_verified"},
	},
}
