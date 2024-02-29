package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/movido/invoicing/models"
)

type CustomError struct {
	StatusCode int    `json:"statusCode,omitempty" db:"ID"`
	Message    string `json:"message,omitempty" db:"ID"`
}

func convertToError(code int, errorObj error) CustomError {
	return CustomError{code, errorObj.Error()}
}

func (a AppController) AddCustomer(c *gin.Context) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// read request
	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle error
		logger.Printf("error occured in reading request %v", err)
		c.JSON(http.StatusInternalServerError, convertToError(http.StatusInternalServerError, err))
		return
	}

	var createCustomerReq models.Customer
	err = json.Unmarshal(jsonData, &createCustomerReq)
	if err != nil {
		logger.Printf("error occured in parsing request %v", err)
		c.JSON(http.StatusBadRequest, convertToError(http.StatusBadRequest, err))
		return
	}

	response, err := a.customerService.AddCustomer(c, createCustomerReq)
	if err != nil {
		logger.Printf("error in processing request %v", err)
		c.JSON(http.StatusInternalServerError, convertToError(http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, response)
}

func (a AppController) GetCustomer(c *gin.Context) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// read path param
	name := c.Param("customer_id")

	customer, err := a.customerService.GetCustomer(c, name)
	if err != nil {
		if err.Error() == "no record found" {
			c.JSON(http.StatusNotFound, convertToError(http.StatusNotFound, err))
			return
		}
		logger.Printf("failed to get data from db %v", err)
		c.JSON(http.StatusInternalServerError, convertToError(http.StatusInternalServerError, err))
	}

	c.JSON(http.StatusOK, customer)
}
