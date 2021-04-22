/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite_impl

import (
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/advanced_network_test"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/basic_datastore_and_api_test"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/basic_datastore_test"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/files_artifact_mounting_test"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/network_partition_test"
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

	// This example Go testsuite is used internally, when developing on Kurtosis Core, to verify functionality
	// When this testsuite is being used in this way, some special tests (which likely won't be interesting
	//  to you) are run
	// Feel free to delete these tests as you see fit
	if suite.isKurtosisCoreDevMode {
		tests["networkPartitionTest"] = network_partition_test.NewNetworkPartitionTest(
			suite.datastoreServiceImage,
			suite.apiServiceImage,
		)
		tests["filesArtifactMountingTest"] = files_artifact_mounting_test.FilesArtifactMountingTest{}
	}

	return tests
}

func (suite ExampleTestsuite) GetNetworkWidthBits() uint32 {
	return 8
}


