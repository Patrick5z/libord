package main

import (
	"libord/config"
	"libord/internal/res"
	"libord/internal/validator"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var configPath string
	var chain string
	var ticks string
	var startBlock int64
	var endBlock int64

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "An validator is used to validate inscriptions",
		Long: `Validator is used to verify the validity of inscriptions. 
Some inscriptions may exhibit double-spending or insufficient balance issues.`,
		Run: func(cmd *cobra.Command, args []string) {
			config.Init(configPath)
			dbConfig := config.Instance().Mysql["app"]
			_db := res.GetDb(dbConfig.Host, dbConfig.Db, dbConfig.User, dbConfig.Password)
			defer _db.Close()

			_validator := &validator.Validator{Chain: chain, Db: _db}
			if _err := _validator.Run(); _err != nil {
				log.Fatalf("validator occur error:%+v", _err)
			}
		},
	}

	cmdRun.Flags().StringVarP(&chain, "chain", "n", "btc", "chain name,e.g:btc,ltc,doge")
	cmdRun.Flags().StringVarP(&configPath, "config", "c", "", "config file path")

	var cmdRevalidate = &cobra.Command{
		Use:   "revalidate",
		Short: "Revalidate all inscriptions",
		Long:  "Revalidate all inscription records in the database table from beginning.",
		Run: func(cmd *cobra.Command, args []string) {
			config.Init(configPath)
			dbConfig := config.Instance().Mysql["app"]
			_db := res.GetDb(dbConfig.Host, dbConfig.Db, dbConfig.User, dbConfig.Password)
			defer _db.Close()

			_validator := &validator.Validator{Chain: chain, Db: _db}
			if _err := _validator.Revalidate(startBlock, endBlock, strings.Split(ticks, ",")); _err != nil {
				log.Fatalf("validator occur error:%+v", _err)
			}
		},
	}
	cmdRevalidate.Flags().StringVarP(&chain, "chain", "n", "btc", "chain name,e.g:btc,ltc,doge")
	cmdRevalidate.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	cmdRevalidate.Flags().StringVarP(&ticks, "ticks", "t", "", "List of ticks that need to be revalidated, separated by commas.")
	cmdRevalidate.Flags().Int64VarP(&startBlock, "start", "s", 0, "start block height")
	cmdRevalidate.Flags().Int64VarP(&endBlock, "end", "e", 0, "end block height")

	var rootCmd = &cobra.Command{Use: "ord-validator"}
	rootCmd.AddCommand(cmdRun)
	rootCmd.AddCommand(cmdRevalidate)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Execute()
}
