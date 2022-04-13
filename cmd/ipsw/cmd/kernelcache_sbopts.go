/*
Copyright © 2018-2022 blacktop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/blacktop/go-macho"
	"github.com/blacktop/ipsw/pkg/kernelcache"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

func init() {
	kernelcacheCmd.AddCommand(kernelSandboxCmd)
	kernelSandboxCmd.Flags().BoolP("diff", "d", false, "Diff two kernel's sandbox operations")
	kernelSandboxCmd.MarkZshCompPositionalArgumentFile(1, "kernelcache*")
}

// kernelSandboxCmd represents the kernelSandboxCmd command
var kernelSandboxCmd = &cobra.Command{
	Use:   "sbopts",
	Short: "List kernel sandbox operations",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		diff, _ := cmd.Flags().GetBool("diff")

		kcPath := filepath.Clean(args[0])

		if _, err := os.Stat(kcPath); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", args[0])
		}

		m, err := macho.Open(kcPath)
		if err != nil {
			return err
		}
		defer m.Close()

		sbOpts, err := kernelcache.GetSandboxOpts(m)
		if err != nil {
			return err
		}

		if diff {
			if len(args) < 2 {
				return fmt.Errorf("please provide two kernelcache files to diff")
			}

			kcPath2 := filepath.Clean(args[1])

			if _, err := os.Stat(kcPath2); os.IsNotExist(err) {
				return fmt.Errorf("file %s does not exist", args[1])
			}

			m2, err := macho.Open(kcPath2)
			if err != nil {
				return err
			}
			defer m2.Close()

			sbOpts2, err := kernelcache.GetSandboxOpts(m2)
			if err != nil {
				return err
			}

			sb1OUT := fmt.Sprintf("Sandbox Operations (%d)\n", len(sbOpts))
			sb1OUT += fmt.Sprintln(strings.Repeat("=", len(sb1OUT)))
			for _, opt := range sbOpts {
				sb1OUT += fmt.Sprintln(opt)
			}

			sb2OUT := fmt.Sprintf("Sandbox Operations (%d)\n", len(sbOpts2))
			sb2OUT += fmt.Sprintln(strings.Repeat("=", len(sb2OUT)))
			for _, opt := range sbOpts2 {
				sb2OUT += fmt.Sprintln(opt)
			}

			dmp := diffmatchpatch.New()

			diffs := dmp.DiffMain(sb1OUT, sb2OUT, true)
			if len(diffs) == 1 {
				if diffs[0].Type == diffmatchpatch.DiffEqual {
					log.Info("No differences found")
				}
			} else {
				fmt.Println(dmp.DiffPrettyText(diffs))
			}
		} else {
			title := fmt.Sprintf("Sandbox Operations (%d)", len(sbOpts))
			fmt.Println(title)
			fmt.Println(strings.Repeat("=", len(title)))
			for _, opt := range sbOpts {
				fmt.Println(opt)
			}
		}

		return nil
	},
}
