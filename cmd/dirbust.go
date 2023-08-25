package cmd

import (
	"fmt"
	dirbust "subRecon/modules/dirbust"

	"github.com/spf13/cobra"
)

var dirbCmd = &cobra.Command{
	Use:   "dirbusting",
	Short: "Module 'dirbust' fuzzing web directory for all specified urls",
	Run: func(cmd *cobra.Command, args []string) {
		urls, _ := cmd.Flags().GetString("urls")
		list, _ := cmd.Flags().GetString("list")
		duration, _ := cmd.Flags().GetString("duration")
		workers, _ := cmd.Flags().GetString("workers")

		_, err := dirbust.DirBustingResolved(urls, list, duration, workers)

		if err != nil {
			fmt.Println(err)
			return
		}
	}}

func init() {
	rootCmd.AddCommand(dirbCmd)
	dirbCmd.Flags().String("urls", "", "List contained specified urls for fuzz")
	dirbCmd.MarkFlagRequired("url")
	dirbCmd.Flags().String("list", "", "List with specified payloads for fuzz")
	dirbCmd.MarkFlagRequired("list")
	dirbCmd.Flags().String("duration", "", "Time execution for each dirbusting host, e.g. 10 minute, --duration 10")
	dirbCmd.MarkFlagRequired("duration")
	dirbCmd.Flags().String("workers", "", "Maximum concurrency of ffuf command, e.g, --workers 5")
	dirbCmd.MarkFlagRequired("workers")
}
