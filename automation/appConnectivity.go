package automation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Tests if 'manavrajvanshinx/best-app:latest' is able to connect to database
func GetAppResponse(ctx context.Context, clientset *kubernetes.Clientset, localPort string) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("TestAppConnectivity() started...")

	// Create appPod template to retrieve the pod name and targetPort
	appPod := &v1.Pod{}
	if err := CreateTypeFromPath(appPod, APP_POD_PATH); err != nil {
		return http.Response{}, fmt.Errorf("GetAppResponse() ended! App Pod with path %s failed! %v. ", APP_POD_PATH, err)
	} else {
		logger.Printf("App Pod with path %s created. ", APP_POD_PATH)
	}

	// Retrieve te pod name and targetPort
	podName := appPod.Name
	podTargetPort := appPod.Spec.Containers[0].Ports[0].ContainerPort

	// Run port-forward command using kubectl
	cmd := exec.Command("kubectl", "port-forward", podName, fmt.Sprintf("%s:%d", localPort, podTargetPort))
	err = cmd.Start()
	if err != nil {
		return http.Response{}, fmt.Errorf("'kubectl port-forward %s %s:%d' failed! %v. ", podName, localPort, podTargetPort, err)
	} else {
		logger.Printf("kubectl port-forward %s %s:%d succesful.", podName, localPort, podTargetPort)
	}

	// Wait for a brief period to let port-forwarding start
	time.Sleep(2 * time.Second)

	// Verify the forwarded port by making an HTTP request
	url := fmt.Sprintf("http://localhost:%s", localPort)
	resp, err := http.Get(url)

	if err != nil {
		return http.Response{}, fmt.Errorf("http://localhost:%s failed! %v,", localPort, err)
	} else {
		logger.Printf("http://localhost:%s succesful.", localPort)
	}

	defer resp.Body.Close()

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.Response{}, errors.New(fmt.Sprintf("GetAppResponse() ended! Error while reading response body: %s", err))
	} else {
		logger.Println("Response: ", string(body))
	}

	return *resp, nil
}
