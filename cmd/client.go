package cmd

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rstms/mabctl/api"
	"github.com/rstms/rspamd-classes/classes"
	"github.com/spf13/viper"
)

type APIClient struct {
	Client *http.Client
	URL    string
}

type APIResponse struct {
	User    string
	Request string
	Message string
	Success bool
}

type APIClassesResponse struct {
	APIResponse
	Classes []classes.SpamClass
}

type APIClassResponse struct {
	APIResponse
	Class string
}

type APIBooksResponse struct {
	APIResponse
	Books []string
}

type APIAddressesResponse struct {
	APIResponse
	Addresses []any
}

type APIPasswordResponse struct {
	APIResponse
	Password string
}

type APIAccountsResponse struct {
	APIResponse
	Accounts map[string]string
}

type APIDumpResponse struct {
	APIResponse
	Classes  []classes.SpamClass
	Books    map[string]any
	Password string
}

type APIRestoreRequest struct {
	Username string
	Dump     api.ConfigDump
}

type APIUsageResponse struct {
	APIResponse
	Help     []string
	Commands []string
}

type APIVersionResponse struct {
	APIResponse
	Name    string
	Version string
	Classes string
	Mabctl  string
	UID     int
	GID     int
}

func GetViperPath(key string) (string, error) {
	path := viper.GetString(key)
	if len(path) < 2 {
		return "", fmt.Errorf("path %s too short: %s", key, path)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(home, path[1:])
	}
	return path, nil

}

func NewAPIClient() (*APIClient, error) {

	certFile, err := GetViperPath("cert")
	if err != nil {
		return nil, err
	}
	keyFile, err := GetViperPath("key")
	if err != nil {
		return nil, err
	}
	caFile, err := GetViperPath("ca")
	if err != nil {
		return nil, err
	}

	api := APIClient{
		URL: viper.GetString("server_url"),
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

func (a *APIClient) Get(path string, response interface{}) (string, error) {
	return a.request("GET", path, nil, response)
}

func (a *APIClient) Post(path string, request, response interface{}) (string, error) {
	return a.request("POST", path, request, response)
}

func (a *APIClient) Put(path string, response interface{}) (string, error) {
	return a.request("PUT", path, nil, response)
}

func (a *APIClient) Delete(path string, response interface{}) (string, error) {
	return a.request("DELETE", path, nil, response)
}

func (a *APIClient) request(method, path string, requestData, responseData interface{}) (string, error) {
	if viper.GetBool("verbose") {
		log.Printf("<-- %s %s", method, a.URL+path)
	}
	var requestBuffer io.Reader
	if requestData != nil {
		requestBytes, err := json.Marshal(requestData)
		if err != nil {
			return "", fmt.Errorf("failed marshalling JSON body for %s request: %v", method, err)
		}
		if viper.GetBool("verbose") {
			log.Printf("request: %s\n", string(requestBytes))
		}
		requestBuffer = bytes.NewBuffer(requestBytes)
	}
	request, err := http.NewRequest(method, a.URL+path, requestBuffer)
	if err != nil {
		return "", fmt.Errorf("failed creating %s request: %v", method, err)
	}
	response, err := a.Client.Do(request)
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
	err = json.Unmarshal(body, responseData)
	if err != nil {
		return "", fmt.Errorf("failed decoding JSON response: %v", err)
	}

	messageID, err := DecodedMessageID(viper.GetString("message_id"))
	if err != nil {
		return "", err
	}
	username := viper.GetString("sender")
	var text []byte

	switch t := responseData.(type) {
	case *APIResponse:
		var data *APIResponse
		data = responseData.(*APIResponse)
		data.Request = messageID
		data.User = username
		text, err = json.MarshalIndent(&data, "", "  ")
	case *APIClassesResponse:
		var data *APIClassesResponse
		data = responseData.(*APIClassesResponse)
		data.Request = messageID
		data.User = username
		text, err = json.MarshalIndent(&data, "", "  ")
	case *APIAddressesResponse:
		var data *APIAddressesResponse
		data = responseData.(*APIAddressesResponse)
		data.Request = messageID
		data.User = username
		text, err = json.MarshalIndent(&data, "", "  ")
	case *api.BooksResponse:
		var data *api.BooksResponse
		data = responseData.(*api.BooksResponse)
		data.User = username
		data.Request = messageID
		text, err = json.MarshalIndent(&data, "", "  ")
	case *APIDumpResponse:
		var data *APIDumpResponse
		data = responseData.(*APIDumpResponse)
		data.Request = messageID
		data.User = username
		text, err = json.MarshalIndent(&data, "", "  ")
	case *APIPasswordResponse:
		var data *APIPasswordResponse
		data = responseData.(*APIPasswordResponse)
		data.Request = messageID
		data.User = username
		text, err = json.MarshalIndent(&data, "", "  ")
	default:
		log.Fatalf("unknown type: %T\n", t)
	}

	if err != nil {
		return "", fmt.Errorf("failed formatting JSON response: %v", err)
	}

	return string(text), nil
}
