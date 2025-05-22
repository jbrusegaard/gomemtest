package main

import (
	"fmt"
	"math/rand"
	"time"
)

// AdvancedLatencyTest runs a more sophisticated memory latency test
// that better simulates what AIDA64 does by avoiding prefetcher optimizations
func AdvancedLatencyTest(sizeInMB int) {
	// Convert MB to bytes
	sizeInBytes := sizeInMB * 1024 * 1024

	// Ensure we have at least 1000 nodes
	const nodeSize = 64 // typical cache line size in bytes
	nodeCount := sizeInBytes / nodeSize
	if nodeCount < 1000 {
		nodeCount = 1000
	}

	fmt.Printf("\nAdvanced Latency Test (%d MB):\n", sizeInMB)
	fmt.Printf("Creating %d nodes of %d bytes each...\n", nodeCount, nodeSize)

	// Create nodes array
	nodes := make([]Node, nodeCount)

	// Setup a random pointer chasing pattern with stride to defeat hardware prefetcher
	setupRandomPattern(&nodes)

	// Run multiple tests with increasing steps
	runLatencyTests(&nodes)
}

// setupRandomPattern creates a random access pattern through the nodes
// that defeats hardware prefetchers by ensuring non-sequential access
func setupRandomPattern(nodes *[]Node) {
	nodeCount := len(*nodes)

	// Create a pseudo-random pattern but with a large enough stride
	// to defeat the hardware prefetcher

	// Using the Knuth algorithm for creating a cyclic permutation
	// where each element is at least some minimum distance from its
	// natural position in the array

	// First create a valid permutation
	indices := make([]int, nodeCount)
	for i := 0; i < nodeCount; i++ {
		indices[i] = i
	}

	// Shuffle with restrictions to ensure hardware prefetcher can't predict it
	for i := 0; i < nodeCount; i++ {
		// Choose a random position at least 16 elements away (prefetchers often work within 16 cache lines)
		minDistance := 16
		if nodeCount > 64 {
			minDistance = nodeCount / 4 // For large arrays, make distance proportionally bigger
		}

		j := (i + minDistance + rand.Intn(nodeCount-minDistance)) % nodeCount
		indices[i], indices[j] = indices[j], indices[i]
	}

	// Link the nodes according to our shuffled pattern
	for i := 0; i < nodeCount-1; i++ {
		(*nodes)[indices[i]].next = &(*nodes)[indices[i+1]]
	}
	// Complete the cycle
	(*nodes)[indices[nodeCount-1]].next = &(*nodes)[indices[0]]
}

// runLatencyTests performs multiple iterations of the latency test
// and calculates average latency while avoiding compiler optimizations
func runLatencyTests(nodes *[]Node) {
	// Start with a random node
	firstNode := &(*nodes)[rand.Intn(len(*nodes))]

	// Number of iterations for each test
	const iterCount = 10000000 // 10 million iterations for better accuracy

	// Warmup to ensure CPU caches are in a realistic state
	warmupIterations := 1000000 // 1 million warmup iterations
	current := firstNode
	for i := 0; i < warmupIterations; i++ {
		current = current.next
	}

	// Actual test
	fmt.Println("Running latency test (this might take a few seconds)...")
	var minLatency, maxLatency, totalLatency float64
	minLatency = 1000000 // Start with a high number

	// Run 5 test passes and take the minimum (most accurate) result
	for pass := 0; pass < 5; pass++ {
		current = firstNode
		start := time.Now()

		for i := 0; i < iterCount; i++ {
			current = current.next // Pointer chasing forces memory read
		}

		elapsed := time.Since(start)
		nsPerAccess := float64(elapsed.Nanoseconds()) / float64(iterCount)

		// We use the current pointer to avoid compiler optimization
		if current == nil {
			fmt.Println("This should never happen - just preventing optimization")
		}

		// Track stats
		if nsPerAccess < minLatency {
			minLatency = nsPerAccess
		}
		if nsPerAccess > maxLatency {
			maxLatency = nsPerAccess
		}
		totalLatency += nsPerAccess

		fmt.Printf("Pass %d: %.2f ns\n", pass+1, nsPerAccess)
	}

	// Report results
	avgLatency := totalLatency / 5
	fmt.Printf("\nMemory Latency Results:\n")
	fmt.Printf("  Minimum: %.2f ns\n", minLatency)
	fmt.Printf("  Maximum: %.2f ns\n", maxLatency)
	fmt.Printf("  Average: %.2f ns\n", avgLatency)
}
