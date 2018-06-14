package models

import "testing"

func TestNotificationEmail_ValidateEmail(t *testing.T) {
	email := "test"

	ne := NotificationEmail{}

	ne.Address = email

	if !ne.ValidateEmail() {
		t.Error("Email is not valid")
	}

}
