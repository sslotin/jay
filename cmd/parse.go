package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse [url] [dir]",
	Short: "Fetches test data from an online judge",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		dir := "tests"
		if len(args) == 2 {
			dir = strings.TrimSuffix(args[1], "/")
		}
		os.MkdirAll(dir, 0700)
		response, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("Failed to perform HTTP request: %s", err)
		}
		defer response.Body.Close()
		if response.StatusCode != 200 {
			return fmt.Errorf("Status code error: %d %s", response.StatusCode, response.Status)
		}

		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return fmt.Errorf("Failed to parse html: %s", err)
		}

		doc.Find("pre").Each(func(i int, s *goquery.Selection) {
			fname := fmt.Sprintf("%s/%d.in", dir, i/2+1)
			if i%2 == 1 {
				fname += ".a"
			}
			fmt.Println(fname)
			file, err := os.Create(fname)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			file.WriteString(s.Text())
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}
