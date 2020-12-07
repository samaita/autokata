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

	if isExist, err = domain.isExist(); err != nil || isExist {
		if isExist {
			err = fmt.Errorf("Domain already exists")
		}
		HTTPInternalServerError(c, err.Error(), response)
		return
	}

	if err = domain.add(); err != nil {
		HTTPInternalServerError(c, err.Error(), response)
		return
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

	if isExist, err = domain.isExist(); err != nil || !isExist {
		if isExist {
			err = fmt.Errorf("Domain not found")
		}
		HTTPInternalServerError(c, err.Error(), response)
		return
	}

	if err = domain.remove(); err != nil {
		HTTPInternalServerError(c, err.Error(), response)
		return
	}

	response[fieldSuccess] = valueSuccess
	HTTPSuccessResponse(c, response)
	return
}

func handleDomainList(c *gin.Context) {
	var (
		err        error
		listDomain []Domain
	)

	response := getInitialResponse()

	if listDomain, err = getAllDomain(); err != nil {
		HTTPInternalServerError(c, err.Error(), response)
		return
	}

	response["data"] = listDomain
	response["total_data"] = len(listDomain)

	response[fieldSuccess] = valueSuccess
	HTTPSuccessResponse(c, response)
	return
}

func handleFeedList(c *gin.Context) {
	var (
		err      error
		listFeed []Feed
	)

	response := getInitialResponse()

	if listFeed, err = getAllFeed(); err != nil {
		HTTPInternalServerError(c, err.Error(), response)
		return
	}

	response["data"] = listFeed
	response["total_data"] = len(listFeed)

	response[fieldSuccess] = valueSuccess
	HTTPSuccessResponse(c, response)
}
