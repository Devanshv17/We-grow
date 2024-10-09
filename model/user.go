package model

import "time"

type User struct {
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Role        string    `json:"role"`
	PhoneNumber string    `json:"phone_number"` // Unique phone number
	Name        string    `json:"name"`
	Gender      string    `json:"gender"` // Should be 'male', 'female', or 'others'
	City        string    `json:"city"`
	ChildDOB    time.Time `json:"child_dob"` // Child's date of birth
}