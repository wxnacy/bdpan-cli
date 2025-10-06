package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/handler"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录网盘",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// err := LoginCommand{}.Run()
		req := dto.NewLoginReq()
		req.GlobalReq = *GetGlobalReq()
		err := handler.GetAuthHandler().CmdLogin(req)
		handleCmdErr(err)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
