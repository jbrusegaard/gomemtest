package main

import (
	"app/pkg/test2"
	"flag"
	"fmt"
)

func main() {
	// Create default config
	config := test2.NewDefaultConfig()

	// Define command line flags
	flag.IntVar(&config.SizeInMB, "size", config.SizeInMB, "Size of memory to test in MB")
	flag.IntVar(&config.Iterations, "iter", config.Iterations, "Number of iterations for memory tests")
	runBasicPtr := flag.Bool("basic", true, "Run basic memory tests")
	runAdvancedPtr := flag.Bool("advanced", true, "Run advanced memory tests")
	runCachePtr := flag.Bool("cache", true, "Run cache detection and testing")
	showHelp := flag.Bool("help", false, "Show help")

	// Parse command line arguments
	flag.Parse()

	// Update config with flag values
	config.RunBasicTests = *runBasicPtr
	config.RunAdvanced = *runAdvancedPtr
	config.RunCacheTests = *runCachePtr

	// Show help if requested
	if *showHelp {
		printHelp()
		return
	}

	// Create tester with the configured settings
	tester := test2.NewMemTester(config)

	// Run all tests
	tester.RunAll()
}

// printHelp shows usage information
func printHelp() {
	fmt.Println("Memory Latency and Cache Test Suite")
	fmt.Println("\nUsage:")
	fmt.Println("  gomemtest [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -size=N      Size of memory to test in MB (default: 256)")
	fmt.Println("  -iter=N      Number of iterations for tests (default: 1,000,000)")
	fmt.Println("  -basic       Run basic memory tests (default: true)")
	fmt.Println("  -advanced    Run advanced latency tests (default: true)")
	fmt.Println("  -cache       Run cache detection and testing (default: true)")
	fmt.Println("  -help        Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  gomemtest -size=512")
	fmt.Println("  gomemtest -cache=false -basic=true -advanced=false")
}
