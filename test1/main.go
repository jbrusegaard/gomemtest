package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Configuration parameters
var (
	arraySize         int
	iterations        int
	threads           int
	verbose           bool
	skipLargeTests    bool
	chartWidth        int
	testSequential    bool
	testThreaded      bool
	testDetailedSizes bool
)

func init() {
	// Default to 256MB for array size
	defaultArraySize := 256 * 1024 * 1024 / 8

	// Parse command-line flags
	flag.IntVar(&arraySize, "size", defaultArraySize, "Size of array to allocate (in elements)")
	flag.IntVar(&iterations, "iter", 10000000, "Number of iterations for memory access test")
	flag.IntVar(&threads, "threads", runtime.NumCPU(), "Number of threads to use for multi-threaded test")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&skipLargeTests, "skip-large", false, "Skip large memory tests")
	flag.IntVar(&chartWidth, "chart-width", 40, "Width of ASCII charts")
	flag.BoolVar(&testSequential, "test-seq", true, "Run sequential vs random access test")
	flag.BoolVar(&testThreaded, "test-threaded", true, "Run multi-threaded test")
	flag.BoolVar(&testDetailedSizes, "test-sizes", true, "Run detailed size tests")
}

func main() {
	flag.Parse()
	fmt.Println("RAM Latency Test - Similar to AIDA64")
	printSystemInfo()

	memorySizeMB := arraySize * 8 / 1024 / 1024
	fmt.Printf("Allocating %d MB of RAM for testing...\n", memorySizeMB)

	// Allocate a large array
	array := make([]int64, arraySize)

	// Initialize the array with indices to create a linked list of pointers
	// This creates random access patterns to prevent CPU prefetching
	indices := rand.Perm(arraySize)
	for i := 0; i < arraySize-1; i++ {
		array[indices[i]] = int64(indices[i+1])
	}
	array[indices[arraySize-1]] = int64(indices[0]) // Close the loop

	fmt.Println("Warming up cache...")
	// Warm up
	j := int64(0)
	for i := 0; i < 1000000; i++ {
		j = array[j]
	}
	// Ensure j is used to prevent compiler optimization
	if j < 0 {
		fmt.Println(j)
	}

	fmt.Println("Running latency test...")

	// Measure random access time
	start := time.Now()
	j = 0
	for i := 0; i < iterations; i++ {
		j = array[j]
	}
	elapsed := time.Since(start)

	// Ensure j is used to prevent compiler optimization
	if j < 0 {
		fmt.Println(j)
	}

	// Calculate average latency per memory access
	avgLatency := float64(elapsed.Nanoseconds()) / float64(iterations)

	fmt.Printf("\nTest completed with %d iterations\n", iterations)
	fmt.Printf("Memory size: %d MB\n", memorySizeMB)
	fmt.Printf("Total time elapsed: %v\n", elapsed)
	fmt.Printf("Average memory latency: %.2f ns\n", avgLatency)
	drawChart("Random Access Latency", []float64{avgLatency}, []string{"256MB"}, "ns")

	// Run additional benchmark tests
	if testDetailedSizes {
		runDetailedBenchmark()
	}

	if testSequential {
		measureSequentialAccess()
	}

	if testThreaded {
		runThreadedTest()
	}
}

func printSystemInfo() {
	fmt.Println("\n==== System Information ====")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()
}

func runDetailedBenchmark() {
	fmt.Println("\n==== Detailed Memory Latency Benchmarks ====")

	// Test different memory block sizes to see effects of caching
	sizes := []int{4 * 1024, 64 * 1024, 1024 * 1024, 8 * 1024 * 1024}

	// Add large test if not skipped
	if !skipLargeTests {
		sizes = append(sizes, 64*1024*1024)
	}

	results := make([]float64, len(sizes))
	labels := make([]string, len(sizes))

	for i, size := range sizes {
		elements := size / 8 // For int64
		labels[i] = fmt.Sprintf("%d KB", size/1024)

		// Create array of appropriate size
		array := make([]int64, elements)

		// Setup random access pattern
		indices := rand.Perm(elements)
		for i := 0; i < elements-1; i++ {
			array[indices[i]] = int64(indices[i+1])
		}
		array[indices[elements-1]] = int64(indices[0])

		// Warm up
		j := int64(0)
		for i := 0; i < 1000000 && i < elements*10; i++ {
			j = array[j]
		}

		// Use volatile pointer to prevent optimization
		if j < 0 {
			fmt.Println(j)
		}

		// Number of iterations for measurement
		iters := 10000000
		if elements < 1000 {
			iters = 100000000 // More iterations for smaller arrays
		} else if elements > 1000000 {
			iters = 1000000 // Fewer iterations for larger arrays
		}

		start := time.Now()
		j = 0
		for i := 0; i < iters; i++ {
			j = array[j]
		}
		elapsed := time.Since(start)

		if j < 0 {
			fmt.Println(j)
		}

		avgLatency := float64(elapsed.Nanoseconds()) / float64(iters)
		results[i] = avgLatency
		fmt.Printf("Block size: %7d KB | Latency: %6.2f ns\n", size/1024, avgLatency)
	}

	drawChart("Memory Latency by Block Size", results, labels, "ns")
}

func measureSequentialAccess() {
	fmt.Println("\n==== Sequential vs Random Access ====")
	size := 64 * 1024 * 1024 // 64MB
	elements := size / 8

	array := make([]int64, elements)

	// Sequential pattern
	for i := 0; i < elements-1; i++ {
		array[i] = int64(i + 1)
	}
	array[elements-1] = 0

	// Random pattern
	randomArray := make([]int64, elements)
	indices := rand.Perm(elements)
	for i := 0; i < elements-1; i++ {
		randomArray[indices[i]] = int64(indices[i+1])
	}
	randomArray[indices[elements-1]] = int64(indices[0])

	// Measure sequential access
	j := int64(0)
	iters := 10000000

	start := time.Now()
	for i := 0; i < iters; i++ {
		j = array[j]
	}
	seqElapsed := time.Since(start)
	if j < 0 {
		fmt.Println(j)
	}

	// Measure random access
	j = 0
	start = time.Now()
	for i := 0; i < iters; i++ {
		j = randomArray[j]
	}
	randElapsed := time.Since(start)
	if j < 0 {
		fmt.Println(j)
	}

	seqLatency := float64(seqElapsed.Nanoseconds()) / float64(iters)
	randLatency := float64(randElapsed.Nanoseconds()) / float64(iters)

	fmt.Printf("Sequential access latency: %.2f ns\n", seqLatency)
	fmt.Printf("Random access latency:    %.2f ns\n", randLatency)

	// Draw chart for sequential vs random
	drawChart("Access Pattern Comparison",
		[]float64{seqLatency, randLatency},
		[]string{"Sequential", "Random"}, "ns")

	// Calculate memory bandwidth
	bytesAccessed := iters * 8
	seqBandwidth := float64(bytesAccessed) / (float64(seqElapsed.Nanoseconds()) / 1e9) / 1e9
	randBandwidth := float64(bytesAccessed) / (float64(randElapsed.Nanoseconds()) / 1e9) / 1e9

	fmt.Printf("Sequential bandwidth: %.2f GB/s\n", seqBandwidth)
	fmt.Printf("Random bandwidth:    %.2f GB/s\n", randBandwidth)

	// Draw bandwidth chart
	drawChart("Memory Bandwidth",
		[]float64{seqBandwidth, randBandwidth},
		[]string{"Sequential", "Random"}, "GB/s")
}

func runThreadedTest() {
	fmt.Println("\n==== Multi-threaded Memory Latency Test ====")
	fmt.Printf("Testing with 1-%d threads...\n", threads)

	// Array to store results for different thread counts
	results := make([]float64, threads)
	labels := make([]string, threads)

	// Allocate array once to avoid repeated allocation
	blockSize := 64 * 1024 * 1024 // 64MB per thread
	if blockSize*threads > 1024*1024*1024 {
		// Limit to 1GB total if too many threads
		blockSize = 1024 * 1024 * 1024 / threads
	}
	elements := blockSize / 8

	// Test with increasing number of threads
	for t := 1; t <= threads; t++ {
		labels[t-1] = fmt.Sprintf("%d", t)

		var wg sync.WaitGroup
		var mu sync.Mutex
		totalLatency := 0.0

		// Create arrays for each thread
		arrays := make([][]int64, t)
		for i := 0; i < t; i++ {
			arrays[i] = make([]int64, elements)
			indices := rand.Perm(elements)
			for j := 0; j < elements-1; j++ {
				arrays[i][indices[j]] = int64(indices[j+1])
			}
			arrays[i][indices[elements-1]] = int64(indices[0])
		}

		// Warm up all arrays
		for i := 0; i < t; i++ {
			j := int64(0)
			for k := 0; k < 100000 && k < elements; k++ {
				j = arrays[i][j]
			}
			if j < 0 {
				fmt.Println(j)
			}
		}

		start := time.Now()

		for i := 0; i < t; i++ {
			wg.Add(1)
			go func(threadID int) {
				defer wg.Done()

				iters := iterations / t
				if iters < 1000000 {
					iters = 1000000
				}

				j := int64(0)
				threadStart := time.Now()

				for k := 0; k < iters; k++ {
					j = arrays[threadID][j]
				}

				threadElapsed := time.Since(threadStart)
				latency := float64(threadElapsed.Nanoseconds()) / float64(iters)

				mu.Lock()
				totalLatency += latency
				mu.Unlock()

				if j < 0 {
					fmt.Println(j)
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)

		avgLatency := totalLatency / float64(t)
		results[t-1] = avgLatency

		fmt.Printf("%d thread(s): %.2f ns average latency (total elapsed: %v)\n",
			t, avgLatency, elapsed)
	}

	drawChart("Multi-threaded Memory Latency", results, labels, "ns")
}

// ASCII chart rendering function
func drawChart(title string, values []float64, labels []string, unit string) {
	fmt.Printf("\n==== %s ====\n", title)

	// Find the max value for scaling
	maxValue := 0.0
	for _, v := range values {
		if v > maxValue {
			maxValue = v
		}
	}

	// Find the max label length for alignment
	maxLabelLen := 0
	for _, label := range labels {
		if len(label) > maxLabelLen {
			maxLabelLen = len(label)
		}
	}

	// Draw each bar
	for i, value := range values {
		// Calculate bar length proportional to value
		barLength := int((value / maxValue) * float64(chartWidth))
		if barLength < 1 {
			barLength = 1
		}

		// Format label with padding for alignment
		paddedLabel := labels[i] + strings.Repeat(" ", maxLabelLen-len(labels[i]))

		// Draw the bar
		fmt.Printf("%s | %s %.2f %s\n",
			paddedLabel,
			strings.Repeat("█", barLength),
			value,
			unit)
	}
	fmt.Println()
}

// Helper function to format a file size in human-readable form
func formatSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
