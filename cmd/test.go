/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// text := "你好"
		for _, v := range []string{"你好", "hellos"} {
			fmt.Println(v, len(v), v[0], v[1], v[2])
			fmt.Println(v, runewidth.StringWidth(v))
			for _, s := range v {
				fmt.Println(s, len(string(s)), unicode.Is(unicode.Han, s))
			}
		}
		sText := "hello 你好"
		textQuoted := strconv.QuoteToASCII(sText)
		textUnquoted := textQuoted[1 : len(textQuoted)-1]
		fmt.Println(textUnquoted)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
