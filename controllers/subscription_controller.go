package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/movido/invoicing/models"
)

func (a AppController) AddSubscription(c *gin.Context) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// get customer_id from request
	customer_id := c.Param("customer_id")

	// read request
	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle error
		logger.Printf("error occured in reading request %v", err)
		c.JSON(http.StatusInternalServerError, convertToError(http.StatusInternalServerError, err))
		return
	}

	var createSubscriptionReq models.Subscription
	err = json.Unmarshal(jsonData, &createSubscriptionReq)
	if err != nil {
		logger.Printf("error occured in parsing request %v", err)
		c.JSON(http.StatusBadRequest, convertToError(http.StatusBadRequest, err))
		return
	}

	if createSubscriptionReq.DurationUnits != "MONTHS" && createSubscriptionReq.DurationUnits != "YEARS" {
		c.JSON(http.StatusBadRequest, convertToError(http.StatusBadRequest, errors.New("invalid duration units")))
		return
	}

	if createSubscriptionReq.BillingFrequencyUnits != "MONTHS" && createSubscriptionReq.BillingFrequencyUnits != "YEARS" {
		c.JSON(http.StatusBadRequest, convertToError(http.StatusBadRequest, errors.New("invalid billing frequency units")))
		return
	}

	response, err := a.subscriptionService.AddSubscription(c, customer_id, createSubscriptionReq)
	if err != nil {
		logger.Printf("error in processing request %v", err)
		c.JSON(http.StatusInternalServerError, convertToError(http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, response)
}

func (a AppController) GetSubscription(c *gin.Context) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	customer_id := c.Param("customer_id")
	subscription_id := c.Param("subscription_id")

	subscription, err := a.subscriptionService.GetSubscription(c, customer_id, subscription_id)
	if err != nil {
		if err.Error() == "no record found" {
			c.JSON(http.StatusNotFound, convertToError(http.StatusNotFound, err))
			return
		}
		logger.Printf("failed to get data from db %v", err)
		c.JSON(http.StatusInternalServerError, convertToError(http.StatusInternalServerError, err))
	}

	c.JSON(http.StatusOK, subscription)
}
