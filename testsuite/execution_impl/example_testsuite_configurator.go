/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution_impl

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/testsuite"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/testsuite_impl"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

type ExampleTestsuiteConfigurator struct {}

func NewExampleTestsuiteConfigurator() *ExampleTestsuiteConfigurator {
	return &ExampleTestsuiteConfigurator{}
}

func (t ExampleTestsuiteConfigurator) SetLogLevel(logLevelStr string) error {
	level, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing loglevel string '%v'", logLevelStr)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	return nil
}

func (t ExampleTestsuiteConfigurator) ParseParamsAndCreateSuite(paramsJsonStr string) (testsuite.TestSuite, error) {
	paramsJsonBytes := []byte(paramsJsonStr)
	var args SmartContractTestsuiteArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the testsuite params JSON")
	}

	if err := validateArgs(args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the deserialized testsuite params")
	}

	suite := testsuite_impl.NewSmartContractTestsuite(args.AvalancheImage)
	return suite, nil
}

func validateArgs(args SmartContractTestsuiteArgs) error {
	if strings.TrimSpace(args.AvalancheImage) == "" {
		return stacktrace.NewError("Avalanche image is empty")
	}
	return nil
}
