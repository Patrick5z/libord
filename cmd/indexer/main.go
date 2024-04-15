package main

import (
	"libord/config"
	"libord/internal/indexer"
	"libord/internal/res"
	"libord/pkg/rpc"
	"log"

	"github.com/spf13/cobra"
)

func main() {
	var startBlock int64
	var endBlock int64
	var configPath string
	var chain string

	var cmdRun = &cobra.Command{
		Use:   "run",
		Short: "An indexer is used to index all inscriptions",
		Long: `An indexer is used to index all brc20/drc20/lrc20/[x]rc20 inscriptions.
regardless of whether the inscriptions are valid or not.`,
		Run: func(cmd *cobra.Command, args []string) {
			config.Init(configPath)
			dbConfig := config.Instance().Mysql["app"]
			_db := res.GetDb(dbConfig.Host, dbConfig.Db, dbConfig.User, dbConfig.Password)
			defer _db.Close()

			rpcConfig := config.Instance().Rpc[chain]
			_btc := &rpc.Btc{
				Chain:    chain,
				Url:      rpcConfig.Url,
				User:     rpcConfig.User,
				Password: rpcConfig.Password,
			}
			_indexer := &indexer.Indexer{Chain: chain, Db: _db, Rpc: _btc}
			if _err := _indexer.Run(startBlock, endBlock, config.Instance().MinConfirmation[chain]); _err != nil {
				log.Fatalf("indexer occur error:%+v", _err)
			}
		},
	}

	cmdRun.Flags().StringVarP(&chain, "chain", "n", "btc", "chain name,e.g:btc,ltc,doge")
	cmdRun.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	cmdRun.Flags().Int64VarP(&startBlock, "start", "s", 0, "start block height")
	cmdRun.Flags().Int64VarP(&endBlock, "end", "e", 0, "end block height")

	var rootCmd = &cobra.Command{Use: "ord-indexer"}
	rootCmd.AddCommand(cmdRun)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Execute()
}
