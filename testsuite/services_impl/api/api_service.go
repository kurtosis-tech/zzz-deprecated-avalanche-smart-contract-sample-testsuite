/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

const (
	healthcheckUrlSlug = "health"
	healthyValue       = "healthy"

	textContentType = "text/plain"

	personEndpoint = "person"
	incrementBooksReadEndpoint = "incrementBooksRead"
)

type Person struct {
	BooksRead int
}

type ApiService struct {
	serviceId services.ServiceID
	ipAddr string
	port int
}

func NewApiService(serviceId services.ServiceID, ipAddr string, port int) *ApiService {
	return &ApiService{serviceId: serviceId, ipAddr: ipAddr, port: port}
}

// ===========================================================================================
//                              Service interface methods
// ===========================================================================================
func (service ApiService) GetServiceID() services.ServiceID {
	return service.serviceId
}

func (service ApiService) GetIPAddress() string {
	return service.ipAddr
}

func (service ApiService) IsAvailable() bool {
	url := fmt.Sprintf("http://%v:%v/%v", service.GetIPAddress(), service.port, healthcheckUrlSlug)
	resp, err := http.Get(url)
	if err != nil {
		logrus.Debugf("An HTTP error occurred when polliong the health endpoint: %v", err)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		logrus.Debugf("Received non-OK status code: %v", resp.StatusCode)
		return false
	}

	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		logrus.Debugf("An error occurred reading the response body: %v", err)
		return false
	}
	bodyStr := string(bodyBytes)

	return bodyStr == healthyValue
}

// ===========================================================================================
//                         API service-specific methods
// ===========================================================================================
func (service ApiService) getPersonUrlForId(id int) string {
	return fmt.Sprintf("http://%v:%v/%v/%v", service.ipAddr, service.port, personEndpoint, id)
}

func (service ApiService) AddPerson(id int) error {
	url := service.getPersonUrlForId(id)
	resp, err := http.Post(url, textContentType, nil)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to add person with ID '%v'", id)
	}
	if resp.StatusCode != http.StatusOK {
		return stacktrace.NewError("Adding person with ID '%v' returned non-OK status code %v", id, resp.StatusCode)
	}
	return nil
}

func (service ApiService) GetPerson(id int) (Person, error) {
	url := service.getPersonUrlForId(id)
	resp, err := http.Get(url)
	if err != nil {
		return Person{}, stacktrace.Propagate(err, "An error occurred making the request to get person with ID '%v'", id)
	}
	if resp.StatusCode != http.StatusOK {
		return Person{}, stacktrace.NewError("Getting person with ID '%v' returned non-OK status code %v", id, resp.StatusCode)
	}
	body := resp.Body
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return Person{}, stacktrace.Propagate(err, "An error occurred reading the response body")
	}

	var person Person
	if err := json.Unmarshal(bodyBytes, &person); err != nil {
		return Person{}, stacktrace.Propagate(err, "An error occurred deserializing the Person JSON")
	}
	return person, nil
}

func (service ApiService) IncrementBooksRead(id int) error {
	url := fmt.Sprintf("http://%v:%v/%v/%v", service.ipAddr, service.port, incrementBooksReadEndpoint, id)
	resp, err := http.Post(url, textContentType, nil)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to increment the books read of person with ID '%v'", id)
	}
	if resp.StatusCode != http.StatusOK {
		return stacktrace.NewError("Incrementing the books read of person with ID '%v' returned non-OK status code %v", id, resp.StatusCode)
	}
	return nil
}
