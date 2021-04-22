/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package datastore

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	healthcheckUrlSlug = "health"
	healthyValue       = "healthy"

	textContentType = "text/plain"
	keyEndpoint = "key"
)

type DatastoreService struct {
	serviceId services.ServiceID
	ipAddr string
	port int
}

func NewDatastoreService(serviceId services.ServiceID, ipAddr string, port int) *DatastoreService {
	return &DatastoreService{serviceId: serviceId, ipAddr: ipAddr, port: port}
}

// ===========================================================================================
//                              Service interface methods
// ===========================================================================================
func (service DatastoreService) GetServiceID() services.ServiceID {
	return service.serviceId
}

func (service DatastoreService) GetIPAddress() string {
	return service.ipAddr
}


func (service DatastoreService) IsAvailable() bool {
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
//                         Datastore service-specific methods
// ===========================================================================================
func (service DatastoreService) GetPort() int {
	return service.port
}

func (service DatastoreService) Exists(key string) (bool, error) {
	url := service.getUrlForKey(key)
	resp, err := http.Get(url)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
	}
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return false, stacktrace.NewError("Got an unexpected HTTP status code: %v", resp.StatusCode)
	}
}

func (service DatastoreService) Get(key string) (string, error) {
	url := service.getUrlForKey(key)
	resp, err := http.Get(url)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
	}
	if resp.StatusCode != http.StatusOK {
		return "", stacktrace.NewError("A non-OK status code was returned: %v", resp.StatusCode)
	}
	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body")
	}
	return string(bodyBytes), nil
}

func (service DatastoreService) Upsert(key string, value string) error {
	url := service.getUrlForKey(key)
	resp, err := http.Post(url, textContentType, strings.NewReader(value))
	if err != nil {
		return stacktrace.Propagate(err, "An error requesting to upsert data '%v' to key '%v'", value, key)
	}
	if resp.StatusCode != http.StatusOK {
		return stacktrace.NewError("A non-%v status code was returned", resp.StatusCode)
	}
	return nil
}

func (service DatastoreService) getUrlForKey(key string) string {
	return fmt.Sprintf("http://%v:%v/%v/%v", service.GetIPAddress(), service.port, keyEndpoint, key)
}


