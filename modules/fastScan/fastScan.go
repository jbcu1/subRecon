package modules

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type IPData struct {
	Detail          string   `json:"detail"`
	CPES            []string `json:"cpes"`
	Hostnames       []string `json:"hostnames"`
	IP              string   `json:"ip"`
	Ports           []int    `json:"ports"`
	Tags            []string `json:"tags"`
	Vulnerabilities []string `json:"vulns"`
}

func FastScan(fileName string, numThreads int) ([]string, error) {

	scanResult := make([]string, 0)

	// Open the file
	fileRead, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return scanResult, err
	}
	defer fileRead.Close()

	// Create the file for writing
	fileWrite, err := os.Create("shodan.recon")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return scanResult, err
	}
	defer fileWrite.Close()

	writer := bufio.NewWriter(fileWrite)

	// Read the IP addresses from the file
	var ips []string
	scanner := bufio.NewScanner(fileRead)
	for scanner.Scan() {
		ips = append(ips, scanner.Text())
	}

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Set the number of goroutines to wait for
	wg.Add(len(ips))

	// Create a channel to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, numThreads)

	for _, ip := range ips {
		semaphore <- struct{}{} // Acquire a semaphore slot

		go func(ip string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the semaphore slot

			url := fmt.Sprintf("https://internetdb.shodan.io/%s", ip)

			// Make a GET request
			response, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error retrieving data for IP %s: %s\n", ip, err)
				return
			}
			defer response.Body.Close()
			fmt.Printf("Response status code for url %s is: %s\n", url, response.Status)

			// Read the response body
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("Error reading response body for IP %s: %s\n", ip, err)
				return
			}

			// Parse the JSON response
			var ipData IPData
			err = json.Unmarshal(body, &ipData)
			if err != nil {
				fmt.Printf("Error parsing JSON for IP %s: %s\n", ip, err)
				return
			}

			// Process the response
			if ipData.Detail == "No information available" {
				fmt.Printf("No information available for IP %s\n", ip)
			} else {
				for _, port := range ipData.Ports {

					strPort := strconv.Itoa(port)
					fmt.Printf("%s:%s\n", ip, strPort)
					formatIP := ip + ":" + strPort
					writer.WriteString(formatIP + "\n")
					//scanResult = append(scanResult, formatIP)
					//fmt.Fprintf(writer, "%s:%s\n", ip, strPort)
					if len(ipData.Hostnames) != 0 {
						for _, host := range ipData.Hostnames {
							formatHost := host + ":" + strPort
							//scanResult = append(scanResult, formatHost)
							writer.WriteString(formatHost + "\n")
							fmt.Printf("%s:%s\n", host, strPort)

						}
					}
				}

			}

		}(ip)

	}

	// Wait for all goroutines to finish
	wg.Wait()
	writer.Flush()

	return scanResult, err
}
