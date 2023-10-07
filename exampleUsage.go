package main

import (
	"bufio"
	_ "bufio"
	"encoding/json"
	"fmt"
	_ "log"
	_ "net/url"
	"os"
	_ "os"
	"os/exec"
	"runtime"
	"strings"
	_ "strings"
	"time"

	"github.com/mbndr/figlet4go"
	_ "github.com/mbndr/figlet4go"
)

var clear map[string]func() //create a map for storing clear funcs
func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

func main() {
	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render("CYBERNO")
	fmt.Print(renderStr)
	var serverAddress string
	fmt.Print("Please insert API server address [Default=https://multiscannerdemo.cyberno.ir/]: ")
	fmt.Scanln(&serverAddress)
	if serverAddress == "" {
		serverAddress = "https://multiscannerdemo.cyberno.ir/"
	}
	if strings.HasSuffix(serverAddress, "/") == false {
		serverAddress += "/"
	}
	var username string
	fmt.Print("Please insert identifier (email): ")
	fmt.Scanln(&username)
	var password string
	fmt.Print("Please insert your password: ")
	fmt.Scanln(&password)
	// Log in

	loginResponse := callWithJSONInput(serverAddress+"user/login", map[string]interface{}{"email": username, "password": password})
	checkResponseResult(loginResponse)
	apikey := loginResponse["data"].(string)
	//
	var index string
	fmt.Print("Please select scan mode:\n1- Scan local folder\n2- Scan file\nEnter Number=")
	fmt.Scanln(&index)
	fmt.Print()
	var scanResponse map[string]interface{}
	if index == "1" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please enter the paths of file to scan (with spaces): ")
		filePath, _ := reader.ReadString('\n')
		filePathArray := strings.Split(strings.TrimSpace(filePath), " ")

		fmt.Print("Enter the name of the selected antivirus (with spaces): ")
		avs, _ := reader.ReadString('\n')
		avsArray := strings.Split(strings.TrimSpace(avs), " ")

		scanResponse = callWithJSONInput(serverAddress+"scan/init", map[string]interface{}{"token": apikey, "avs": avsArray, "paths": filePathArray})
		checkResponseResult(scanResponse)
	} else {
		//Initialize scan
		var filePath string
		fmt.Print("Please enter the paths of file to scan: ")
		fmt.Scanln(&filePath)

		var avs string
		fmt.Print("Enter the name of the selected antivirus (with spaces): ")
		fmt.Scanln(&avs)

		dataInput := map[string]string{}
		dataInput["token"] = apikey
		dataInput["avs"] = avs
		scanResponse = callWithFormInput(serverAddress+"scan/multiscanner/init", dataInput, "file", filePath)
		checkResponseResult(scanResponse)
	}

	guid := scanResponse["guid"].(string)
	// Check Password  in Path Address
	if scanResponse["password_protected"] != nil {
		passwordProtected := scanResponse["password_protected"].([]interface{})
		for _, item := range passwordProtected {
			fmt.Print(fmt.Sprintf("|Enter the Password file -> %s |: ", item))
			var password string
			fmt.Scanln(&password)
			callWithJSONInput(fmt.Sprintf(serverAddress+"scan/extract/%s", guid), map[string]interface{}{
				"token":    apikey,
				"path":     item,
				"password": password,
			})
		}
	}

	fmt.Println("=========  Start Scan ===========")
	scanStartResponse := callWithJSONInput(fmt.Sprintf(serverAddress+"scan/start/%s", guid), map[string]interface{}{
		"token": apikey,
	})
	// Wait for scan results
	if scanStartResponse["success"].(bool) {
		isFinished := false
		for !isFinished {
			fmt.Println("Waiting for result...")
			scanResultResponse := callWithJSONInput(serverAddress+"scan/result/"+guid, map[string]interface{}{
				"token": apikey,
			})
			if scanResultResponse["data"].(map[string]interface{})["finished_at"] != nil {
				isFinished = true
				data, _ := json.MarshalIndent(scanResultResponse["data"], "", "    ")
				fmt.Println(string(data))
			}
			time.Sleep(5 * time.Second)
		}
	} else {
		fmt.Println(getError(scanStartResponse))
	}

}
