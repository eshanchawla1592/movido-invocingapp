package models

type Customer struct {
	CustomerId   string `json:"customerId,omitempty" db:"ID"`
	CustomerName string `json:"customerName" db:"CUSTOMER_NAME"`
	PhoneNumber  string `json:"phoneNumber" db:"PHONE_NUMBER"`
	Email        string `json:"email" db:"EMAIL"`
}
