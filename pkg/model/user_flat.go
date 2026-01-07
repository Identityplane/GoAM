package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// UserFlat represents a user in the system without the complexity of user attributes
// It is used to simplify the user object for external APIs and clients
// It only supports 1 email address and 1 phone number as well as 1 username
type UserFlat struct {
	ID          string     `json:"id" db:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Tenant      string     `json:"tenant" db:"tenant" example:"acme"`
	Realm       string     `json:"realm" db:"realm" example:"customers"`
	Status      string     `json:"status" db:"status" example:"active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at" example:"2024-01-01T00:00:00Z"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at" example:"2024-01-01T00:00:00Z"`

	// Username
	PreferredUsername string `json:"preferred_username,omitempty" example:"john.doe"`
	Website           string `json:"website,omitempty" example:"https://example.com"`
	Zoneinfo          string `json:"zoneinfo,omitempty" example:"Europe/Berlin"`
	Birthdate         string `json:"birthdate,omitempty" example:"1990-01-01"`
	Gender            string `json:"gender,omitempty" example:"male"`
	Profile           string `json:"profile,omitempty" example:"https://example.com/profile"`
	GivenName         string `json:"given_name,omitempty" example:"John"`
	MiddleName        string `json:"middle_name,omitempty" example:"Doe"`
	Locale            string `json:"locale,omitempty" example:"en-US"`
	Picture           string `json:"picture,omitempty" example:"https://example.com/picture.jpg"`
	Name              string `json:"name,omitempty" example:"John Doe"`
	Nickname          string `json:"nickname,omitempty" example:"john.doe"`
	FamilyName        string `json:"family_name,omitempty" example:"Doe"`

	// Email Attribute
	Email           string     `json:"email,omitempty" db:"email" example:"john.doe@example.com"`
	EmailVerified   *bool      `json:"email_verified,omitempty" db:"email_verified" example:"true"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at" example:"2024-01-01T00:00:00Z"`

	// Phone Attribute
	Phone           string     `json:"phone,omitempty" db:"phone" example:"+1234567890"`
	PhoneVerified   *bool      `json:"phone_verified,omitempty" db:"phone_verified" example:"true"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty" db:"phone_verified_at" example:"2024-01-01T00:00:00Z"`
}

func (user *User) ToUserFlat() *UserFlat {
	userFlat := &UserFlat{
		ID:          user.ID,
		Tenant:      user.Tenant,
		Realm:       user.Realm,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}

	// Map the first email attribute
	emails, _, err := GetAttributes[EmailAttributeValue](user, AttributeTypeEmail)
	if err == nil && len(emails) > 0 && emails[0].Email != "" {
		userFlat.Email = emails[0].Email
		verified := emails[0].Verified
		userFlat.EmailVerified = &verified
		userFlat.EmailVerifiedAt = emails[0].VerifiedAt
	}

	// Map the first phone attribute
	phones, _, err := GetAttributes[PhoneAttributeValue](user, AttributeTypePhone)
	if err == nil && len(phones) > 0 && phones[0].Phone != "" {
		userFlat.Phone = phones[0].Phone
		verified := phones[0].Verified
		userFlat.PhoneVerified = &verified
		userFlat.PhoneVerifiedAt = phones[0].VerifiedAt
	}

	// Map the first username attribute
	usernames, _, err := GetAttributes[UsernameAttributeValue](user, AttributeTypeUsername)
	if err == nil && len(usernames) > 0 {
		userFlat.PreferredUsername = usernames[0].PreferredUsername
		userFlat.Website = usernames[0].Website
		userFlat.Zoneinfo = usernames[0].Zoneinfo
		userFlat.Birthdate = usernames[0].Birthdate
		userFlat.Gender = usernames[0].Gender
		userFlat.Profile = usernames[0].Profile
		userFlat.GivenName = usernames[0].GivenName
		userFlat.MiddleName = usernames[0].MiddleName
		userFlat.Locale = usernames[0].Locale
		userFlat.Picture = usernames[0].Picture
		userFlat.Name = usernames[0].Name
		userFlat.Nickname = usernames[0].Nickname
		userFlat.FamilyName = usernames[0].FamilyName
	}

	return userFlat
}

func (userflat *UserFlat) ToUser() *User {
	user := &User{
		ID:          userflat.ID,
		Tenant:      userflat.Tenant,
		Realm:       userflat.Realm,
		Status:      userflat.Status,
		CreatedAt:   userflat.CreatedAt,
		UpdatedAt:   userflat.UpdatedAt,
		LastLoginAt: userflat.LastLoginAt,
	}

	// Add the email attribute (only if email is present)
	if userflat.Email != "" {
		emailVerified := false
		if userflat.EmailVerified != nil {
			emailVerified = *userflat.EmailVerified
		}
		user.AddAttribute(&UserAttribute{
			Type: AttributeTypeEmail,
			Value: EmailAttributeValue{
				Email:      userflat.Email,
				Verified:   emailVerified,
				VerifiedAt: userflat.EmailVerifiedAt,
			},
		})
	}

	// Add the phone attribute (only if phone is present)
	if userflat.Phone != "" {
		phoneVerified := false
		if userflat.PhoneVerified != nil {
			phoneVerified = *userflat.PhoneVerified
		}
		user.AddAttribute(&UserAttribute{
			Type: AttributeTypePhone,
			Value: PhoneAttributeValue{
				Phone:      userflat.Phone,
				Verified:   phoneVerified,
				VerifiedAt: userflat.PhoneVerifiedAt,
			},
		})
	}

	// Add the username attribute (only if at least one field is present)
	usernameValue := UsernameAttributeValue{
		PreferredUsername: userflat.PreferredUsername,
		Website:           userflat.Website,
		Zoneinfo:          userflat.Zoneinfo,
		Birthdate:         userflat.Birthdate,
		Gender:            userflat.Gender,
		Profile:           userflat.Profile,
		GivenName:         userflat.GivenName,
		MiddleName:        userflat.MiddleName,
		Locale:            userflat.Locale,
		Picture:           userflat.Picture,
		Name:              userflat.Name,
		Nickname:          userflat.Nickname,
		FamilyName:        userflat.FamilyName,
	}
	// Only create username attribute if at least one field is non-empty
	if usernameValue.PreferredUsername != "" ||
		usernameValue.Website != "" ||
		usernameValue.Zoneinfo != "" ||
		usernameValue.Birthdate != "" ||
		usernameValue.Gender != "" ||
		usernameValue.Profile != "" ||
		usernameValue.GivenName != "" ||
		usernameValue.MiddleName != "" ||
		usernameValue.Locale != "" ||
		usernameValue.Picture != "" ||
		usernameValue.Name != "" ||
		usernameValue.Nickname != "" ||
		usernameValue.FamilyName != "" {
		user.AddAttribute(&UserAttribute{
			Type:  AttributeTypeUsername,
			Value: usernameValue,
		})
	}

	return user
}

func (uf *UserFlat) UnmarshalJSON(data []byte) error {
	// Make a mirror struct where time fields are strings
	type Alias UserFlat
	aux := &struct {
		CreatedAt       *string `json:"created_at"`
		UpdatedAt       *string `json:"updated_at"`
		LastLoginAt     *string `json:"last_login_at"`
		EmailVerifiedAt *string `json:"email_verified_at"`
		PhoneVerifiedAt *string `json:"phone_verified_at"`
		*Alias
	}{
		Alias: (*Alias)(uf),
	}

	// First, unmarshal into aux
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Now manually parse each timestamp
	parseTime := func(s *string) (*time.Time, error) {
		if s == nil || *s == "" {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, *s)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}

	var err error
	if t, err := parseTime(aux.CreatedAt); err != nil {
		return fmt.Errorf("invalid created_at: %w", err)
	} else if t != nil {
		uf.CreatedAt = *t
	}
	if t, err := parseTime(aux.UpdatedAt); err != nil {
		return fmt.Errorf("invalid updated_at: %w", err)
	} else if t != nil {
		uf.UpdatedAt = *t
	}
	if uf.LastLoginAt, err = parseTime(aux.LastLoginAt); err != nil {
		return fmt.Errorf("invalid last_login_at: %w", err)
	}
	if uf.EmailVerifiedAt, err = parseTime(aux.EmailVerifiedAt); err != nil {
		return fmt.Errorf("invalid email_verified_at: %w", err)
	}
	if uf.PhoneVerifiedAt, err = parseTime(aux.PhoneVerifiedAt); err != nil {
		return fmt.Errorf("invalid phone_verified_at: %w", err)
	}

	return nil
}
