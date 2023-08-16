package automation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	v1 "k8s.io/api/core/v1"
)

// Tests if 'manavrajvanshinx/best-app:latest' is able to connect to database
func GetAppResponse(ctx context.Context) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("TestAppConnectivity() started...")

	// Create appSvc template from automation.APP_SVC_PATH to get port number
	appSvc := &v1.Service{}
	if err := CreateTypeFromPath(appSvc, APP_SVC_PATH); err != nil {
		return http.Response{}, fmt.Errorf("GetAppResponse() ended! App Service with path %s failed! %v. ", APP_SVC_PATH, err)
	} else {
		logger.Printf("App Service with path %s created. ", APP_SVC_PATH)
	}

	// Send GET request
	client := http.Client{}
	resp, err := client.Get("http://localhost:" + fmt.Sprintf("%d", appSvc.Spec.Ports[0].Port))
	if err != nil {
		return http.Response{}, errors.New(fmt.Sprintf("GetAppResponse() ended! Error while performing GET: %s", err))
	}
	defer resp.Body.Close()
	logger.Println("Response status:", string(resp.Status))

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.Response{}, errors.New(fmt.Sprintf("GetAppResponse() ended! Error while reading response body: %s", err))
	}
	logger.Println("Response:", string(body))

	return res, nil
}
