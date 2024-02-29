package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/movido/invoicing/controllers"
	"github.com/movido/invoicing/event_models"
	"github.com/movido/invoicing/service"
	"github.com/segmentio/kafka-go"
)

func main() {
	// default router
	r := gin.Default()

	customerService := service.NewCustomerService()
	subscriptionService := service.NewSubscriptionService()
	kafkaProducer := service.NewKafkaProducer("broker:9092", "invoicing")
	emailKafkaProducer := service.NewKafkaProducer("broker:9092", "email")
	billingService := service.NewBillingService(emailKafkaProducer)
	kafkaEmailReconciling := service.NewKafkaProducer("broker:9092", "email-reconciling")
	emailService := service.NewEmailService(kafkaEmailReconciling)
	kafkaProducerForReconciling := service.NewKafkaProducer("broker:9092", "reconciling")
	appController := controllers.NewController(customerService, subscriptionService, kafkaProducer)
	routerGroup := r.Group("/api/v1")

	// Ping test
	routerGroup.POST("/customers", appController.AddCustomer)
	routerGroup.GET("/customers/:customer_id", appController.GetCustomer)
	routerGroup.POST("/customers/:customer_id/subscription", appController.AddSubscription)
	routerGroup.GET("/customers/:customer_id/subscription/:subscription_id", appController.GetSubscription)

	routerGroup.GET("/internal/bill-customer", appController.FilterAndInitiateInvoicing)

	// start kafka consumer in the background
	go StartKafkaConsumerForBillingInitiation(kafkaProducerForReconciling, billingService)
	go StartEmailConsumer(kafkaProducerForReconciling, emailService)
	r.Run(":8080")
}

func StartKafkaConsumerForBillingInitiation(producer service.KafkaProducer, billingService service.BillingService) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	conf := kafka.ReaderConfig{
		Brokers:  []string{"broker:9092"},
		Topic:    "invoicing",
		GroupID:  "1",
		MaxBytes: 10,
	}

	reader := kafka.NewReader(conf)

	logger.Print("starting kafka consumer")

	for {
		// blocking call
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("error occured", err)
			continue
		}
		fmt.Println("Message is : ", string(m.Value))

		var invoiceData event_models.Invoice

		err = json.Unmarshal(m.Value, &invoiceData)
		if err != nil {
			// will publish to manual handling
			logger.Print("failed to parse message: ", err)
		}

	retry:
		err = billingService.BillCustomer(context.Background(), invoiceData)
		if err != nil {
			logger.Print("failed to consumer message: ", err)
			// publish for retry if
			invoiceData.RetryCount = invoiceData.RetryCount + 1
			if invoiceData.RetryCount < 3 {
				// wait for 5 seconds
				time.Sleep(5 * time.Second)
				goto retry
			} else {
				// publish for manual reconcilation
				data, _ := json.Marshal(invoiceData)
				producer.Writer.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte(invoiceData.SubscriptionId),
					Value: data,
				})

				logger.Printf("Subscription sent for manual billing: %s", invoiceData.SubscriptionId)
			}
		}
	}

}

func StartEmailConsumer(producer service.KafkaProducer, emailService service.EmailService) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	conf := kafka.ReaderConfig{
		Brokers:  []string{"broker:9092"},
		Topic:    "email",
		GroupID:  "1",
		MaxBytes: 10,
	}

	reader := kafka.NewReader(conf)

	logger.Print("starting kafka consumer")

	for {
		// blocking call
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("error occured", err)
			continue
		}
		fmt.Println("Message is : ", string(m.Value))
		var emailMessage event_models.Email

		json.Unmarshal(m.Value, &emailMessage)
		emailService.SendEmail(context.Background(), emailMessage)
	}

}
