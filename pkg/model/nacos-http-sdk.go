package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/nacos-group/nacos-k8s-sync/pkg/logger"
	"io"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/http"
	"net/url"
	"os"
)

type NacosHttpSdk struct {
	nacosAddresses []string
	nacosNamespace string
	accessKey      string
	secretKey      string
}

type ServiceDetail struct {
	NamespaceId      string  `json:"namespaceId"`
	GroupName        string  `json:"groupName"`
	Name             string  `json:"name"`
	ProtectThreshold float64 `json:"protectThreshold"`
	Metadata         struct {
	} `json:"metadata"`
	Selector struct {
		Type        string `json:"type"`
		ContextType string `json:"contextType"`
	} `json:"selector"`
	Clusters []struct {
		Name          string `json:"name"`
		HealthChecker struct {
			Type                    string `json:"type"`
			TimeoutMs               int    `json:"timeoutMs"`
			InternalMs              int    `json:"internalMs"`
			HealthyCheckThreshold   int    `json:"healthyCheckThreshold"`
			UnhealthyCheckThreshold int    `json:"unhealthyCheckThreshold"`
		} `json:"healthChecker"`
		Metadata struct {
		} `json:"metadata"`
	} `json:"clusters"`
}

func NewNacosHttpSdk(options NacosOptions) *NacosHttpSdk {
	return &NacosHttpSdk{
		nacosAddresses: options.ServersIP,
		nacosNamespace: options.Namespace,
		accessKey:      os.Getenv("NACOS_ACCESS_KEY"),
		secretKey:      os.Getenv("NACOS_SECRET_KEY"),
	}
}

func (n *NacosHttpSdk) UpdateServiceHealthCheckTypeToNone(key ServiceKey) bool {
	svc, err := n.getService(key.ServiceName)
	if err != nil {
		logger.Error("failed to get service from nacos, service name: " + key.ServiceName)
		return false
	}

	for _, cluster := range svc.Clusters {
		if cluster.Name == "DEFAULT" {
			if cluster.HealthChecker.Type == "NONE" {
				return true
			}
		}
	}

	params := url.Values{}
	params.Add("namespaceId", n.nacosNamespace)
	params.Add("serviceName", key.ServiceName)
	params.Add("clusterName", "DEFAULT")
	params.Add("checkPort", "80")
	params.Add("useInstancePort4Check", "true")
	params.Add("healthChecker", "{\"type\":\"NONE\"}")
	requestBody := params.Encode()
	bodyReader := bytes.NewBufferString(requestBody)

	url := "http://" + n.nacosAddresses[rand.Intn(len(n.nacosAddresses))] +
		"/nacos/v1/ns/cluster"

	req, err := http.NewRequest("PUT", url, bodyReader)
	if err != nil {
		logger.Error("failed to create request to nacos, service name: " + key.ServiceName)
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to send request to nacos, service name: " + key.ServiceName)
		return false
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.Error("failed to read body from nacos response, url: " + url)
		return false
	}

	if resp.StatusCode != 200 {
		logger.Error("failed to update service in nacos, service name: " + key.ServiceName + ", url: " + url + " status code: " + resp.Status + ", response: " + string(bytes))
		return false
	}

	return true
}

func (n *NacosHttpSdk) getService(serviceName string) (ServiceDetail, error) {
	url := "http://" + n.nacosAddresses[rand.Intn(len(n.nacosAddresses))] + "/nacos/v1/ns/service?serviceName=" + serviceName + "&namespaceId=" + n.nacosNamespace

	resp, err := http.Get(url)

	if err != nil {
		logger.Error("failed to get service from nacos, service name: " + serviceName)
		return ServiceDetail{}, err
	}

	bytes, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		logger.Error("failed to read body from nacos response, url: " + url)
		return ServiceDetail{}, err
	}

	if resp.StatusCode != 200 {
		logger.Error("failed to get service from nacos, service name: " + serviceName + " status code: " + resp.Status + ", response: " + string(bytes))
		return ServiceDetail{}, errors.New("failed to get service from nacos, code: " + resp.Status)
	}
	var serviceDetail ServiceDetail
	err = json.Unmarshal(bytes, &serviceDetail)

	if err != nil {
		logger.Error("failed to unmarshal nacos response, url: " + url)
		return ServiceDetail{}, err
	}

	return serviceDetail, nil
}
