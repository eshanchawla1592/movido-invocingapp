package service

type CustomerService struct {
}

func NewCustomerService() CustomerService {
	return CustomerService{}
}

type SubscriptionService struct {
}

func NewSubscriptionService() SubscriptionService {
	return SubscriptionService{}
}

type BillingService struct {
	producer KafkaProducer
}

type EmailService struct {
	producer KafkaProducer
}

func NewEmailService(producer KafkaProducer) EmailService {
	return EmailService{
		producer: producer,
	}
}

func NewBillingService(producer KafkaProducer) BillingService {
	return BillingService{
		producer: producer,
	}
}
