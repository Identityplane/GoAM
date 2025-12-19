package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Identityplane/GoAM/pkg/model"
	services_interface "github.com/Identityplane/GoAM/pkg/services"
)

// UserClaimsService handles User Claims related operations
type UserClaimsService struct {
}

// NewOAuth2Service creates a new OAuth2Service instance
func NewUserClaimsService() services_interface.UserClaimsService {
	return &UserClaimsService{}
}

func (s *UserClaimsService) GetUserClaims(user model.User, scope string, oauth2Session *model.Oauth2Session) (map[string]interface{}, error) {

	// If userid is empty we return an error. This might be the case if a client uses the client_credentials grant and then accesses the userinfo endpoint
	if user.ID == "" {
		return nil, fmt.Errorf("internal server error. User ID is empty")
	}

	// now we map the user attributes into claims
	// we need to check the sesssion scopes and map the attributes accordingly
	claims := make(map[string]interface{})
	scopes := strings.Split(scope, " ")

	// We always return the sub claim
	claims["sub"] = user.ID

	if oauth2Session != nil && !oauth2Session.AuthTime.IsZero() {
		claims["auth_time"] = oauth2Session.AuthTime.Unix()
	}

	if slices.Contains(scopes, "email") {
		setEmailClaimsForUser(user, claims)
	}

	if slices.Contains(scopes, "profile") {
		setProfileClaimsForUser(user, claims)
	}

	return claims, nil
}

func setEmailClaimsForUser(user model.User, claims map[string]interface{}) {

	email, _, err := model.GetAttribute[model.EmailAttributeValue](&user, model.AttributeTypeEmail)
	if err != nil {
		return
	}

	if email != nil {
		claims["email"] = email.Email
		claims["email_verified"] = email.Verified
	} else {
		claims["email"] = ""
		claims["email_verified"] = false
	}
}

func setProfileClaimsForUser(user model.User, claims map[string]interface{}) {
	profile, _, err := model.GetAttribute[model.UsernameAttributeValue](&user, model.AttributeTypeUsername)
	if err != nil {
		return
	}

	if profile != nil {
		if profile.Website != "" {
			claims["website"] = profile.Website
		}
		if profile.Zoneinfo != "" {
			claims["zoneinfo"] = profile.Zoneinfo
		}
		if profile.Birthdate != "" {
			claims["birthdate"] = profile.Birthdate
		}
		if profile.Gender != "" {
			claims["gender"] = profile.Gender
		}
		if profile.Profile != "" {
			claims["profile"] = profile.Profile
		}
		if profile.PreferredUsername != "" {
			claims["preferred_username"] = profile.PreferredUsername
		}
		if profile.GivenName != "" {
			claims["given_name"] = profile.GivenName
		}
		if profile.MiddleName != "" {
			claims["middle_name"] = profile.MiddleName
		}
		if profile.Locale != "" {
			claims["locale"] = profile.Locale
		}
		if profile.Picture != "" {
			claims["picture"] = profile.Picture
		}
		if profile.UpdatedAt != "" {
			claims["updated_at"] = profile.UpdatedAt
		}
		if profile.Name != "" {
			claims["name"] = profile.Name
		}
		if profile.Nickname != "" {
			claims["nickname"] = profile.Nickname
		}
		if profile.FamilyName != "" {
			claims["family_name"] = profile.FamilyName
		}
	}
}
