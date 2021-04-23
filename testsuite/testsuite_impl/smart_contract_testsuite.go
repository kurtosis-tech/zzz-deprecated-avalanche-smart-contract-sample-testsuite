/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite_impl

import (
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl/smart_contract_test"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
)

type SmartContractTestsuite struct {
	avalancheImage string
}

func NewSmartContractTestsuite(avalancheImage string) *SmartContractTestsuite {
	return &SmartContractTestsuite{avalancheImage: avalancheImage}
}

func (suite SmartContractTestsuite) GetTests() map[string]testsuite.Test {
	tests := map[string]testsuite.Test{
		"smartContractTest": smart_contract_test.NewSmartContractTest(suite.avalancheImage),
	}

	return tests
}

func (suite SmartContractTestsuite) GetNetworkWidthBits() uint32 {
	return 8
}


