/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package nginx_static

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-libs/golang/lib/services"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"net/http"
)

const (
	listenPort = 8080

	dockerImage = "flashspys/nginx-static"

	nginxStaticFilesDirpath = "/static"
)

/*
An Nginx service that serves files mounted in the /static directory
 */
type NginxStaticService struct {
	serviceId services.ServiceID
	ipAddr    string
}

func (n NginxStaticService) GetServiceID() services.ServiceID {
	return n.serviceId;
}

func (n NginxStaticService) GetIPAddress() string {
	return n.ipAddr;
}

func (n NginxStaticService) IsAvailable() bool {
	_, err := http.Get(fmt.Sprintf("%v:%v", n.ipAddr, listenPort))
	return err != nil
}

func (n NginxStaticService) GetFileContents(filename string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%v:%v/%v", n.ipAddr, listenPort, filename))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the contents of file '%v'", filename)
	}
	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body);
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body when getting the contents of file '%v'", filename)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}

