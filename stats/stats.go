package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"

	"github.com/kwilteam/kwil-db/core/log"
)

// StatsMonitor polls all the rpc servers in the network to collect the
// mempool stats throughout the test duration.
// Once stats monitor is stopped (ctrl-c), it retrieves all the blocks
// mined during the test duration and analyzes the block info and
// generate basic throughput metrics.

type rpcClient struct {
	// Client for the RPC server
	client *rpchttp.HTTP

	address string
}

type statsMonitor struct {
	// Client for the RPC server
	client *rpcClient
	// statsDir to save the stats
	statsDir string

	// Metrics
	stats    *stats
	analysis *derivedMetrics

	// Logger
	log log.Logger
}

type mempoolData struct {
	UnconfirmedTxs      int   `json:"unconfirmed-txs"`
	UnconfirmedTxsBytes int64 `json:"unconfirmed-tx-bytes"`
}

type blockData struct {
	Height    int64     `json:"height"`
	Time      time.Time `json:"time"`
	TxCount   int64     `json:"tx-count"`
	BlockSize int64     `json:"block-size"`
	Rounds    int32     `json:"rounds"`

	// Mempool data
	UnconfirmedTxs      int   `json:"unconfirmed-txs"`
	UnconfirmedTxsBytes int64 `json:"unconfirmed-tx-bytes"`
}

type stats struct {
	// Test Boundaries
	StartBlock int64 `json:"start-block"`
	EndBlock   int64 `json:"end-block"`

	// Per Client Mempool Stats
	MempoolStats map[int64]mempoolData `json:"mempool-stats"`

	// Block data
	Blocks map[int64]blockData `json:"blocks"`

	// Metrics to be calculated
	BlockRate                float64 `json:"blockRate"`
	TransactionRate          float64 `json:"transactionRate"`
	TransactionCountPerBlock float64 `json:"transactionCountPerBlock"`
	PayloadRate              float64 `json:"payloadRate"`
	PayloadSizePerBlock      float64 `json:"payloadSizePerBlock"`

	MempoolTxRate         float64 `json:"mempoolTxRate"`
	MempoolTxSizeRate     float64 `json:"mempoolTxSizeRate"`
	MempoolTxPerBlock     float64 `json:"mempoolTxPerBlock"`
	MempoolTxSizePerBlock float64 `json:"mempoolTxSizePerBlock"`
}

func (s *statsMonitor) saveAs(statsDir string) error {
	bts, err := json.MarshalIndent(s.stats, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(statsDir, "stats.json"), bts, 0644)
	if err != nil {
		return err
	}

	bts, err = json.MarshalIndent(s.analysis, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(statsDir, "analysis.json"), bts, 0644)
}

func newStatsMonitor(address string, statsDir string, log log.Logger) (*statsMonitor, error) {
	os.MkdirAll(statsDir, 0755)
	s := &statsMonitor{
		log:      log,
		statsDir: statsDir,
		stats: &stats{
			Blocks:       make(map[int64]blockData),
			MempoolStats: make(map[int64]mempoolData),
		},
	}

	client, err := rpchttp.New(address, "/websocket")
	if err != nil {
		return nil, err
	}
	s.client = &rpcClient{
		client:  client,
		address: address,
	}

	return s, nil
}

func (s *statsMonitor) Run(signalChan chan os.Signal) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)

	select {
	case <-signalChan:
		s.log.Info("Received signal to stop the stats monitor")
	case <-ctx.Done():
		s.log.Info("Context is done")
		cancel()
	}

	err := s.retrieveMetrics()
	if err != nil {
		return err
	}

	s.analyze()

	err = s.saveAs(s.statsDir)
	if err != nil {
		return err
	}

	return nil
}

func (s *statsMonitor) Start(ctx context.Context) {
	s.log.Info("Starting the stats monitor")
	// Record the block height at the start of the test
	res, err := s.client.client.Status(ctx)
	if err != nil {
		s.log.Error("Failed to get the chain status", log.Error(err))
		return
	}

	s.stats.StartBlock = res.SyncInfo.LatestBlockHeight

	// Start the polling routine to keep track of the unconfirmed transactions and the block height
	// for each validator node in the network (as different nodes might have different Txs in the mempool)
	client := s.client
	s.stats.MempoolStats = make(map[int64]mempoolData)
	s.log.Info("Launching the polling routine for the client", log.String("address", client.address))
	go func(rc *rpcClient) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				mstats := s.stats.MempoolStats
				// Get current block height
				status, err := rc.client.Status(ctx)
				if err != nil {
					s.log.Error("Failed to get the chain status", log.Error(err))
					return
				}
				height := status.SyncInfo.LatestBlockHeight

				res, err := rc.client.NumUnconfirmedTxs(ctx)
				if err != nil {
					s.log.Error("Failed to get the unconfirmed transactions", log.Error(err))
					return
				}
				stats := mempoolData{
					UnconfirmedTxs:      res.Total,
					UnconfirmedTxsBytes: res.TotalBytes,
				}

				val, ok := mstats[height]
				if !ok {
					s.stats.MempoolStats[height] = stats
				} else {
					// As we are polling every second, mempool might have accumulated more Txs
					// since the last poll so we need to update the stats. There is no way that the
					// tx count will decrease within a block height, as mempool operations are atomic
					// within block boundaries.
					if stats.UnconfirmedTxs > val.UnconfirmedTxs {
						s.stats.MempoolStats[height] = stats
					}
				}
				time.Sleep(1 * time.Second)
			}
		}
	}(client)
}

func (s *statsMonitor) retrieveMetrics() error {
	ctx := context.Background()
	s.log.Info("Retrieving the metrics")
	// Record the block height at the end of the test
	res, err := s.client.client.Status(ctx)
	if err != nil {
		return err
	}
	s.stats.EndBlock = res.SyncInfo.LatestBlockHeight

	for i := s.stats.StartBlock; i <= s.stats.EndBlock; i++ {
		block, err := s.client.client.Block(ctx, &i)
		if err != nil {
			return err
		}

		sz := txSize(block.Block.Txs.ToSliceOfBytes())
		val := blockData{
			Height:    i,
			Time:      block.Block.Header.Time,
			TxCount:   int64(len(block.Block.Txs)),
			BlockSize: sz,
		}

		res, ok := s.stats.MempoolStats[i]
		if ok {
			val.UnconfirmedTxs = maxInt(res.UnconfirmedTxs, val.UnconfirmedTxs)
			val.UnconfirmedTxsBytes = maxInt64(res.UnconfirmedTxsBytes, val.UnconfirmedTxsBytes)
		}

		s.stats.Blocks[i] = val

		// Fetch and update the round number for the previous block
		round := block.Block.LastCommit.Round
		if i != s.stats.StartBlock {
			prevBlock := s.stats.Blocks[i-1]
			prevBlock.Rounds = round
			s.stats.Blocks[i-1] = prevBlock
		}
	}

	return nil
}

func (s *statsMonitor) analyze() {
	startTime := s.stats.Blocks[s.stats.StartBlock].Time
	endTime := s.stats.Blocks[s.stats.EndBlock].Time
	testDuration := endTime.Sub(startTime).Minutes()

	// Calculate the block rate
	totalBlocks := int64(len(s.stats.Blocks))
	// s.stats.BlockRate = float64(totalBlocks) / testDuration

	// Calculate the transaction rate
	var totalTxs, totalPayloadSz, totalBytes, totalCount int64

	for _, block := range s.stats.Blocks {
		totalTxs += block.TxCount
		totalPayloadSz += block.BlockSize
		totalBytes += block.BlockSize
		totalCount += int64(block.UnconfirmedTxs)
	}

	s.log.Info("Analyze metrics", log.Float("test-duration(min)", testDuration), log.Int("start-block", s.stats.StartBlock), log.Int("end-block", s.stats.EndBlock), log.Int("total-blocks", totalBlocks), log.Int("total-txs", totalTxs), log.Int("total-block-size", totalPayloadSz), log.Int("total-mempool-tx-size", totalBytes), log.Int("total-mempool-txs", totalCount))

	s.stats.TransactionRate = float64(totalTxs) / testDuration
	s.stats.TransactionCountPerBlock = float64(totalTxs) / float64(totalBlocks)
	s.stats.PayloadRate = float64(totalPayloadSz) / testDuration
	s.stats.PayloadSizePerBlock = float64(totalPayloadSz) / float64(totalBlocks)
	s.stats.MempoolTxRate = float64(totalCount) / testDuration
	s.stats.MempoolTxSizeRate = float64(totalBytes) / testDuration
	s.stats.MempoolTxPerBlock = float64(totalCount) / float64(totalBlocks)
	s.stats.MempoolTxSizePerBlock = float64(totalBytes) / float64(totalBlocks)

	// fmt.Println("Mempool Txs Count:", totalCount, "  Rate: ", s.stats.MempoolTxRate, " avg per block: ", s.stats.MempoolTxPerBlock)

	s.analyzeStats()
}

func (s *statsMonitor) analyzeStats() {
	startTime := s.stats.Blocks[s.stats.StartBlock].Time
	endTime := s.stats.Blocks[s.stats.EndBlock].Time
	testDuration := endTime.Sub(startTime).Minutes()

	heights := make([]int64, 0, len(s.stats.Blocks))
	for height := range s.stats.Blocks {
		heights = append(heights, height)
	}
	sort.Slice(heights, func(i, j int) bool { return heights[i] < heights[j] })

	var (
		numTransactions []int64
		blockDurations  []float64
		blockSizes      []float64
		// mempoolSizes    []int64
		blockTimes []time.Time
		// rounds                                   []int32
		totalTxs                                 int64
		totalPayloadSize                         float64
		totalMempoolTxsSize                      int64
		totalUnconfirmedTxs                      int64
		blockRate                                float64
		transactionRate                          float64
		transactionPerBlock                      float64
		obsvdStorage, thrtlStorage, relBlockRate float64
	)

	blockTimes = append(blockTimes, s.stats.Blocks[heights[0]].Time)
	blockDurations = append(blockDurations, 0) // First block duration is zero

	for i, height := range heights {
		block := s.stats.Blocks[height]
		// rounds = append(rounds, block.Rounds)
		numTransactions = append(numTransactions, block.TxCount)
		totalTxs += block.TxCount
		blockSize := float64(block.BlockSize) / (1024 * 1024) // Convert to MB
		blockSizes = append(blockSizes, blockSize)
		totalPayloadSize += blockSize
		// mempoolSizes = append(mempoolSizes, block.UnconfirmedTxsBytes)
		totalMempoolTxsSize += block.UnconfirmedTxsBytes
		totalUnconfirmedTxs += int64(block.UnconfirmedTxs)

		if i > 0 {
			prevTime := blockTimes[i-1]
			duration := block.Time.Sub(prevTime).Seconds()
			blockDurations = append(blockDurations, duration)
		}
		blockTimes = append(blockTimes, block.Time)
	}

	// Calculate derived metrics
	numBlocks := float64(len(heights))
	transactionRate = float64(totalTxs) / testDuration
	transactionPerBlock = float64(totalTxs) / numBlocks
	blockRate = numBlocks / testDuration
	expectedBlockRate := 10.0  // Assuming block rate is every 10 seconds (replace with actual)
	blockSizeThreshold := 12.0 // Assuming 12 MB block size
	relBlockRate = blockRate / expectedBlockRate
	obsvdStorage = median(blockSizes) * blockRate * 24 * 60 / 1000
	thrtlStorage = expectedBlockRate * blockSizeThreshold * 24 * 60 / 1000

	metrics := derivedMetrics{
		NumBlocks:              int(numBlocks),
		TestDurationMinutes:    testDuration,
		TotalTransactions:      totalTxs,
		TotalPayloadSizeMB:     totalPayloadSize,
		ExpectedBlockRate:      expectedBlockRate,
		BlockRate:              blockRate,
		RelativeBlockRate:      relBlockRate,
		TransactionRate:        transactionRate,
		AvgTransactionPerBlock: transactionPerBlock,
		ObservedStorageGB:      obsvdStorage,
		TheoreticalStorageGB:   thrtlStorage,
		BlockUtilizationPct:    (median(blockSizes) * 100) / blockSizeThreshold,
	}

	metrics.BlockSizes.Min = min(blockSizes)
	metrics.BlockSizes.Max = max(blockSizes)
	metrics.BlockSizes.Mean = mean(blockSizes)
	metrics.BlockSizes.Median = median(blockSizes)

	metrics.BlockDurations.Min = min(blockDurations)
	metrics.BlockDurations.Max = max(blockDurations)
	metrics.BlockDurations.Mean = mean(blockDurations)
	metrics.BlockDurations.Median = median(blockDurations)

	metrics.TransactionsPerBlock.Min = slices.Min(numTransactions)
	metrics.TransactionsPerBlock.Max = slices.Max(numTransactions)
	metrics.TransactionsPerBlock.Median = medianInt(numTransactions)

	s.analysis = &metrics

	fmt.Printf("\n--- Analysis Metrics ---\n")
	fmt.Printf("Number of Blocks:               %v\n", int(numBlocks))
	fmt.Printf("Test Duration (minutes):        %.2f\n", testDuration)
	fmt.Printf("Total Transactions Processed:    %v\n", totalTxs)
	fmt.Printf("Total Payload Size (MB):         %.2f\n", totalPayloadSize)
	fmt.Println()
	fmt.Printf("Expected Block Rate:             %.2f\n", expectedBlockRate)
	fmt.Printf("Actual Block Rate (per min):     %.2f\n", blockRate)
	fmt.Printf("Relative Block Rate:             %.2f\n", relBlockRate)
	fmt.Println()
	fmt.Printf("Transaction Rate (per min):      %.2f\n", transactionRate)
	fmt.Printf("Avg Transaction Count per Block: %.2f\n", transactionPerBlock)
	fmt.Println()
	fmt.Printf("Block Sizes: min %.2f MB, max %.2f MB, mean %.2f MB, median %.2f MB\n",
		slices.Min(blockSizes), slices.Max(blockSizes), median(blockSizes), median(blockSizes))
	fmt.Printf("Block Durations: min %.2f s, max %.2f s, mean %.2f s, median %.2f s\n",
		slices.Min(blockDurations), slices.Max(blockDurations), mean(blockDurations), median(blockDurations))
	fmt.Printf("Transactions per block: min %d, max %d, median %.2f\n",
		slices.Min(numTransactions), slices.Max(numTransactions), medianInt(numTransactions))
	fmt.Println()
	fmt.Printf("Observed Storage (GB):           %.2f\n", obsvdStorage)
	fmt.Printf("Theoretical Storage (GB):        %.2f\n", thrtlStorage)
	fmt.Printf("Block Utilization %%:            %.2f%%\n", (median(blockSizes)*100)/blockSizeThreshold)
}

// DerivedMetrics holds the results of the analysis for JSON output
type derivedMetrics struct {
	NumBlocks              int     `json:"num_blocks"`
	TestDurationMinutes    float64 `json:"test_duration_minutes"`
	TotalTransactions      int64   `json:"total_transactions"`
	TotalPayloadSizeMB     float64 `json:"total_payload_size_mb"`
	ExpectedBlockRate      float64 `json:"expected_block_rate"`
	BlockRate              float64 `json:"block_rate"`
	RelativeBlockRate      float64 `json:"relative_block_rate"`
	TransactionRate        float64 `json:"transaction_rate"`
	AvgTransactionPerBlock float64 `json:"avg_transaction_per_block"`
	BlockSizes             struct {
		Min    float64 `json:"min_mb"`
		Max    float64 `json:"max_mb"`
		Mean   float64 `json:"mean_mb"`
		Median float64 `json:"median_mb"`
	} `json:"block_sizes"`
	BlockDurations struct {
		Min    float64 `json:"min_sec"`
		Max    float64 `json:"max_sec"`
		Mean   float64 `json:"mean_sec"`
		Median float64 `json:"median_sec"`
	} `json:"block_durations"`
	TransactionsPerBlock struct {
		Min    int64   `json:"min"`
		Max    int64   `json:"max"`
		Median float64 `json:"median"`
	} `json:"transactions_per_block"`
	ObservedStorageGB    float64 `json:"observed_storage_gb"`
	TheoreticalStorageGB float64 `json:"theoretical_storage_gb"`
	BlockUtilizationPct  float64 `json:"block_utilization_pct"`
}

func txSize(txs [][]byte) int64 {
	var size int64
	for _, tx := range txs {
		size += int64(len(tx))
	}
	return size
}

func mean(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func median(data []float64) float64 {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func medianInt(data []int64) float64 {
	sorted := make([]int64, len(data))
	copy(sorted, data)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return float64(sorted[mid-1]+sorted[mid]) / 2
	}
	return float64(sorted[mid])
}

func min(data []float64) float64 {
	minVal := data[0]
	for _, v := range data[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func max(data []float64) float64 {
	maxVal := data[0]
	for _, v := range data[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
