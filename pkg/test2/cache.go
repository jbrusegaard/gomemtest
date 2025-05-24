package test2

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
	"unsafe"
)

// EstimateCacheSizes attempts to estimate cache sizes
// Note: This is an approximate method and not guaranteed to be accurate
func (m *MemTester) EstimateCacheSizes() CacheSizes {
	// Default values based on common CPU architectures
	// These will be overridden if our estimation is successful
	result := CacheSizes{
		L1: 32 * 1024,       // 32KB L1
		L2: 256 * 1024,      // 256KB L2
		L3: 8 * 1024 * 1024, // 8MB L3
	}

	fmt.Println("\n==== Cache Size Estimation ====")
	fmt.Println("Running memory bandwidth test with different buffer sizes to detect cache levels...")

	// Test increasing buffer sizes from 4KB to 64MB
	sizes := []int{
		4 * 1024,         // 4KB
		8 * 1024,         // 8KB
		16 * 1024,        // 16KB
		32 * 1024,        // 32KB
		64 * 1024,        // 64KB
		128 * 1024,       // 128KB
		256 * 1024,       // 256KB
		512 * 1024,       // 512KB
		1024 * 1024,      // 1MB
		2 * 1024 * 1024,  // 2MB
		4 * 1024 * 1024,  // 4MB
		8 * 1024 * 1024,  // 8MB
		16 * 1024 * 1024, // 16MB
		32 * 1024 * 1024, // 32MB
	}

	// Array to store bandwidth results
	bandwidths := make([]float64, len(sizes))

	// Run the test for each buffer size
	for i, size := range sizes {
		// Create buffer
		elements := size / 8 // Each element is 8 bytes
		buffer := make([]int64, elements)

		// Initialize with random values
		for j := range buffer {
			buffer[j] = rand.Int63()
		}

		// Force all memory into RAM
		runtime.GC()

		// Iterations should be inversely proportional to size
		// to keep test duration reasonable
		iters := m.Config.Iterations / (size / 1024)
		if iters < 10 {
			iters = 10
		}

		// Warm up
		for j := 0; j < elements; j++ {
			_ = buffer[j]
		}

		// Measure sequential access bandwidth
		start := time.Now()
		sum := int64(0)

		for iter := 0; iter < iters; iter++ {
			for j := 0; j < elements; j++ {
				sum += buffer[j]
			}
		}

		elapsed := time.Since(start)
		bytesAccessed := int64(iters) * int64(elements) * 8
		bandwidthGBs := float64(bytesAccessed) / elapsed.Seconds() / 1e9
		bandwidths[i] = bandwidthGBs

		fmt.Printf("Buffer size: %7s, Bandwidth: %6.2f GB/s\n",
			formatSize(size), bandwidthGBs)

		// Prevent optimization
		if sum == 0 {
			fmt.Println("Should not happen")
		}
	}

	// Analyze results to detect cache boundaries
	// Look for significant drops in bandwidth
	l1Index, l2Index, l3Index := -1, -1, -1

	for i := 1; i < len(bandwidths); i++ {
		// If bandwidth drops more than 30%
		drop := 1.0 - (bandwidths[i] / bandwidths[i-1])

		if drop > 0.3 {
			if l1Index == -1 {
				l1Index = i
			} else if l2Index == -1 {
				l2Index = i
			} else if l3Index == -1 {
				l3Index = i
				break
			}
		}
	}

	// Update result with detected sizes
	if l1Index != -1 {
		result.L1 = sizes[l1Index-1]
	}
	if l2Index != -1 {
		result.L2 = sizes[l2Index-1]
	}
	if l3Index != -1 {
		result.L3 = sizes[l3Index-1]
	}

	fmt.Println("\n==== Cache Size Detection Results ====")
	fmt.Printf("L1 Cache (estimated): %s\n", formatSize(result.L1))
	fmt.Printf("L2 Cache (estimated): %s\n", formatSize(result.L2))
	fmt.Printf("L3 Cache (estimated): %s\n", formatSize(result.L3))
	fmt.Println("Note: These are estimates based on bandwidth patterns and may not be accurate.")

	return result
}

// RunCacheTests performs tests to measure cache latency and bandwidth
func (m *MemTester) RunCacheTests(cacheSizes CacheSizes) {
	fmt.Println("\n==== Cache Performance Tests ====")

	// Test L1, L2, L3 caches and main memory
	testSizes := []struct {
		name string
		size int
	}{
		{"L1 Cache", cacheSizes.L1 / 2},
		{"L2 Cache", cacheSizes.L2 / 2},
		{"L3 Cache", cacheSizes.L3 / 2},
		{"Main Memory", 64 * 1024 * 1024}, // 64MB, likely beyond all cache levels
	}

	// For each cache level, measure both latency and bandwidth
	for _, test := range testSizes {
		fmt.Printf("\nTesting %s (%s):\n", test.name, formatSize(test.size))

		// Measure latency with pointer chasing
		testCacheLatency(test.size, test.name)

		// Measure bandwidth with sequential access
		testCacheBandwidth(test.size, test.name)
	}
}

// testCacheLatency measures memory latency using pointer chasing
func testCacheLatency(size int, name string) {
	// Create a buffer that fits in the target cache
	nodeCount := size / 64 // using 64 byte nodes
	if nodeCount < 100 {
		nodeCount = 100 // ensure minimum size
	}

	// Create nodes
	nodes := make([]Node, nodeCount)

	// Create a random permutation
	indices := rand.Perm(nodeCount)

	// Link nodes in random order
	for i := 0; i < nodeCount-1; i++ {
		nodes[indices[i]].Next = &nodes[indices[i+1]]
	}
	nodes[indices[nodeCount-1]].Next = &nodes[indices[0]]

	// Force nodes into memory
	runtime.GC()

	// Warm up - walk through a small portion to load into cache
	current := &nodes[0]
	for i := 0; i < nodeCount; i++ {
		current = current.Next
	}

	// Measure latency
	iterations := 1000000
	start := time.Now()

	// Walk through the linked list
	for i := 0; i < iterations; i++ {
		current = current.Next
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(iterations)

	fmt.Printf("  %s latency: %.2f ns\n", name, nsPerAccess)

	// To prevent the compiler from optimizing
	if current == nil {
		fmt.Println("This should never happen")
	}
}

// testCacheBandwidth measures memory bandwidth using sequential access
func testCacheBandwidth(size int, name string) {
	// Create a buffer that fits in the target cache
	elements := size / 8 // Each element is 8 bytes
	buffer := make([]int64, elements)

	// Initialize with sequential values
	for i := range buffer {
		buffer[i] = int64(i)
	}

	// Force buffer into memory
	runtime.GC()

	// Warm up - read through the entire buffer
	sum := int64(0)
	for i := 0; i < elements; i++ {
		sum += buffer[i]
	}

	// Measure read bandwidth
	iterations := 1000
	if elements > 100000 {
		iterations = 100 // Fewer iterations for large buffers
	}

	start := time.Now()

	// Sequential reads
	for iter := 0; iter < iterations; iter++ {
		for i := 0; i < elements; i++ {
			sum += buffer[i]
		}
	}

	elapsed := time.Since(start)
	bytesRead := int64(elements) * 8 * int64(iterations)
	readBandwidthGBs := float64(bytesRead) / elapsed.Seconds() / 1e9

	// Measure write bandwidth
	start = time.Now()

	// Sequential writes
	for iter := 0; iter < iterations; iter++ {
		for i := 0; i < elements; i++ {
			buffer[i] = int64(i) + sum
		}
	}

	elapsed = time.Since(start)
	bytesWritten := int64(elements) * 8 * int64(iterations)
	writeBandwidthGBs := float64(bytesWritten) / elapsed.Seconds() / 1e9

	// Calculate combined read+write bandwidth
	copyStart := time.Now()

	// Copy operations (read + write)
	tempBuffer := make([]int64, elements)
	for iter := 0; iter < iterations; iter++ {
		copy(tempBuffer, buffer)
	}

	elapsed = time.Since(copyStart)
	bytesCopied := int64(elements) * 8 * int64(iterations)
	copyBandwidthGBs := float64(bytesCopied) / elapsed.Seconds() / 1e9

	fmt.Printf("  %s read bandwidth:      %.2f GB/s\n", name, readBandwidthGBs)
	fmt.Printf("  %s write bandwidth:     %.2f GB/s\n", name, writeBandwidthGBs)
	fmt.Printf("  %s copy bandwidth:      %.2f GB/s\n", name, copyBandwidthGBs)

	// To prevent the compiler from optimizing
	if sum == 0 {
		fmt.Printf("%p", unsafe.Pointer(&tempBuffer[0]))
	}
}

// formatSize formats a file size in human-readable form
func formatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/1024/1024)
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/1024/1024/1024)
}
