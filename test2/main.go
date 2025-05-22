package main

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

const (
	// Size of the array in bytes (256 MB by default)
	arraySize = 256 * 1024 * 1024
	// Number of iterations for the test
	iterations = 1000000
)

// Node represents a node in a linked list for pointer chasing
type Node struct {
	next  *Node
	dummy [56]byte // padding to make Node size 64 bytes (cache line size)
}

// randomAccessTest measures latency for random memory access
func randomAccessTest() {
	// Create a large array
	data := make([]int64, arraySize/8)

	// Fill with some values
	for i := range data {
		data[i] = int64(i)
	}

	// Random access
	fmt.Println("\nRandom Access Test:")
	fmt.Printf("Array size: %d MB\n", arraySize/(1024*1024))

	// Warm up
	for i := 0; i < 1000; i++ {
		idx := rand.Intn(len(data))
		_ = data[idx]
	}

	// Measure
	start := time.Now()
	sum := int64(0)

	for i := 0; i < iterations; i++ {
		idx := rand.Intn(len(data))
		sum += data[idx]
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(iterations)

	fmt.Printf("Random access latency: %.2f ns (sum: %d)\n", nsPerAccess, sum)
}

// sequentialAccessTest measures latency for sequential memory access
func sequentialAccessTest() {
	// Create a large array
	data := make([]int64, arraySize/8)

	// Fill with some values
	for i := range data {
		data[i] = int64(i)
	}

	// Sequential access
	fmt.Println("\nSequential Access Test:")
	fmt.Printf("Array size: %d MB\n", arraySize/(1024*1024))

	// Warm up
	for i := 0; i < 1000; i++ {
		_ = data[i%len(data)]
	}

	// Measure
	start := time.Now()
	sum := int64(0)

	for i := 0; i < iterations; i++ {
		idx := i % len(data)
		sum += data[idx]
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(iterations)

	fmt.Printf("Sequential access latency: %.2f ns (sum: %d)\n", nsPerAccess, sum)
}

// pointerChasingTest provides a more accurate latency measurement
// by creating a linked list with randomized pointers, then traversing it
func pointerChasingTest() {
	fmt.Println("\nPointer Chasing Test (Most Accurate for Latency):")

	// Create array of nodes
	nodeCount := arraySize / 64 // 64 bytes per node
	nodes := make([]Node, nodeCount)

	// Create a random permutation for true random access pattern
	indices := rand.Perm(nodeCount)

	// Link nodes in a random order to force cache misses
	for i := 0; i < nodeCount-1; i++ {
		nodes[indices[i]].next = &nodes[indices[i+1]]
	}
	// Connect the last node back to a random node (not the first)
	randomIdx := rand.Intn(nodeCount-2) + 1
	nodes[indices[nodeCount-1]].next = &nodes[indices[randomIdx]]

	// Start at a random position
	current := &nodes[indices[0]]

	// Warm up
	for i := 0; i < 1000; i++ {
		current = current.next
	}

	// Measure pointer chasing latency
	start := time.Now()
	node := current
	count := 0

	for i := 0; i < iterations; i++ {
		node = node.next
		count++
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(iterations)

	fmt.Printf("Array size: %d MB, Nodes: %d\n", arraySize/(1024*1024), nodeCount)
	fmt.Printf("Pointer chasing latency: %.2f ns (count: %d)\n", nsPerAccess, count)
	fmt.Printf("Memory address of last node: %p\n", unsafe.Pointer(node))
}
