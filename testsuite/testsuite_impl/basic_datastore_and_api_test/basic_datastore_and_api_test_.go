/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package basic_datastore_and_api_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/services_impl/api"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	apiServiceId services.ServiceID = "api"

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxNumPolls = 15

	testPersonId = 23
	testNumBooksRead = 3
)

type BasicDatastoreAndApiTest struct {
	datstoreImage string
	apiImage string
}

func NewBasicDatastoreAndApiTest(datstoreImage string, apiImage string) *BasicDatastoreAndApiTest {
	return &BasicDatastoreAndApiTest{datstoreImage: datstoreImage, apiImage: apiImage}
}

func (b BasicDatastoreAndApiTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	datastoreInitializer := datastore.NewDatastoreContainerInitializer(b.datstoreImage)
	uncastedDatastoreSvc, datastoreChecker, err := networkCtx.AddService(datastoreServiceId, datastoreInitializer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}
	if err := datastoreChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to start")
	}

	// Go doesn't have generics so we need to do this cast
	datastoreSvc := uncastedDatastoreSvc.(*datastore.DatastoreService)

	apiInitializer := api.NewApiContainerInitializer(b.apiImage, datastoreSvc)
	_, apiChecker, err := networkCtx.AddService(apiServiceId, apiInitializer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}
	if err := apiChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxNumPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API service to start")
	}
	return networkCtx, nil
}


func (b BasicDatastoreAndApiTest) Run(network networks.Network, testCtx testsuite.TestContext) {
	// Go doesn't have generics so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)

	uncastedApiService, err := castedNetwork.GetService(apiServiceId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the API service"))
	}
	apiService := uncastedApiService.(*api.ApiService)

	logrus.Infof("Verifying that person with test ID '%v' doesn't already exist...", testPersonId)
	if _, err = apiService.GetPerson(testPersonId); err == nil {
		testCtx.Fatal(stacktrace.NewError("Expected an error trying to get a person who doesn't exist yet, but didn't receive one"))
	}
	logrus.Infof("Verified that test person doesn't already exist")

	logrus.Infof("Adding test person with ID '%v'...", testPersonId)
	if err := apiService.AddPerson(testPersonId); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred adding person with test ID '%v'", testPersonId))
	}
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read by %v...", testNumBooksRead)
	for i := 0; i < testNumBooksRead; i++ {
		if err := apiService.IncrementBooksRead(testPersonId); err != nil {
			testCtx.Fatal(stacktrace.Propagate(err, "An error occurred incrementing the number of books read"))
		}
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read...")
	person, err := apiService.GetPerson(testPersonId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the test person to verify the number of books read"))
	}
	logrus.Info("Retrieved test person")

	testCtx.AssertTrue(
		person.BooksRead == testNumBooksRead,
		stacktrace.NewError(
			"Expected number of book read '%v' != actual number of books read '%v'",
			testNumBooksRead,
			person.BooksRead,
		),
	)
}

func (test *BasicDatastoreAndApiTest) GetTestConfiguration() testsuite.TestConfiguration {
	return testsuite.TestConfiguration{}
}

func (test *BasicDatastoreAndApiTest) GetExecutionTimeout() time.Duration {
	return 60 * time.Second
}

func (test *BasicDatastoreAndApiTest) GetSetupTimeout() time.Duration {
	return 60 * time.Second
}
