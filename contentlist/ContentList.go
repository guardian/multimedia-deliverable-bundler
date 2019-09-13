package contentlist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ContentList struct {
	FileId    string `json:"fileId"`
	StorageId string `json:"storageId"`
}

/**
read and close the HTTP body
*/
func readBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	rtn, readErr := ioutil.ReadAll(response.Body)

	return rtn, readErr
}

func handleResponse(response *http.Response) ([]byte, error) {
	switch response.StatusCode {
	case 200:
		return readBody(response)
	case 400:
		body, readErr := readBody(response)
		if readErr != nil {
			return nil, readErr
		} else {
			bodyString := string(body)
			errMsg := fmt.Sprintf("API returned bad data: %s", bodyString)
			return nil, errors.New(errMsg)
		}
	case 403:
		body, readErr := readBody(response)
		if readErr != nil {
			return nil, readErr
		} else {
			bodyString := string(body)
			errMsg := fmt.Sprintf("API returned permission denied: %s", bodyString)
			return nil, errors.New(errMsg)
		}
	case 500:
	case 502:
	case 503:
	case 504:
		body, readErr := readBody(response)
		if readErr != nil {
			return nil, readErr
		} else {
			bodyString := string(body)
			errMsg := fmt.Sprintf("API returned not available: %s", bodyString)
			return nil, errors.New(errMsg)
		}
	default:
		body, readErr := readBody(response)
		if readErr != nil {
			return nil, readErr
		} else {
			bodyString := string(body)
			errMsg := fmt.Sprintf("API returned unexpected status %d: %s", response.StatusCode, bodyString)
			return nil, errors.New(errMsg)
		}
	}
	return nil, errors.New("code bug, should not reach this point")
}

/**
get content list data from either a file:// or an http:// URL
*/
func GetContentList(uriString string, token string) ([]ContentList, error) {
	uriData, err := url.Parse(uriString)

	if err != nil {
		return nil, err
	}

	if uriData.Scheme == "file" {
		return GetFileContentList(uriData.Path)
	} else {
		return DownloadContentList(uriString, token)
	}
}

/**
get content list data from a file. This is automaticlaly called by GetContentList for a file:// URL
*/
func GetFileContentList(path string) ([]ContentList, error) {
	var rtn []ContentList

	fileContent, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}

	parseErr := json.Unmarshal(fileContent, rtn)
	if parseErr != nil {
		return nil, parseErr
	} else {
		return rtn, nil
	}
}

/**
download the given URL and parse it as JSON into an array of ContentList objects. This is automatically called by GetContentList
*/
func DownloadContentList(uri string, token string) ([]ContentList, error) {
	client := http.Client{}
	var contentList []ContentList
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Authentication-Token", token)

	for {
		response, doErr := client.Do(req)
		if doErr != nil {
			return nil, doErr
		}

		bodyContent, responseErr := handleResponse(response)

		if responseErr != nil {
			return nil, responseErr
		}

		if response.StatusCode == 502 || response.StatusCode == 503 {
			log.Printf("Got a server unavailable error, retrying in 3s...")
			time.Sleep(3 * time.Second)
		} else {
			jsonErr := json.Unmarshal(bodyContent, contentList)
			if jsonErr != nil {
				log.Printf("Could not understand response from server: %s", jsonErr.Error())
				return nil, jsonErr
			} else {
				return contentList, nil
			}
		}
	}
}
