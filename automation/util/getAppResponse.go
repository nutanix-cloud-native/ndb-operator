package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// Tests if pod is able to connect to database
func GetAppResponse(ctx context.Context, clientset *kubernetes.Clientset, pod *corev1.Pod, localPort string) (res http.Response, err error) {
	logger := GetLogger(ctx)
	logger.Println("GetAppResponse() started...")
	errBaseMsg := "GetAppResponse() ended"

	// Retrieve the pod name and targetPort
	podName := pod.Name
	podTargetPort := pod.Spec.Containers[0].Ports[0].ContainerPort

	// Run port-forward command using kubectl
	cmd := exec.Command("kubectl", "port-forward", podName, fmt.Sprintf("%s:%d", localPort, podTargetPort))
	err = cmd.Start()
	if err != nil {
		return http.Response{}, fmt.Errorf("%s! kubectl port-forward %s %s:%d failed! %v. ", errBaseMsg, podName, localPort, podTargetPort, err)
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

	logger.Println(fmt.Sprintf("%s!", errBaseMsg))

	return *resp, nil
}
