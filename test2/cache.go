package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
	"unsafe"
)

// CacheSizes represents the typical cache sizes for different levels
type CacheSizes struct {
	L1 int
	L2 int
	L3 int
}

// EstimateCacheSizes attempts to estimate cache sizes
// Note: This is an approximate method and not guaranteed to be accurate
func EstimateCacheSizes() CacheSizes {
	// Default values based on common CPU architectures
	// These will be overridden if our estimation is successful
	result := CacheSizes{
		L1: 32 * 1024,       // 32KB L1
		L2: 256 * 1024,      // 256KB L2
		L3: 8 * 1024 * 1024, // 8MB L3
	}

	// Try to estimate L1 cache size
	result.L1 = estimateCacheSize(4*1024, 64*1024, 4*1024)

	// Try to estimate L2 cache size
	result.L2 = estimateCacheSize(64*1024, 1024*1024, 64*1024)

	// Try to estimate L3 cache size
	result.L3 = estimateCacheSize(1024*1024, 32*1024*1024, 1024*1024)

	return result
}

// estimateCacheSize tries to detect a cache size by measuring access times
// for different array sizes and detecting a significant jump in latency
func estimateCacheSize(minSize, maxSize, step int) int {
	const samplesPerSize = 3
	const threshold = 1.5 // latency increase threshold to detect new cache level

	var lastLatency float64
	var bestSize int

	for size := minSize; size <= maxSize; size += step {
		// Create array of the current size
		data := make([]int64, size/8)

		// Initialize with some values
		for i := range data {
			data[i] = int64(i)
		}

		// Measure access time
		var totalLatency float64
		for s := 0; s < samplesPerSize; s++ {
			// Warm up
			for i := 0; i < 1000; i++ {
				j := i % len(data)
				_ = data[j]
			}

			start := time.Now()
			for i := 0; i < 100000; i++ {
				j := (i * 16) % len(data) // stride to defeat prefetcher
				_ = data[j]
			}
			totalLatency += float64(time.Since(start).Nanoseconds())
		}

		avgLatency := totalLatency / samplesPerSize

		// If we have a previous measurement
		if lastLatency > 0 {
			// Check for significant increase in latency
			if avgLatency > lastLatency*threshold {
				bestSize = size - step
				break
			}
		}

		lastLatency = avgLatency
	}

	return bestSize
}

// DetectCPUCacheInfo prints information about the CPU cache
func DetectCPUCacheInfo() {
	fmt.Println("\nCPU Cache Information")
	fmt.Println("====================")

	// Get estimated cache sizes
	caches := EstimateCacheSizes()

	fmt.Printf("Estimated L1 Cache: %d KB\n", caches.L1/1024)
	fmt.Printf("Estimated L2 Cache: %d KB\n", caches.L2/1024)
	fmt.Printf("Estimated L3 Cache: %d MB\n", caches.L3/(1024*1024))

	// Print cache line size (typical value, could be detected but requires architecture-specific code)
	fmt.Printf("Cache line size: 64 bytes (typical)\n")
	fmt.Printf("CPU cores: %d\n", runtime.NumCPU())
}

// MeasureCacheLatency measures and compares latency for different cache levels
func MeasureCacheLatency() {
	fmt.Println("\nCache Latency Comparison")
	fmt.Println("======================")

	// Sizes to test - targeting L1, L2, L3 and main memory
	sizes := []struct {
		name string
		size int
	}{
		{"L1 Cache", 16 * 1024},           // 16KB (should fit in L1)
		{"L2 Cache", 128 * 1024},          // 128KB (should fit in L2)
		{"L3 Cache", 4 * 1024 * 1024},     // 4MB (should fit in L3)
		{"Main Memory", 64 * 1024 * 1024}, // 64MB (beyond cache)
	}

	for _, size := range sizes {
		// Create nodes for pointer chasing
		nodeSize := int(unsafe.Sizeof(Node{}))
		nodeCount := size.size / nodeSize
		if nodeCount < 1000 {
			nodeCount = 1000 // Ensure we have enough nodes
		}

		nodes := make([]Node, nodeCount)

		// Create linked list with random order
		indices := rand.Perm(nodeCount)
		for i := 0; i < nodeCount-1; i++ {
			nodes[indices[i]].next = &nodes[indices[i+1]]
		}
		nodes[indices[nodeCount-1]].next = &nodes[indices[0]]

		// Warmup
		current := &nodes[indices[0]]
		for i := 0; i < 1000000; i++ {
			current = current.next
		}

		// Measure latency
		const testIterations = 10000000
		start := time.Now()
		for i := 0; i < testIterations; i++ {
			current = current.next
		}
		elapsed := time.Since(start)

		nsPerAccess := float64(elapsed.Nanoseconds()) / float64(testIterations)

		fmt.Printf("%s (%d KB): %.2f ns per access\n",
			size.name, size.size/1024, nsPerAccess)

		// Prevent optimization by using current
		if current == nil {
			fmt.Println("This should not happen")
		}
	}
}
