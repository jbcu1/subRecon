package cmd

import (
	"fmt"
	"strconv"
	FastScan "subRecon/modules/fastScan"

	"github.com/spf13/cobra"
)

var cmdFastScan = &cobra.Command{
	Use:   "fastscan",
	Short: "Module 'fastscan' using for initial fast scan provided list of IPs with shodan API",
	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := cmd.Flags().GetString("file")
		numThreadsStr := cmd.Flag("c").Value.String()

		// Convert the string value to an integer
		numThreads, err := strconv.Atoi(numThreadsStr)
		if err != nil {
			fmt.Println("Invalid value provided for 'c'")
			return
		}

		_, err = FastScan.FastScan(fileName, numThreads)

		if err != nil {
			fmt.Println(err)
			return
		}
	}}

func init() {
	rootCmd.AddCommand(cmdFastScan)
	cmdFastScan.Flags().String("file", "", "Provide a list with IPs for scan")
	cmdFastScan.MarkFlagRequired("file")
	cmdFastScan.Flags().String("c", "", "Provide a number of threads")
	cmdFastScan.MarkFlagRequired("c")
}
