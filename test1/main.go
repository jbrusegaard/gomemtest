package main

import (
	"app/pkg/test1"
	"flag"
)

func main() {
	// Create default config
	config := test1.NewDefaultConfig()

	// Parse command-line flags
	flag.IntVar(&config.ArraySize, "size", config.ArraySize, "Size of array to allocate (in elements)")
	flag.IntVar(&config.Iterations, "iter", config.Iterations, "Number of iterations for memory access test")
	flag.IntVar(&config.Threads, "threads", config.Threads, "Number of threads to use for multi-threaded test")
	flag.BoolVar(&config.Verbose, "verbose", config.Verbose, "Enable verbose output")
	flag.BoolVar(&config.SkipLargeTests, "skip-large", config.SkipLargeTests, "Skip large memory tests")
	flag.IntVar(&config.ChartWidth, "chart-width", config.ChartWidth, "Width of ASCII charts")
	flag.BoolVar(&config.TestSequential, "test-seq", config.TestSequential, "Run sequential vs random access test")
	flag.BoolVar(&config.TestThreaded, "test-threaded", config.TestThreaded, "Run multi-threaded test")
	flag.BoolVar(&config.TestDetailedSizes, "test-sizes", config.TestDetailedSizes, "Run detailed size tests")
	flag.Parse()

	// Create tester with the configured settings
	tester := test1.NewMemTester(config)

	// Run all tests
	tester.RunAll()
}
