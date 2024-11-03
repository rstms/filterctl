package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rstms/rspamd-classes/classes"
	"github.com/spf13/viper"
)

type APIClient struct {
	Client *http.Client
	URL    string
}

type APIResponse struct {
	Success bool
	Message string
	Classes []classes.SpamClass
}

func NewAPIClient() (*APIClient, error) {

	certFile := viper.GetString("cert")
	keyFile := viper.GetString("key")
	caFile := viper.GetString("ca")
	address := viper.GetString("address")
	port := viper.GetInt("port")

	api := APIClient{
		URL: fmt.Sprintf("https://%s:%d", address, port),
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("error loading client certificate pair: %v", err)
	}

	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("error loading certificate authority file: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		//RootCAs:      caCertPool,
	}
	api.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &api, nil
}

func (a *APIClient) Get(path string) (string, error) {
	return a.request("GET", path)
}

func (a *APIClient) Put(path string) (string, error) {
	return a.request("PUT", path)
}

func (a *APIClient) Delete(path string) (string, error) {
	return a.request("DELETE", path)
}

func (a *APIClient) request(method, path string) (string, error) {
	if viper.GetBool("verbose") {
		log.Printf("<-- %s %s", method, a.URL+path)
	}
	r, err := http.NewRequest(method, a.URL+path, nil)
	if err != nil {
		return "", fmt.Errorf("failed creating %s request: %v", method, err)
	}
	response, err := a.Client.Do(r)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failure reading response body: %v", err)
	}
	if response.StatusCode < 200 && response.StatusCode > 299 {
		return "", fmt.Errorf("API returned status [%d] %s", response.StatusCode, response.Status)
	}
	if viper.GetBool("verbose") {
		log.Printf("--> %v\n", string(body))
	}
	data := APIResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", fmt.Errorf("failed decoding JSON response: %v", err)
	}

	text, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed formatting JSON response: %v", err)
	}

	return string(text), nil
}
