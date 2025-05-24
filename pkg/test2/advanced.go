package test2

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

// AdvancedLatencyTest runs a more sophisticated memory latency test
// that better simulates what AIDA64 does by avoiding prefetcher optimizations
func (m *MemTester) AdvancedLatencyTest(sizeInMB int) {
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

	// Create a random permutation
	fmt.Println("Creating random memory access pattern...")
	indices := rand.Perm(nodeCount)

	// Link nodes in random order to form a circular list
	for i := 0; i < nodeCount-1; i++ {
		nodes[indices[i]].Next = &nodes[indices[i+1]]
	}
	nodes[indices[nodeCount-1]].Next = &nodes[indices[0]] // Close the loop

	// Flush cache and ensure nodes are in memory
	fmt.Println("Warming up cache...")
	runtime.GC()

	// Start from a node
	current := &nodes[indices[0]]

	// Warm up
	for i := 0; i < 1000; i++ {
		current = current.Next
	}

	// Measure latency
	fmt.Println("Measuring memory latency...")
	iterations := m.Config.Iterations
	start := time.Now()

	// Walk through the linked list, this will cause cache misses
	for i := 0; i < iterations; i++ {
		current = current.Next
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(iterations)

	// To prevent compiler from optimizing away the loop
	if current == nil {
		fmt.Println("This should never happen")
	}

	fmt.Printf("Advanced memory latency: %.2f ns\n", nsPerAccess)

	// Try to detect if the CPU has hardware prefetchers
	// A significant different between this test and the pointer chasing test
	// can indicate prefetcher activity
	fmt.Println("\nTesting for hardware prefetching effects...")
	m.testPrefetcher()
}

// testPrefetcher detects CPU prefetching by comparing sequential vs. random patterns
func (m *MemTester) testPrefetcher() {
	// Size of test array in int64 elements
	const size = 1024 * 1024 // 8 MB of int64 values

	// Create and initialize test array
	data := make([]int64, size)
	for i := range data {
		data[i] = int64(i)
	}

	// Test 1: Sequential access
	start := time.Now()
	sum := int64(0)
	for i := 0; i < m.Config.Iterations/4; i++ {
		idx := i % size
		sum += data[idx]
	}
	seqElapsed := time.Since(start)
	seqNsPerAccess := float64(seqElapsed.Nanoseconds()) / float64(m.Config.Iterations/4)

	// Test 2: Random access
	start = time.Now()
	sum = 0
	for i := 0; i < m.Config.Iterations/4; i++ {
		idx := rand.Intn(size)
		sum += data[idx]
	}
	randElapsed := time.Since(start)
	randNsPerAccess := float64(randElapsed.Nanoseconds()) / float64(m.Config.Iterations/4)

	// Test 3: Strided access (every 16th element)
	start = time.Now()
	sum = 0
	for i := 0; i < m.Config.Iterations/4; i++ {
		idx := (i * 16) % size
		sum += data[idx]
	}
	strideElapsed := time.Now().Sub(start)
	strideNsPerAccess := float64(strideElapsed.Nanoseconds()) / float64(m.Config.Iterations/4)

	fmt.Printf("Sequential access: %.2f ns\n", seqNsPerAccess)
	fmt.Printf("Random access:     %.2f ns\n", randNsPerAccess)
	fmt.Printf("Strided access:    %.2f ns\n", strideNsPerAccess)

	// Calculate ratios
	randomToSequentialRatio := randNsPerAccess / seqNsPerAccess
	strideToSequentialRatio := strideNsPerAccess / seqNsPerAccess

	fmt.Printf("\nRandom/Sequential ratio: %.2fx\n", randomToSequentialRatio)
	fmt.Printf("Stride/Sequential ratio: %.2fx\n", strideToSequentialRatio)

	// Interpret results
	if randomToSequentialRatio > 3.0 {
		fmt.Println("\nHardware prefetcher detected: Sequential access is significantly faster than random access.")
		fmt.Println("Your CPU likely has an active hardware prefetcher that improves sequential workloads.")
	} else {
		fmt.Println("\nHardware prefetcher may be disabled or less aggressive.")
	}

	if strideToSequentialRatio > 1.5 {
		fmt.Println("Stride prefetcher appears to be less effective with the chosen stride.")
	} else {
		fmt.Println("Stride prefetcher seems effective or the chosen stride matches prefetcher pattern.")
	}
}
