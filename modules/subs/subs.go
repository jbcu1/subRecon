package modules

import (
	"fmt"

	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func createFilename(domain string) (string, error) {
	file := strings.Split(domain, ".")[0]
	fileName := file + ".txt"
	err := ioutil.WriteFile(fileName, []byte(""), 0644)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func findDomainSubfinder(domain string, wg *sync.WaitGroup, logger *log.Logger) {
	defer wg.Done()
	logger.Println("Executing subfinder")
	subfinderCmd := exec.Command("subfinder", "-d", domain, "-pc", "/Users/qwerty/pentesting/configs/subfinder/provider-config.yaml", "-all", "-nc", "-o", "subfinder.result", "-v")
	err := executeCommand(subfinderCmd, logger)
	if err != nil {
		logger.Println("Error executing subfinder command:", err)
		return
	}
}
func findDomainAssetfinder(domain string, wg *sync.WaitGroup, logger *log.Logger) {
	defer wg.Done()

	logger.Println("Executing assetfinder")
	assetfinderCmd := exec.Command("assetfinder", "--subs-only", domain)

	output, err := assetfinderCmd.Output()
	//err = executeCommand(assetfinderCmd, logger)
	//logger.Println("Executing command:", strings.Join(exec.Cmd.Args, " "))
	if err != nil {
		logger.Println("Error executing assetfinder command:", err)
		return
	}

	err = ioutil.WriteFile("assetfinder.result", output, 0644)
	if err != nil {
		logger.Println("Error writing assetfinder result to file:", err)
		return
	}

}
func findDomainFindomain(domain string, wg *sync.WaitGroup, logger *log.Logger) {
	defer wg.Done()
	logger.Println("Executing findomain")
	findomainCmd := exec.Command("findomain", "-t", domain, "-u", "findomain.result")
	err := executeCommand(findomainCmd, logger)
	if err != nil {
		logger.Println("Error executing findomain command:", err)
		return
	}

}
func findDomainGithub(domain string, wg *sync.WaitGroup, logger *log.Logger) {
	defer wg.Done()
	logger.Println("Executing github-subdomains")
	githubSubdomainsCmd := exec.Command("github-subdomains", "-d", domain, "-o", "github-subs.result")
	err := executeCommand(githubSubdomainsCmd, logger)
	if err != nil {
		logger.Println("Error executing github-subdomains command:", err)
		return
	}

}

func findDomainChaos(domain string, wg *sync.WaitGroup, logger *log.Logger) {
	defer wg.Done()
	logger.Println("Executing Chaos")
	chaosCmd := exec.Command("chaos", "-d", domain, "-verbose", "-o", "chaos.result", "-key", os.Getenv("CHAOS_KEY"))
	err := executeCommand(chaosCmd, logger)
	if err != nil {
		logger.Println("Error executing chaos command:", err)
		return
	}

}

func executeCommand(cmd *exec.Cmd, logger *log.Logger) error {
	logger.Println("Executing command:", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func filterResult(fileName string) error {
	fmt.Printf("Created file with name %s\n", fileName)

	subfinderResult, err := ioutil.ReadFile("subfinder.result")
	if err != nil {
		return err
	}

	findomainResult, err := ioutil.ReadFile("findomain.result")
	if err != nil {
		return err
	}

	chaosResult, err := ioutil.ReadFile("chaos.result")
	if err != nil {
		return err
	}

	assetfinderResult, err := ioutil.ReadFile("assetfinder.result")
	if err != nil {
		return err
	}

	githubSubsResult, err := ioutil.ReadFile("github-subs.result")
	if err != nil {
		return err
	}

	allResults := strings.Join([]string{string(githubSubsResult), string(subfinderResult), string(chaosResult), string(findomainResult), string(assetfinderResult)}, "\n")
	uniqueResults := removeDuplicates(strings.Split(allResults, "\n"))

	err = ioutil.WriteFile(fileName, []byte(strings.Join(uniqueResults, "\n")), 0644)
	if err != nil {
		return err
	}

	err = cleanupFiles("subfinder.result", "findomain.result", "chaos.result", "assetfinder.result", "github-subs.result")
	if err != nil {
		return err
	}

	return nil
}

func removeDuplicates(slice []string) []string {
	encountered := map[string]bool{}
	result := []string{}
	for _, item := range slice {
		if !encountered[item] {
			encountered[item] = true
			result = append(result, item)
		}
	}
	return result
}

func cleanupFiles(filenames ...string) error {
	for _, filename := range filenames {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func SubdomainRaw(domain string) ([]string, error) {
	allSubs := make([]string, 0)

	fileName, err := createFilename(domain)
	if err != nil {
		log.Fatal(err)
		return allSubs, err
	}

	logFile, err := os.OpenFile("journal.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error opening log file:", err)
		return allSubs, err
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)

	var wg sync.WaitGroup

	fmt.Println("Executing Subfinder.")
	wg.Add(1)
	go findDomainSubfinder(domain, &wg, logger)

	fmt.Println("Executing Assetfinder.")
	wg.Add(1)
	go findDomainAssetfinder(domain, &wg, logger)

	fmt.Println("Executing Chaos.")
	wg.Add(1)
	go findDomainChaos(domain, &wg, logger)

	fmt.Println("Executing Findomain.")
	wg.Add(1)
	go findDomainFindomain(domain, &wg, logger)

	fmt.Println("Executing Github-subdomains.")
	wg.Add(1)
	go findDomainGithub(domain, &wg, logger)

	wg.Wait()

	err = filterResult(fileName)
	if err != nil {
		log.Fatal(err)
		return allSubs, err
	}

	return allSubs, nil
}
