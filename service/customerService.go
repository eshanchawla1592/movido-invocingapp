package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/movido/invoicing/models"
)

func (CustomerService) AddCustomer(ctx context.Context, createCustomerReq models.Customer) (models.Customer, error) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// create UUID
	id := uuid.New()

	// The database is called testDb
	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/testdb")
	//db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/testdb")
	if err != nil {
		logger.Printf("error occured in connecting to DB %v", err)
		return models.Customer{}, err
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	// insert query
	query := "INSERT INTO CUSTOMER VALUES (?, ?, ?, ?)"

	// create statement
	stmt, err := db.PrepareContext(context.TODO(), query)
	if err != nil {
		logger.Printf("error %s when preparing SQL statement", err)
		return models.Customer{}, err
	}

	// execute query
	dbResponse, err := stmt.ExecContext(context.TODO(), id, createCustomerReq.CustomerName,
		createCustomerReq.PhoneNumber, createCustomerReq.Email)

	if err != nil {
		logger.Printf("error %s when inserting row into customer table", err)
		return models.Customer{}, err
	}

	rows, err := dbResponse.RowsAffected()
	if err != nil {
		logger.Printf("error %s when finding rows affected", err)
		return models.Customer{}, err
	}

	logger.Printf("%d customer created ", rows)

	// add record id
	createCustomerReq.CustomerId = id.String()

	return createCustomerReq, nil
}

func (CustomerService) GetCustomer(ctx context.Context, customer_id string) (models.Customer, error) {
	// logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := sql.Open("mysql", "root:password@tcp(db:3306)/testdb")
	//db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/testdb")
	if err != nil {
		logger.Printf("error in connecting to db %v", err)
		return models.Customer{}, err
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	query := fmt.Sprintf("SELECT * FROM CUSTOMER WHERE ID='%s'", customer_id)
	res, err := db.Query(query)
	if err != nil {
		logger.Printf("error in reading from db %v", err)
		return models.Customer{}, err
	}

	// check if atleast 1 record exist
	if res.Next() {
		var customer models.Customer
		err := res.Scan(&customer.CustomerId, &customer.CustomerName,
			&customer.PhoneNumber, &customer.Email)

		if err != nil {
			logger.Printf("error in scanning record %v", err)
			return models.Customer{}, err
		}

		return customer, nil
	} else {
		return models.Customer{}, errors.New("no record found")
	}
}
