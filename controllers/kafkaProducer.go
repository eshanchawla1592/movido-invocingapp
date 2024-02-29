package controllers

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func (a AppController) FilterAndInitiateInvoicing(c *gin.Context) {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	logger.Print("initiating invoicing")

	err := a.kafkaProducer.FilterAndInitiateInvoicing(c)

	// if error occured, return error http status to cron job, forcing re-execution
	if err != nil {
		logger.Print("failed in billing customer")
		c.JSON(http.StatusInternalServerError, convertToError(http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, nil)
	return
}
