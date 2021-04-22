/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package datastore

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"os"
)

const (
	port = 1323

	testVolumeMountpoint = "/test-volume"
)

type DatastoreContainerInitializer struct {
	dockerImage string
}

func NewDatastoreContainerInitializer(dockerImage string) *DatastoreContainerInitializer {
	return &DatastoreContainerInitializer{dockerImage: dockerImage}
}

func (d DatastoreContainerInitializer) GetDockerImage() string {
	return d.dockerImage
}

func (d DatastoreContainerInitializer) GetUsedPorts() map[string]bool {
	return map[string]bool{
		fmt.Sprintf("%v/tcp", port): true,
	}
}

func (d DatastoreContainerInitializer) GetServiceWrappingFunc() func (serviceId services.ServiceID, ipAddr string) services.Service {
	return func(serviceId services.ServiceID, ipAddr string) services.Service {
		return NewDatastoreService(serviceId, ipAddr, port);
	}
}

func (d DatastoreContainerInitializer) GetFilesToGenerate() map[string]bool {
	return map[string]bool{}
}

func (d DatastoreContainerInitializer) InitializeGeneratedFiles(filesToGenerate map[string]*os.File) error {
	return nil
}

func (d DatastoreContainerInitializer) GetFilesArtifactMountpoints() map[services.FilesArtifactID]string {
	return map[services.FilesArtifactID]string{}
}

func (d DatastoreContainerInitializer) GetTestVolumeMountpoint() string {
	return testVolumeMountpoint
}

func (d DatastoreContainerInitializer) GetStartCommand(mountedFileFilepaths map[string]string, ipPlaceholder string) ([]string, error) {
	// We have a launch command specified in the Dockerfile the datastore service was built with, so we
	//  don't explicitly specify one
	return nil, nil
}

