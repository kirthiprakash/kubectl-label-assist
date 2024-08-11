package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Table struct {
	Items []Item `json:"items"`
}

type Item struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Labels map[string]string `json:"labels"`
}

var cacheDir = ".cache"

type customRoundTripper struct {
	roundTripper http.RoundTripper
}

func (c *customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Accept", "application/json;as=Table;g=meta.k8s.io;v=v1")
	return c.roundTripper.RoundTrip(req)
}

func main() {
	resourceType := flag.String("resource", "pods", "Resource type (e.g., pods, deployments, sts, cm, nodes, services)")
	namespace := flag.String("namespace", "default", "Namespace to query, or 'all' for all namespaces (default is 'default')")
	flag.Parse()

	// Check if the namespace is set to "all" for all namespaces
	if *namespace == "all" {
		*namespace = ""
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		log.Fatalf("Error creating kubernetes client config: %v", err)
	}

	config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return &customRoundTripper{roundTripper: rt}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating kubernetes client: %v", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Error creating cache directory: %v", err)
	}

	apiPath := getAPIPath(*resourceType, *namespace)
	cachedResponse, exists := getCachedResponse(apiPath)
	var result []byte
	if exists {
		result = []byte(cachedResponse)
	} else {
		result, err = clientset.RESTClient().Get().RequestURI(apiPath).DoRaw(context.TODO())
		if err != nil {
			log.Fatalf("Error getting resource: %v", err)
		}
		setCachedResponse(apiPath, string(result))
	}

	var table Table
	if err := json.Unmarshal(result, &table); err != nil {
		log.Fatalf("Error unmarshaling response: %v", err)
	}

	for _, item := range table.Items {
		for key, value := range item.Metadata.Labels {
			fmt.Printf("%s=%s\n", key, value)
		}
	}
}

func getAPIPath(resourceType, namespace string) string {
	switch resourceType {
	case "pods", "pod", "po":
		return formatPath(namespace, "/pods")
	case "deployments", "deployment", "deploy":
		return formatPath(namespace, "/apis/apps/v1/deployments")
	case "sts", "statefulsets", "statefulset":
		return formatPath(namespace, "/apis/apps/v1/statefulsets")
	case "cm", "configmaps", "configmap":
		return formatPath(namespace, "/configmaps")
	case "nodes", "node":
		return "/api/v1/nodes"
	case "services", "service", "svc":
		return formatPath(namespace, "/services")
	default:
		return formatPath(namespace, fmt.Sprintf("/%s", resourceType))
	}
}

func formatPath(namespace, path string) string {
	if strings.HasPrefix(path, "/apis") {
		if namespace == "" {
			return path
		}
		parts := strings.Split(path, "/")
		if len(parts) < 5 {
			return path
		}
		return fmt.Sprintf("%s/namespaces/%s/%s", strings.Join(strings.Split(path, "/")[1:4], "/"), namespace, strings.Split(path, "/")[4])
	}

	if namespace == "" {
		return fmt.Sprintf("/api/v1%s", path)
	}
	return fmt.Sprintf("/api/v1/namespaces/%s%s", namespace, path)
}

func getCachedResponse(apiPath string) (string, bool) {
	escapedPath := escapeCacheKey(apiPath)
	cacheFilePath := filepath.Join(cacheDir, escapedPath)

	data, err := ioutil.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false
		}
		log.Printf("Error reading cache file: %v", err)
		return "", false
	}

	var cachedData struct {
		Timestamp time.Time
		Response  string
	}
	if err := json.Unmarshal(data, &cachedData); err != nil {
		log.Printf("Error unmarshaling cache data: %v", err)
		return "", false
	}

	if time.Since(cachedData.Timestamp) > 5*time.Minute {
		return "", false
	}

	return cachedData.Response, true
}

func setCachedResponse(apiPath, response string) {
	escapedPath := escapeCacheKey(apiPath)
	cacheFilePath := filepath.Join(cacheDir, escapedPath)

	cachedData := struct {
		Timestamp time.Time
		Response  string
	}{
		Timestamp: time.Now(),
		Response:  response,
	}

	data, err := json.Marshal(cachedData)
	if err != nil {
		log.Printf("Error marshaling cache data: %v", err)
		return
	}

	if err := ioutil.WriteFile(cacheFilePath, data, 0644); err != nil {
		log.Printf("Error writing cache file: %v", err)
	}
}

func escapeCacheKey(apiPath string) string {
	return strings.ReplaceAll(apiPath, "/", "_")
}
