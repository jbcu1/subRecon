package cmd

import (
	"fmt"
	subs "subRecon/modules/subs"

	"github.com/spf13/cobra"
)

var subsCmd = &cobra.Command{
	Use:   "subs",
	Short: "Module 'subs' using for gathering subdomains from open sources, e.g. securitytrails, census etc",
	Run: func(cmd *cobra.Command, args []string) {
		domain, _ := cmd.Flags().GetString("domain")

		_, err := subs.SubdomainRaw(domain)

		if err != nil {
			fmt.Println(err)
			return
		}
	}}

func init() {
	rootCmd.AddCommand(subsCmd)
	subsCmd.Flags().String("domain", "", "Provide a root domain for gathering subdomains")
	subsCmd.MarkFlagRequired("url")
}
