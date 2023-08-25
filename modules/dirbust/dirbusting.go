package modules

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func parseCSV(file *os.File) {
	csvFiles, err := ioutil.ReadDir("ffuf")
	if err != nil {
		log.Fatal(err)
	}

	for _, csvFile := range csvFiles {
		if strings.HasSuffix(csvFile.Name(), ".csv") {
			csvPath := "ffuf/" + csvFile.Name()
			csvData, err := ioutil.ReadFile(csvPath)
			if err != nil {
				log.Fatal(err)
			}

			lines := strings.Split(string(csvData), "\n")
			for _, line := range lines[1:] {
				fields := strings.Split(line, ",")
				if len(fields) >= 7 {
					fmt.Fprintf(file, "%s %s %s\n", fields[2], fields[5], fields[6])
				}
			}
		}
	}

	fmt.Println("Dirs.txt created")

	// Clean up

	err = os.RemoveAll("ffuf")
	if err != nil {
		log.Fatal(err)
	}
}

//const maxWorkers = 5

func DirBustingResolved(filename string, wordlist string, timeExec string, workers string) ([]string, error) {

	err := os.Mkdir("ffuf", 0755)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Dir ffuf has been created")

	file, err := os.Create("dirs.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Fprintf(file, "URL ------------- STATUS_CODE ------------- CONTENT_LENGTH\n")

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	urls := make([]string, 0)

	scanner := bufio.NewScanner(f)

	logFile, err := os.OpenFile("journal.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error opening log file:", err)
		return urls, err
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	durationUnits, err := strconv.Atoi(timeExec)
	if err != nil {
		log.Fatal(err)
	}

	timeMeasure := time.Minute

	timeoutDuration := time.Duration(durationUnits) * timeMeasure

	maxWorkers, err := strconv.Atoi(workers)
	// Create a wait group to wait for all worker goroutines to finish
	var wg sync.WaitGroup

	// Create a buffered channel to control the worker pool
	workerPool := make(chan struct{}, maxWorkers)

	for scanner.Scan() {
		host := scanner.Text()
		url := host
		urls = append(urls, url)
		csvName := strings.Replace(strings.Replace(strings.Split(url, "//")[1], "/", "", 1), ".", "_", -1)

		workerPool <- struct{}{} // Add a worker to the pool

		wg.Add(1)
		go func() {
			defer func() {
				<-workerPool // Release a worker from the pool
				wg.Done()
			}()

			// Execute the command with timeout
			ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
			defer cancel()

			cmd := exec.Command("ffuf", "-u", host+"/FUZZ", "-w", wordlist, "-o", "ffuf/"+csvName+".csv", "-of", "csv", "-ac", "-r", "-t", "10")
			logger.Println("Executing command:", strings.Join(cmd.Args, " "))
			//cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Start()
			if err != nil {
				log.Fatal(err)
			}

			// Wait for the command to finish or the timeout to expire
			select {
			case <-ctx.Done():
				// Handle the timeout case
				log.Println("Command execution timed out")
				logger.Println("Command executed. With timeout:", strings.Join(cmd.Args, " "))
				// Kill the ongoing command
				if err := cmd.Process.Kill(); err != nil {
					log.Println("Failed to kill process:", err)
				}

			case err := <-wait(cmd):
				// Command finished executing
				logger.Println("Command executed:", strings.Join(cmd.Args, " "))
				if err != nil {
					log.Fatal(err)
				}
			}
		}()
	}

	// Wait for all worker goroutines to finish
	wg.Wait()

	parseCSV(file)
	return urls, nil
}

func wait(cmd *exec.Cmd) <-chan error {
	errc := make(chan error, 1)
	go func() {
		errc <- cmd.Wait()
		close(errc)
	}()
	return errc
}
