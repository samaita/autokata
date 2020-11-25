package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func handleDomainAdd(c *gin.Context) {
	var (
		err     error
		isExist bool
	)

	response := getInitialResponse()

	domain := NewDomain()
	domain.DomainName = c.PostForm("domain_name")
	domain.DomainURL = c.PostForm("domain_url")
	domain.FeedsURL = c.PostForm("feeds_url")
	if !domain.validate() {
		HTTPBadRequest(c, fmt.Errorf("Bad Input").Error(), response)
		return
	}

	if isExist, err = domain.isExist(); err != nil {
		HTTPInternalServerError(c, err.Error(), response)
		return
	}

	if !isExist {
		if err = domain.add(); err != nil {
			HTTPInternalServerError(c, err.Error(), response)
			return
		}
	}

	response[fieldSuccess] = valueSuccess
	HTTPSuccessResponse(c, response)
	return
}

func handleDomainRemove(c *gin.Context) {
	var (
		err     error
		isExist bool
	)

	response := getInitialResponse()

	domain := NewDomain()
	domain.DomainURL = c.PostForm("domain_url")

	if isExist, err = domain.isExist(); err != nil {
		HTTPInternalServerError(c, err.Error(), response)
		return
	}

	if isExist {
		if err = domain.remove(); err != nil {
			HTTPInternalServerError(c, err.Error(), response)
			return
		}
	}

	response[fieldSuccess] = valueSuccess
	HTTPSuccessResponse(c, response)
	return
}
