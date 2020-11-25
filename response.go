package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	keyTimeStart = "time_start"
	fieldSuccess = "success"

	valueSuccess = 1
)

type HeaderHTTPResponse struct {
	ProcessTime  string `json:"process_time"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type DataHTTPResponse struct {
	Header  HeaderHTTPResponse `json:"header"`
	Data    interface{}        `json:"data"`
	Success bool               `json:"is_success"`
}

func getInitialResponse() map[string]interface{} {
	d := make(map[string]interface{})
	d[fieldSuccess] = false
	return d
}

func (d *DataHTTPResponse) getProcessTime(c *gin.Context) {
	tc, exist := c.Get(keyTimeStart)
	if exist {
		t, ok := tc.(time.Time)
		if ok {
			d.Header.ProcessTime = time.Since(t).String()
		} else {
			d.Header.ProcessTime = "-2"
		}
	} else {
		d.Header.ProcessTime = "-1"
	}
}

// HTTPBadRequest : function for returning error response caused by invalid input / request
func HTTPBadRequest(c *gin.Context, errMsg string, data interface{}) {
	var response DataHTTPResponse
	response.getProcessTime(c)
	response.Header.ErrorMessage = errMsg
	response.Data = data

	c.JSON(http.StatusBadRequest, response)
}

// HTTPInternalServerError : function for returning error response caused by something is wrong when execute request
func HTTPInternalServerError(c *gin.Context, errMsg string, data interface{}) {
	var response DataHTTPResponse
	response.getProcessTime(c)
	response.Header.ErrorMessage = errMsg
	response.Data = data

	c.JSON(http.StatusInternalServerError, response)
}

// HTTPSuccessResponse is a response that indicating server fulfill the request
func HTTPSuccessResponse(c *gin.Context, data interface{}) {
	var response DataHTTPResponse
	response.getProcessTime(c)
	response.Data = data
	response.Success = true

	c.JSON(http.StatusOK, response)
}
