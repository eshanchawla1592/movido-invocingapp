package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/movido/invoicing/models"
)

func (a SubscriptionService) AddSubscription(ctx context.Context, customer_id string, subscription models.Subscription) (models.Subscription, error) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// create UUID
	id := uuid.New()

	// The database is called testDb
	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/testdb")
	//db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/testdb")
	if err != nil {
		logger.Printf("error occured in connecting to DB %v", err)
		return models.Subscription{}, err

	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	query := "INSERT INTO SUBSCRIPTION VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.PrepareContext(context.TODO(), query)
	if err != nil {
		logger.Printf("error %s when preparing SQL statement", err)
		return models.Subscription{}, err
	}

	// parse created date
	created_At, err := time.Parse("2006-01-02", subscription.CreatedDate)
	if err != nil {
		logger.Print("invalid created date")
		return models.Subscription{}, err
	}

	var end_date, next_bill_date string
	// get subscription end date
	if subscription.DurationUnits == "MONTHS" {
		end_date = created_At.AddDate(0, subscription.Duration, 0).Format("2006-01-02")
	} else {
		end_date = created_At.AddDate(subscription.Duration, 0, 0).Format("2006-01-02")
	}

	// get billing date
	if subscription.BillingFrequencyUnits == "MONTHS" {
		next_bill_date = created_At.AddDate(0, subscription.BillingFrequency, 0).Format("2006-01-02")
	} else {
		next_bill_date = created_At.AddDate(subscription.BillingFrequency, 0, 0).Format("2006-01-02")
	}

	response, err := stmt.ExecContext(context.TODO(), id, customer_id, created_At, end_date,
		subscription.Duration, subscription.DurationUnits, subscription.BillingFrequency, subscription.BillingFrequencyUnits,
		subscription.Price, subscription.Currency, subscription.ProductCode, next_bill_date)
	if err != nil {
		logger.Printf("error when inserting row into subscription table: %v", err)
		return models.Subscription{}, err
	}

	rows, err := response.RowsAffected()
	if err != nil {
		logger.Printf("error when finding rows affected: %v", err)
		return models.Subscription{}, err
	}

	logger.Printf("%d subscription created", rows)

	subscription.SubscriptionId = id.String()
	subscription.CustomerId = customer_id

	return subscription, nil
}

func (a SubscriptionService) GetSubscription(ctx context.Context, customer_id, subscription_id string) (models.Subscription, error) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// The database is called testDb
	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/testdb")
	//db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/testdb")
	if err != nil {
		logger.Printf("error in connecting to db %v", err)
		return models.Subscription{}, err
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	// query
	query := fmt.Sprintf("SELECT ID, CUSTOMER_ID, CREATED_DATE,DURATION, DURATION_UNITS, BILLING_FREQUENCY, BILLING_FREQUENCY_UNITS, PRICE, CURRENCY, PRODUCT_CODE FROM SUBSCRIPTION WHERE CUSTOMER_ID='%s' AND ID='%s'", customer_id, subscription_id)
	response, err := db.Query(query)
	if err != nil {
		logger.Printf("error in reading from db %v", err)
		return models.Subscription{}, err
	}

	if response.Next() {
		var subscription models.Subscription
		err := response.Scan(&subscription.SubscriptionId, &subscription.CustomerId, &subscription.CreatedDate,
			&subscription.Duration, &subscription.DurationUnits, &subscription.BillingFrequency, &subscription.BillingFrequencyUnits,
			&subscription.Price, &subscription.Currency, &subscription.ProductCode)

		if err != nil {
			logger.Printf("error in scanning record %v", err)
			return models.Subscription{}, err
		}

		return subscription, nil
	} else {
		return models.Subscription{}, errors.New("no record found")
	}
}
