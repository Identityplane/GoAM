package node_email

import (
	"context"
	"time"

	"github.com/Identityplane/GoAM/internal/auth/graph/node_utils"
	"github.com/Identityplane/GoAM/pkg/model"
	"github.com/google/uuid"
)

var SaveEmailNode = &model.NodeDefinition{
	Name:                 "saveEmail",
	PrettyName:           "Save Email",
	Description:          "Saves the email address to the user's attributes",
	Category:             "Email",
	Type:                 model.NodeTypeLogic,
	RequiredContext:      []string{"email"},
	OutputContext:        []string{"email"},
	PossibleResultStates: []string{"success", "email_taken"},
	Run:                  RunSaveEmailNode,
}

func RunSaveEmailNode(state *model.AuthenticationSession, node *model.GraphNode, input map[string]string, services *model.Repositories) (*model.NodeResult, error) {

	user, err := node_utils.TryLoadUserFromContext(state, services)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	email := state.Context["email"]

	// Check if there is already a user that has this email but is a different user
	otherUser, err := services.UserRepo.GetByAttributeIndex(context.Background(), model.AttributeTypeEmail, email)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	if otherUser != nil && otherUser.ID != user.ID {
		errorMsg := "Email already in use"
		state.Error = &errorMsg
		return model.NewNodeResultWithCondition("email_taken")
	}

	// Check if the user already has an email attribute
	emailValue, attribute, err := model.GetAttribute[model.EmailAttributeValue](user, model.AttributeTypeEmail)
	if err != nil {
		return model.NewNodeResultWithError(err)
	}

	newEmailValue := &model.EmailAttributeValue{}
	newEmailValue.Email = email

	now := time.Now()
	if state.Context["email_verified"] == "true" {
		newEmailValue.Verified = true
		newEmailValue.VerifiedAt = &now
	} else {
		newEmailValue.Verified = false
		newEmailValue.VerifiedAt = nil
	}

	if emailValue != nil {

		attribute.Value = newEmailValue
		attribute.Index = email
		services.UserRepo.UpdateUserAttribute(context.Background(), attribute)

		// Update the attribute in the user's UserAttributes slice
		for i, attr := range user.UserAttributes {
			if attr.ID == attribute.ID {
				user.UserAttributes[i].Value = newEmailValue
				break
			}
		}

		return model.NewNodeResultWithCondition("success")
	} else {
		// User does not have an email attribute, so we create a new one
		services.UserRepo.CreateUserAttribute(context.Background(), &model.UserAttribute{
			ID:    uuid.NewString(),
			Type:  model.AttributeTypeEmail,
			Value: newEmailValue,
			Index: email,
		})

		user.UserAttributes = append(user.UserAttributes, *attribute)
	}

	return model.NewNodeResultWithCondition("success")
}
