package main

import (
	"flag"
	"fmt"
	"runtime"
)

func init() {
	// Set GOMAXPROCS to use all available cores
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	// Define command line flags
	testType := flag.String("test", "all", "Type of test to run: basic, advanced, cache, or all")
	sizeInMB := flag.Int("size", 256, "Size of memory to test in MB")
	cacheTest := flag.Bool("cache", false, "Detect and test CPU cache characteristics")
	help := flag.Bool("help", false, "Show help")

	// Parse command line arguments
	flag.Parse()

	// Show help if requested
	if *help {
		showHelp()
		return
	}

	// Print system info
	printSystemInfo()

	// Run the requested test
	switch *testType {
	case "basic":
		fmt.Println("\nRunning basic memory latency tests...")
		randomAccessTest()
		sequentialAccessTest()
		pointerChasingTest()

	case "advanced":
		fmt.Println("\nRunning advanced memory latency test...")
		AdvancedLatencyTest(*sizeInMB)

	case "cache":
		fmt.Println("\nRunning cache detection and latency tests...")
		DetectCPUCacheInfo()
		MeasureCacheLatency()

	case "all":
		fmt.Println("\nRunning all memory latency tests...")
		randomAccessTest()
		sequentialAccessTest()
		pointerChasingTest()
		AdvancedLatencyTest(*sizeInMB)
		if *cacheTest {
			DetectCPUCacheInfo()
			MeasureCacheLatency()
		}

	default:
		fmt.Printf("Unknown test type: %s\n", *testType)
		showHelp()
	}
}

// showHelp displays usage information for the program
func showHelp() {
	fmt.Println("Memory Latency Test Tool")
	fmt.Println("=======================")
	fmt.Println("Usage: memtest [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -test=<type>    Type of test to run: basic, advanced, cache, or all (default: all)")
	fmt.Println("  -size=<MB>      Size of memory to test in MB (default: 256)")
	fmt.Println("  -cache          Run cache detection and latency tests")
	fmt.Println("  -help           Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  memtest -test=basic")
	fmt.Println("  memtest -test=advanced -size=512")
	fmt.Println("  memtest -test=cache")
	fmt.Println("  memtest -size=1024 -cache")
}

// printSystemInfo prints information about the system
func printSystemInfo() {
	fmt.Println("Memory Latency Test")
	fmt.Println("===================")

	// Get CPU information
	fmt.Printf("CPU: %d cores\n", runtime.NumCPU())

	// Get memory information
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("Total Memory: %.1f GB\n", float64(memStats.TotalAlloc)/(1024*1024*1024))

	// Get OS information
	fmt.Printf("OS: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
