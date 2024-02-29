package service

import (
	"context"
	"encoding/json"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/movido/invoicing/event_models"
	"github.com/segmentio/kafka-go"
)

func (a EmailService) SendEmail(ctx context.Context, invoiceData event_models.Email) error {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	err := SendEmailMock()
	if err != nil {
		d, _ := json.Marshal(invoiceData)
		a.producer.Writer.WriteMessages(ctx, kafka.Message{Key: []byte(invoiceData.Email),
			Value: d})

		return err
	}

	logger.Print("Email sent successfully")
	return nil
}

func SendEmailMock() error {
	return nil
}
