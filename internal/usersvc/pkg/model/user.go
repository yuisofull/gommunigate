package model

type User struct {
	ID             *string `json:"id,omitempty"`
	UUID           *string `json:"uuid"`
	Email          *string `json:"email"`
	PhoneNumber    *string `json:"phoneNumber"`
	UserName       *string `json:"userName"`
	ProfilePicture *string `json:"profilePicture"`
	Bio            *string `json:"bio"`
	AuthProvider   *string `json:"authProvider"`
}
