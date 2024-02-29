package controllers

import (
	"github.com/movido/invoicing/service"
)

func NewController(customerService service.CustomerService,
	subscriptionService service.SubscriptionService, kafkaProducer service.KafkaProducer) AppController {
	return AppController{
		customerService:     customerService,
		subscriptionService: subscriptionService,
		kafkaProducer:       kafkaProducer,
	}
}

type AppController struct {
	customerService     service.CustomerService
	subscriptionService service.SubscriptionService
	kafkaProducer       service.KafkaProducer
}
