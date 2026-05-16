package main

import (
	"cache"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	concurrency       = 1000
	requestsPerWorker = 1000
	keyRange          = 100000
	valueSize         = 128
	cacheExpiration   = time.Minute
)

type benchResult struct {
	name    string
	total   int64
	success int64
	failed  int64
	cost    time.Duration
}

func main() {
	ctx := context.Background()
	endpoints := []string{"127.0.0.1:2379"}

	cli, err := cache.NewEtcdClusterClient(ctx, 60, endpoints, "cache")
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	printConfig()

	runAndPrint("preheat set", func() benchResult {
		return preheat(ctx, cli)
	})
	runAndPrint("hot get", func() benchResult {
		return hotGet(ctx, cli)
	})
	runAndPrint("mixed 90% get 10% set", func() benchResult {
		return mixedReadWrite(ctx, cli)
	})
	runAndPrint("miss with getter", func() benchResult {
		return missWithGetter(ctx, cli)
	})
}

func preheat(ctx context.Context, cli *cache.ClusterClient) benchResult {
	return runBench("preheat set", func(workerID, reqID int) error {
		id := workerID*requestsPerWorker + reqID
		key := cacheKey(id % keyRange)
		return cli.Set(ctx, key, makeValue(id))
	})
}

func hotGet(ctx context.Context, cli *cache.ClusterClient) benchResult {
	return runBench("hot get", func(workerID, reqID int) error {
		id := randID(workerID, reqID, keyRange)
		_, err := cli.Get(ctx, cacheKey(id), nil, cacheExpiration)
		return err
	})
}

func mixedReadWrite(ctx context.Context, cli *cache.ClusterClient) benchResult {
	return runBench("mixed 90% get 10% set", func(workerID, reqID int) error {
		id := randID(workerID, reqID, keyRange)
		if reqID%10 == 0 {
			return cli.Set(ctx, cacheKey(id), makeValue(id))
		}

		_, err := cli.Get(ctx, cacheKey(id), nil, cacheExpiration)
		return err
	})
}

func missWithGetter(ctx context.Context, cli *cache.ClusterClient) benchResult {
	base := keyRange * 10
	getter := func(args ...any) (any, error) {
		id := args[0].(int)
		return string(makeValue(id)), nil
	}

	return runBench("miss with getter", func(workerID, reqID int) error {
		id := base + randID(workerID, reqID, keyRange)
		_, err := cli.Get(ctx, cacheKey(id), getter, cacheExpiration, id)
		return err
	})
}

func runBench(name string, request func(workerID, reqID int) error) benchResult {
	var success int64
	var failed int64
	var printedErr int64
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				if err := request(workerID, j); err != nil {
					atomic.AddInt64(&failed, 1)
					if atomic.AddInt64(&printedErr, 1) <= 5 {
						fmt.Printf("[%s] request failed: %v\n", name, err)
					}
					continue
				}
				atomic.AddInt64(&success, 1)
			}
		}(i)
	}
	wg.Wait()

	return benchResult{
		name:    name,
		total:   success + failed,
		success: success,
		failed:  failed,
		cost:    time.Since(start),
	}
}

func randID(workerID, reqID, max int) int {
	r := rand.New(rand.NewSource(int64(workerID*requestsPerWorker + reqID)))
	return r.Intn(max)
}

func cacheKey(id int) string {
	return fmt.Sprintf("user:%d", id)
}

func makeValue(id int) []byte {
	prefix := fmt.Sprintf("value:%d:", id)
	if len(prefix) >= valueSize {
		return []byte(prefix[:valueSize])
	}
	return []byte(prefix + strings.Repeat("x", valueSize-len(prefix)))
}

func runAndPrint(name string, fn func() benchResult) {
	fmt.Printf("\n===== %s =====\n", name)
	printResult(fn())
}

func printConfig() {
	fmt.Printf("concurrency: %d\n", concurrency)
	fmt.Printf("requests per worker: %d\n", requestsPerWorker)
	fmt.Printf("key range: %d\n", keyRange)
	fmt.Printf("value size: %d bytes\n", valueSize)
	fmt.Printf("total requests per case: %d\n", concurrency*requestsPerWorker)
}

func printResult(result benchResult) {
	errorRate := 0.0
	qps := 0.0
	if result.total > 0 {
		errorRate = float64(result.failed) / float64(result.total) * 100
	}
	if result.cost > 0 {
		qps = float64(result.total) / result.cost.Seconds()
	}

	fmt.Printf("total: %d\n", result.total)
	fmt.Printf("cost: %v\n", result.cost)
	fmt.Printf("qps: %.2f\n", qps)
	fmt.Printf("success: %d\n", result.success)
	fmt.Printf("failed: %d\n", result.failed)
	fmt.Printf("error rate: %.2f%%\n", errorRate)
}
