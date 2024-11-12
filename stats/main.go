package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwilteam/kwil-db/core/log"
)

var (
	statsDir string

	rpcServer string
)

func main() {
	// Flag support for stats file name, rpc servers to query.
	flag.StringVar(&statsDir, "output", ".stats", "stats directory to write stats.json and analysis.json files")
	flag.StringVar(&rpcServer, "rpcserver", "http://localhost:26657", "rpc server address to query stats from")

	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	logger := log.New(log.Config{
		Level:       log.InfoLevel.String(),
		OutputPaths: []string{"stdout"},
		Format:      log.FormatPlain,
		EncodeTime:  log.TimeEncodingEpochMilli, // for readability, log.TimeEncodingRFC3339Milli
	})

	statsMonitor, err := newStatsMonitor(rpcServer, statsDir, logger)
	if err != nil {
		logger.Error("failed to create stats monitor", log.Error(err))
		os.Exit(1)
	}

	err = statsMonitor.Run(signalChan)
	if err != nil {
		logger.Error("failed to run stats monitor", log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}
