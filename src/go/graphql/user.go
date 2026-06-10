package graphql

import (
	"context"
	"fmt"
	"m2cpcli/tools"
)

func UserIdByEmailOrId(ctx context.Context, userEmailOrId string) (UUID, error) {
	var err error
	var userId UUID
	if !tools.IsValidUuid(userEmailOrId) {
		currentName := userEmailOrId
		var user *User
		user, err = UserByEmail(ctx, currentName)
		if err != nil {
			return "", err
		}
		userId = user.Id
	} else {
		userId = UUID(userEmailOrId)
	}
	return userId, nil
}

// UserInfo returns user information of the current logged-in user.
func UserInfo(ctx context.Context) (*User, error) {
	queryString := `query CurrentUser {
	  currentUser {
		displayName
		email
		id
		sshPublicKey
	  }
	}`

	client, req := PrepareClientAndRequest(ctx, queryString)

	var result struct {
		User User `json:"currentUser"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return &result.User, nil
}

// UserByEmail returns the user with the given email.
func UserByEmail(ctx context.Context, email string) (*User, error) {
	queryString := `query($email: String){
		users(where: {email: {eq: $email}}){
			items {
				id
				displayName
				email
				sshPublicKey
			}
		}
	}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("email", email)

	var result struct {
		Items struct {
			Items []User `json:"items"`
		} `json:"users"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Items.Items) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return &result.Items.Items[0], nil
}

// UserModify updates the user with the given userId. If sshKey is not empty, it will be updated.
func UserModify(ctx context.Context, userId UUID, sshKey string) error {
	// Todo: When more
	queryString := `mutation($userId: UUID!, $sshKey: String){
		updateUsers(users: [{id: $userId, sshPublicKey: $sshKey}]){
			id
		}
	}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("userId", userId)
	if sshKey != "" {
		req.Var("sshKey", sshKey)
	}

	var result struct {
		UpdateUser []User `json:"updateUsers"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return err
	}

	if len(result.UpdateUser) != 1 {
		return fmt.Errorf("failed to update user")
	}

	return nil
}
