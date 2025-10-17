/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/handler"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/pkg/bdtools"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// text := "你好"
		// for _, v := range []string{"你好", "hellos"} {
		// fmt.Println(v, len(v), v[0], v[1], v[2])
		// fmt.Println(v, runewidth.StringWidth(v))
		// for _, s := range v {
		// fmt.Println(s, len(string(s)), unicode.Is(unicode.Han, s))
		// }
		// }
		// sText := "hello 你好"
		// textQuoted := strconv.QuoteToASCII(sText)
		// textUnquoted := textQuoted[1 : len(textQuoted)-1]
		// fmt.Println(textUnquoted)
		fsids := []uint64{842918256960809, 94896934453058, 873691114866544}
		selectFiles, err := bdtools.BatchGetFileInfos(handler.GetFileHandler().GetAccessToken(), fsids)
		files := make([]*model.File, 0)
		for _, file := range selectFiles {
			files = append(files, model.NewFile(file))
		}
		res, err := handler.GetFileHandler().BatchRenameFiles(files)
		handleCmdErr(err)
		fmt.Printf("%#v", res)
		if res.IsError() {
			fmt.Println(res.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
