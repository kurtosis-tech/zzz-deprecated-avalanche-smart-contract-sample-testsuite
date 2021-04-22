/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package advanced_network_test

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/networks_impl"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	testPersonId = 46
)

type AdvancedNetworkTest struct {
	datastoreServiceImage string
	apiServiceImage string

	personModifyingApiServiceId services.ServiceID
	personRetrievingApiServiceId services.ServiceID
}

func NewAdvancedNetworkTest(datastoreServiceImage string, apiServiceImage string) *AdvancedNetworkTest {
	return &AdvancedNetworkTest{datastoreServiceImage: datastoreServiceImage, apiServiceImage: apiServiceImage}
}

func (test *AdvancedNetworkTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	network := networks_impl.NewTestNetwork(networkCtx, test.datastoreServiceImage, test.apiServiceImage)

	if err := network.AddDatastore(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore")
	}

	personModifyingApiServiceId, err := network.AddApiService()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the person-modifying API service")
	}
	test.personModifyingApiServiceId = personModifyingApiServiceId

	personRetrievingApiServiceId, err := network.AddApiService()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the person-retrieving API service")
	}
	test.personRetrievingApiServiceId = personRetrievingApiServiceId

	return network, nil
}

func (test *AdvancedNetworkTest) Run(network networks.Network, testCtx testsuite.TestContext) {
	castedNetwork := network.(*networks_impl.TestNetwork)
	personModifier, err := castedNetwork.GetApiService(test.personModifyingApiServiceId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the person-modifying API service"))
	}
	personRetriever, err := castedNetwork.GetApiService(test.personRetrievingApiServiceId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the person-retrieving API service"))
	}

	logrus.Infof("Adding test person via person-modifying API service...")
	if err := personModifier.AddPerson(testPersonId); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred adding test person"))
	}
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read through person-modifying API service ...")
	if err := personModifier.IncrementBooksRead(testPersonId); err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred incrementing the number of books read"))
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read person-retrieving API service...")
	person, err := personRetriever.GetPerson(testPersonId)
	if err != nil {
		testCtx.Fatal(stacktrace.Propagate(err, "An error occurred getting the test person"))
	}
	logrus.Info("Retrieved test person")

	testCtx.AssertTrue(
		person.BooksRead == 1,
		stacktrace.NewError(
			"Expected number of books read to be incremented, but was '%v'",
			person.BooksRead,
		),
	)
}

func (test *AdvancedNetworkTest) GetTestConfiguration() testsuite.TestConfiguration {
	return testsuite.TestConfiguration{}
}


func (test *AdvancedNetworkTest) GetExecutionTimeout() time.Duration {
	return 60 * time.Second
}

func (test *AdvancedNetworkTest) GetSetupTimeout() time.Duration {
	return 60 * time.Second
}
