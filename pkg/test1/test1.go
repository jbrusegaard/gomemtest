// Package memtest provides memory testing functionality similar to AIDA64
package test1

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Config holds all configuration parameters for memory tests
type Config struct {
	ArraySize         int
	Iterations        int
	Threads           int
	Verbose           bool
	SkipLargeTests    bool
	ChartWidth        int
	TestSequential    bool
	TestThreaded      bool
	TestDetailedSizes bool
}

// NewDefaultConfig creates a Config with sensible defaults
func NewDefaultConfig() *Config {
	// Default to 256MB for array size
	defaultArraySize := 256 * 1024 * 1024 / 8

	return &Config{
		ArraySize:         defaultArraySize,
		Iterations:        10000000,
		Threads:           runtime.NumCPU(),
		Verbose:           false,
		SkipLargeTests:    false,
		ChartWidth:        40,
		TestSequential:    true,
		TestThreaded:      true,
		TestDetailedSizes: true,
	}
}

// MemTester is the main struct for memory testing
type MemTester struct {
	Config *Config
}

// NewMemTester creates a new memory tester with the given configuration
func NewMemTester(config *Config) *MemTester {
	if config == nil {
		config = NewDefaultConfig()
	}
	return &MemTester{Config: config}
}

// PrintSystemInfo prints information about the system
func (m *MemTester) PrintSystemInfo() {
	fmt.Println("\n==== System Information ====")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()
}

// RunAll executes all configured memory tests
func (m *MemTester) RunAll() {
	fmt.Println("RAM Latency Test - Similar to AIDA64")
	m.PrintSystemInfo()

	memorySizeMB := m.Config.ArraySize * 8 / 1024 / 1024
	fmt.Printf("Allocating %d MB of RAM for testing...\n", memorySizeMB)

	// Allocate a large array
	array := make([]int64, m.Config.ArraySize)

	// Initialize the array with indices to create a linked list of pointers
	// This creates random access patterns to prevent CPU prefetching
	indices := rand.Perm(m.Config.ArraySize)
	for i := 0; i < m.Config.ArraySize-1; i++ {
		array[indices[i]] = int64(indices[i+1])
	}
	array[indices[m.Config.ArraySize-1]] = int64(indices[0]) // Close the loop

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
	for i := 0; i < m.Config.Iterations; i++ {
		j = array[j]
	}
	elapsed := time.Since(start)

	// Ensure j is used to prevent compiler optimization
	if j < 0 {
		fmt.Println(j)
	}

	// Calculate average latency per memory access
	avgLatency := float64(elapsed.Nanoseconds()) / float64(m.Config.Iterations)

	fmt.Printf("\nTest completed with %d iterations\n", m.Config.Iterations)
	fmt.Printf("Memory size: %d MB\n", memorySizeMB)
	fmt.Printf("Total time elapsed: %v\n", elapsed)
	fmt.Printf("Average memory latency: %.2f ns\n", avgLatency)
	m.drawChart("Random Access Latency", []float64{avgLatency}, []string{"256MB"}, "ns")

	// Run additional benchmark tests
	if m.Config.TestDetailedSizes {
		m.RunDetailedBenchmark()
	}

	if m.Config.TestSequential {
		m.MeasureSequentialAccess()
	}

	if m.Config.TestThreaded {
		m.RunThreadedTest()
	}
}

// RunDetailedBenchmark tests memory latency with different block sizes
func (m *MemTester) RunDetailedBenchmark() {
	fmt.Println("\n==== Detailed Memory Latency Benchmarks ====")

	// Test different memory block sizes to see effects of caching
	sizes := []int{4 * 1024, 64 * 1024, 1024 * 1024, 8 * 1024 * 1024}

	// Add large test if not skipped
	if !m.Config.SkipLargeTests {
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

	m.drawChart("Memory Latency by Block Size", results, labels, "ns")
}

// MeasureSequentialAccess compares sequential vs random memory access
func (m *MemTester) MeasureSequentialAccess() {
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
	m.drawChart("Access Pattern Comparison",
		[]float64{seqLatency, randLatency},
		[]string{"Sequential", "Random"}, "ns")

	// Calculate memory bandwidth
	bytesAccessed := iters * 8
	seqBandwidth := float64(bytesAccessed) / (float64(seqElapsed.Nanoseconds()) / 1e9) / 1e9
	randBandwidth := float64(bytesAccessed) / (float64(randElapsed.Nanoseconds()) / 1e9) / 1e9

	fmt.Printf("Sequential bandwidth: %.2f GB/s\n", seqBandwidth)
	fmt.Printf("Random bandwidth:    %.2f GB/s\n", randBandwidth)

	// Draw bandwidth chart
	m.drawChart("Memory Bandwidth",
		[]float64{seqBandwidth, randBandwidth},
		[]string{"Sequential", "Random"}, "GB/s")
}

// RunThreadedTest runs multi-threaded memory tests
func (m *MemTester) RunThreadedTest() {
	fmt.Println("\n==== Multi-threaded Memory Latency Test ====")
	fmt.Printf("Testing with 1-%d threads...\n", m.Config.Threads)

	// Array to store results for different thread counts
	results := make([]float64, m.Config.Threads)
	labels := make([]string, m.Config.Threads)

	// Allocate array once to avoid repeated allocation
	blockSize := 64 * 1024 * 1024 // 64MB per thread
	if blockSize*m.Config.Threads > 1024*1024*1024 {
		// Limit to 1GB total if too many threads
		blockSize = 1024 * 1024 * 1024 / m.Config.Threads
	}
	elements := blockSize / 8

	// Test with increasing number of threads
	for t := 1; t <= m.Config.Threads; t++ {
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

				iters := m.Config.Iterations / t
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

	m.drawChart("Multi-threaded Memory Latency", results, labels, "ns")
}

// ASCII chart rendering function
func (m *MemTester) drawChart(title string, values []float64, labels []string, unit string) {
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
		barLength := int((value / maxValue) * float64(m.Config.ChartWidth))
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

// FormatSize formats a file size in human-readable form
func FormatSize(bytes int) string {
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
