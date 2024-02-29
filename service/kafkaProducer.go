package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/movido/invoicing/event_models"
	"github.com/movido/invoicing/models"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	Writer *kafka.Writer
}

func NewKafkaProducer(kafkaURL, topic string) KafkaProducer {
	return KafkaProducer{Writer: &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}}
}

func (producer KafkaProducer) FilterAndInitiateInvoicing(ctx context.Context) error {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// get all customers to be invoiced today
	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/testdb")
	//db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/testdb")
	if err != nil {
		logger.Printf("error in connecting to db %v", err)
		return err
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	today := time.Now().Format("2006-01-02")

	query := fmt.Sprintf("SELECT ID, CUSTOMER_ID, CREATED_DATE,DURATION, DURATION_UNITS, BILLING_FREQUENCY, BILLING_FREQUENCY_UNITS, PRICE, CURRENCY, PRODUCT_CODE FROM SUBSCRIPTION WHERE NEXT_BILL_DATE='%s' AND END_DATE>='%s'", today, today)
	response, err := db.Query(query)
	if err != nil {
		logger.Printf("error in reading from db %v", err)
		return err
	}

	// publish kafka message for every customer to be billed today
	for response.Next() {
		var subscription models.Subscription
		err := response.Scan(&subscription.SubscriptionId, &subscription.CustomerId, &subscription.CreatedDate,
			&subscription.Duration, &subscription.DurationUnits, &subscription.BillingFrequency, &subscription.BillingFrequencyUnits,
			&subscription.Price, &subscription.Currency, &subscription.ProductCode)
		if err != nil {
			logger.Printf("error in scanning record %v", err)
			return err
		}

		// Need to publish invoicing message now
		invoice := event_models.Invoice{
			RetryCount:            0,
			CustomerId:            subscription.CustomerId,
			SubscriptionId:        subscription.SubscriptionId,
			Price:                 subscription.Price,
			Currency:              subscription.Currency,
			BillingFrequency:      subscription.BillingFrequency,
			BillingFrequencyUnits: subscription.BillingFrequencyUnits,
			ProductCode:           subscription.ProductCode,
		}

		invoiceData, _ := json.Marshal(invoice)

		kafkaMessage := kafka.Message{
			Key:   []byte(uuid.New().String()),
			Value: invoiceData,
		}

		err = producer.Writer.WriteMessages(ctx, kafkaMessage)
		if err != nil {
			logger.Printf("error in publishing kafka message: %v", err)
			return err
		}
		logger.Print("published message for subscription_id", subscription.SubscriptionId)
	}

	return nil
}
