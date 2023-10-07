package main

import (
	"bytes"
	_ "bytes"
	"crypto/sha256"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	_ "net/http"
	"os"
	_ "strings"
)

const (
	USER_AGENT = "Cyberno-API-Sample-Golang"
)

func getSHA256(filePath string) string {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	hash := sha256.New()
	hash.Write([]byte(fileContent))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func getError(returnValue map[string]interface{}) string {
	error := "Error!\n"
	if _, ok := returnValue["error_code"].(float64); ok {
		error += fmt.Sprintf("Error code: %d\n", int(returnValue["error_code"].(float64)))
	}
	if _, ok := returnValue["error_desc"].(string); ok {
		error += fmt.Sprintf("Error description: %s\n", returnValue["error_desc"])
	}
	return error
}

func checkResponseResult(response map[string]interface{}) {
	if response["success"] == false {
		fmt.Println(response["error_code"])
		fmt.Println(response["error_desc"])
		os.Exit(0)
	}
	CallClear()
}

func callWithJSONInput(api string, jsonInput map[string]interface{}) map[string]interface{} {
	// Set up request headers
	headers := http.Header{
		"Content-Type": []string{"application/json"},
		"User-Agent":   []string{USER_AGENT},
	}

	// Set up request body
	jsonBody, err := json.Marshal(jsonInput)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return map[string]interface{}{"success": false, "error_code": 900}
	}

	// Send request to API
	req, err := http.NewRequest("POST", api, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return map[string]interface{}{"success": false, "error_code": 900}
	}
	req.Header = headers
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(jsonBody)))

	// Set up client
	client := &http.Client{}

	// Send request and get response
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return map[string]interface{}{"success": false, "error_code": 900}
	}
	defer resp.Body.Close()

	// Parse response body
	var values map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&values)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return map[string]interface{}{"success": false, "error_code": 900}
	}

	// Check for HTTP errors
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return map[string]interface{}{"success": false, "error_code": 900}
		}
		var values map[string]interface{}
		err = json.Unmarshal(data, &values)
		if err != nil {
			return map[string]interface{}{"success": false, "error_code": 900}
		}
		return map[string]interface{}{"success": false, "error_code": 900}
	}

	// Return response values
	return values
}

func callWithFormInput(api string, dataInput map[string]string, fileParamName string, filePath string) map[string]interface{} {
	fileHandle, err := os.Open(filePath)
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item1",
		}
	}
	defer fileHandle.Close()

	fileInfo, err := fileHandle.Stat()
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item2",
		}
	}

	file, err := ioutil.ReadAll(fileHandle)
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item3",
		}
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fileParamName, fileInfo.Name())
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item4",
		}
	}
	part.Write(file)

	for key, value := range dataInput {
		_ = writer.WriteField(key, value)
	}

	err = writer.Close()
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item5",
		}
	}

	request, err := http.NewRequest("POST", api, body)
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item6",
		}
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("User-Agent", USER_AGENT)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": err.Error(),
		}
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item8",
		}
	}

	var values map[string]interface{}
	err = json.Unmarshal(responseBody, &values)
	if err != nil {
		return map[string]interface{}{
			"success":    false,
			"error_code": "item9",
		}
	}

	return values
}
