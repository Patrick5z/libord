## libord

Inscription indexer and validator.

Currently supported: brc-20, drc-20, ltc-20

## Building the source
1. You need to install the Go programming language environment.
1. Apply the ddl.sql file from the scripts directory to the MySQL database.
1. Modify the configurations in config.toml to match your own environment.

```shell
make ord-indexer
make ord-validator
```

Two binary files will be generated in the build directory.
```shell
./ord-indexer --help
./ord-indexer help run
./ord-validator --help
./ord-validator help run
```

## Run
You can use crontab to start the two processes independently without interfering with each other. ord-indexer is responsible for indexing data from blocks on the chain, while ord-validator verifies the legality of the data indexed by ord-indexer.

### Sample
```shell
./ord-indexer run --chain=btc --config=./config/config.toml >> ./logs/indexer-out.log 2>&1
./ord-validator run --config=./config/config.toml >> ./logs/validator-out.log 2>&1
```

If the --config parameter is not specified, it will default to looking for the ./config/config.toml file in the current directory.
