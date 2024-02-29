package event_models

type Invoice struct {
	RetryCount            int     `json:"retryCount,omitempty"`
	CustomerId            string  `json:"customerId,omitempty"`
	SubscriptionId        string  `json:"subscriptionId,omitempty"`
	CreatedDate           string  `json:"created_date,omitempty"`
	Duration              int     `json:"duration,omitempty"`
	DurationUnits         string  `json:"durationUnits,omitempty"`
	BillingFrequency      int     `json:"billingFrequency,omitempty"`
	BillingFrequencyUnits string  `json:"billingFrequencyUnits,omitempty"`
	Price                 float64 `json:"price"`
	Currency              string  `json:"currency,omitempty"`
	ProductCode           string  `json:"productCode,omitempty"`
}

type Email struct {
	Message string `json:"message,omitempty"`
	Email   string `json:"email,omitempty"`
}
