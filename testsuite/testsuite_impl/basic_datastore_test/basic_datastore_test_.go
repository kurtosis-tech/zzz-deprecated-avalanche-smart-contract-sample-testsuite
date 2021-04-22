/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package basic_datastore_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/services_impl/datastore"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	datastoreServiceId services.ServiceID = "datastore"

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxPolls = 15

	testKey = "test-key"
	testValue = "test-value"
)

type BasicDatastoreTest struct {
	datastoreImage string
}

func NewBasicDatastoreTest(datastoreImage string) *BasicDatastoreTest {
	return &BasicDatastoreTest{datastoreImage: datastoreImage}
}

func (test BasicDatastoreTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	datastoreContainerInitializer := datastore.NewDatastoreContainerInitializer(test.datastoreImage)
	_, availabilityChecker, err := networkCtx.AddService(datastoreServiceId, datastoreContainerInitializer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}
	if err := availabilityChecker.WaitForStartup(waitForStartupTimeBetweenPolls, waitForStartupMaxPolls); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	return networkCtx, nil
}

func (test BasicDatastoreTest) Run(network networks.Network, testCtx testsuite.TestContext) {
	// Necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	uncastedService, err := castedNetwork.GetService(datastoreServiceId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the datastore service"))
	}

	// Necessary again due to no Go generics
	castedService := uncastedService.(*datastore.DatastoreService)

	logrus.Infof("Verifying that key '%v' doesn't already exist...", testKey)
	exists, err := castedService.Exists(testKey)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred checking if the test key exists"))
	}
	testCtx.AssertTrue(!exists, stacktrace.NewError("Test key should not exist yet"))
	logrus.Infof("Confirmed that key '%v' doesn't already exist", testKey)

	logrus.Infof("Inserting value '%v' at key '%v'...", testKey, testValue)
	if err := castedService.Upsert(testKey, testValue); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred upserting the test key"))
	}
	logrus.Infof("Inserted value successfully")

	logrus.Infof("Getting the key we just inserted to verify the value...")
	value, err := castedService.Get(testKey)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the test key after upload"))
	}
	testCtx.AssertTrue(
		value == testValue,
		stacktrace.NewError("Returned value '%v' != test value '%v'", value, testValue))
	logrus.Info("Value verified")
}


func (test *BasicDatastoreTest) GetTestConfiguration() testsuite.TestConfiguration {
	return testsuite.TestConfiguration{}
}

func (test *BasicDatastoreTest) GetExecutionTimeout() time.Duration {
	return 60 * time.Second
}

func (test *BasicDatastoreTest) GetSetupTimeout() time.Duration {
	return 60 * time.Second
}
