/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite_impl

import (
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/advanced_network_test"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/basic_datastore_and_api_test"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/basic_datastore_test"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
)

type ExampleTestsuite struct {
	apiServiceImage string
	datastoreServiceImage string
	isKurtosisCoreDevMode bool
}

func NewExampleTestsuite(apiServiceImage string, datastoreServiceImage string, isKurtosisCoreDevMode bool) *ExampleTestsuite {
	return &ExampleTestsuite{apiServiceImage: apiServiceImage, datastoreServiceImage: datastoreServiceImage, isKurtosisCoreDevMode: isKurtosisCoreDevMode}
}

func (suite ExampleTestsuite) GetTests() map[string]testsuite.Test {
	tests := map[string]testsuite.Test{
		"basicDatastoreTest": basic_datastore_test.NewBasicDatastoreTest(suite.datastoreServiceImage),
		"basicDatastoreAndApiTest": basic_datastore_and_api_test.NewBasicDatastoreAndApiTest(
			suite.datastoreServiceImage,
			suite.apiServiceImage,
		),
		"advancedNetworkTest": advanced_network_test.NewAdvancedNetworkTest(
			suite.datastoreServiceImage,
			suite.apiServiceImage,
		),
	}

	return tests
}

func (suite ExampleTestsuite) GetNetworkWidthBits() uint32 {
	return 8
}


