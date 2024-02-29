package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/movido/invoicing/event_models"
	"github.com/segmentio/kafka-go"
)

func (a BillingService) BillCustomer(ctx context.Context, invoiceData event_models.Invoice) error {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// The database is called testDb
	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/testdb")
	//db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/testdb")
	if err != nil {
		logger.Printf("error occured in connecting to DB %v", err)
		return err

	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	// Prepare email template for current customer

	tmpl, err := template.ParseFiles("./template/email.html")
	if err != nil {
		return err
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, invoiceData); err != nil {
		return err
	}

	result := tpl.String()
	var next_bill_date time.Time
	today := time.Now()
	if invoiceData.BillingFrequencyUnits == "YEARS" {
		next_bill_date = today.AddDate(invoiceData.BillingFrequency, 0, 0)
	} else {
		next_bill_date = today.AddDate(0, invoiceData.BillingFrequency, 0)
	}

	// Now we will publish message to email service and update db for next billing date

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	queryString := fmt.Sprintf("update SUBSCRIPTION SET NEXT_BILL_DATE='%s' where ID='%s'", next_bill_date.Format("2006-01-02"),
		invoiceData.SubscriptionId)

	_, err = tx.Exec(queryString)
	if err != nil {
		tx.Rollback()
		logger.Printf("Failed to update billing date, transaction rolled back. Reason: %v", err.Error())
		return err
	}

	emailDate := event_models.Email{Message: result, Email: invoiceData.CustomerId + "@gmail.com"}
	data, _ := json.Marshal(emailDate)

	kafkaMessage := kafka.Message{
		Key:   []byte(uuid.New().String()),
		Value: data,
	}

	err = a.producer.Writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		tx.Rollback()
		logger.Printf("Failed to publish message for email service, transaction rolled back. Reason: %v", err.Error())
		return err
	}

	logger.Print("Updated next billing date and published message")

	tx.Commit()
	return nil
}
