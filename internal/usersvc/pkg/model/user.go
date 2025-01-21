package model

type User struct {
	ID             *string `json:"id,omitempty"`
	UUID           *string `json:"uuid,omitempty"`
	Email          *string `json:"email,omitempty"`
	PhoneNumber    *string `json:"phoneNumber,omitempty"`
	UserName       *string `json:"userName,omitempty"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	AuthProvider   *string `json:"authProvider,omitempty"`
}
