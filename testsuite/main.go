/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kurtosis-tech/avalanche-smart-contract-sample-testsuite/testsuite/execution_impl"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/execution"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	successExitCode = 0
	failureExitCode = 1
)

func main() {
	customParamsJsonArg := flag.String(
		"custom-params-json",
		"{}",
		"JSON string containing custom data that the testsuite will deserialize to modify runtime behaviour",
	)

	kurtosisApiSocketArg := flag.String(
		"kurtosis-api-socket",
		"",
		"Socket in the form of address:port of the Kurtosis API container",
	)

	logLevelArg := flag.String(
		"log-level",
		"",
		"String indicating the loglevel that the test suite should output with",
	)

	flag.Parse()

	configurator := execution_impl.NewSmartContractTestsuiteConfigurator()

	suiteExecutor := execution.NewTestSuiteExecutor(*kurtosisApiSocketArg, *logLevelArg, *customParamsJsonArg, configurator)
	if err := suiteExecutor.Run(context.Background()); err != nil {
		logrus.Errorf("An error occurred running the test suite executor:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}
