// Package test2 provides memory testing functionality focused on cache and latency measurement
package test2

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
	"unsafe"
)

// Config holds all configuration parameters for memory tests
type Config struct {
	SizeInMB      int
	Iterations    int
	RunBasicTests bool
	RunAdvanced   bool
	RunCacheTests bool
}

// NewDefaultConfig creates a Config with sensible defaults
func NewDefaultConfig() *Config {
	return &Config{
		SizeInMB:      256, // 256 MB default memory test size
		Iterations:    1000000,
		RunBasicTests: true,
		RunAdvanced:   true,
		RunCacheTests: true,
	}
}

// Node represents a node in a linked list for pointer chasing
type Node struct {
	Next  *Node
	Dummy [56]byte // padding to make Node size 64 bytes (cache line size)
}

// CacheSizes represents the typical cache sizes for different levels
type CacheSizes struct {
	L1 int
	L2 int
	L3 int
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

// RunAll executes all memory tests based on the configuration
func (m *MemTester) RunAll() {
	fmt.Println("Memory Latency and Cache Test Suite")
	m.PrintSystemInfo()

	if m.Config.RunBasicTests {
		m.RandomAccessTest()
		m.SequentialAccessTest()
		m.PointerChasingTest()
	}

	if m.Config.RunAdvanced {
		m.AdvancedLatencyTest(m.Config.SizeInMB)
	}

	if m.Config.RunCacheTests {
		cacheSizes := m.EstimateCacheSizes()
		m.RunCacheTests(cacheSizes)
	}
}

// RandomAccessTest measures latency for random memory access
func (m *MemTester) RandomAccessTest() {
	// Create a large array
	data := make([]int64, m.Config.SizeInMB*1024*1024/8)

	// Fill with some values
	for i := range data {
		data[i] = int64(i)
	}

	// Random access
	fmt.Println("\nRandom Access Test:")
	fmt.Printf("Array size: %d MB\n", m.Config.SizeInMB)

	// Warm up
	for i := 0; i < 1000; i++ {
		idx := rand.Intn(len(data))
		_ = data[idx]
	}

	// Measure
	start := time.Now()
	sum := int64(0)

	for i := 0; i < m.Config.Iterations; i++ {
		idx := rand.Intn(len(data))
		sum += data[idx]
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(m.Config.Iterations)

	fmt.Printf("Random access latency: %.2f ns (sum: %d)\n", nsPerAccess, sum)
}

// SequentialAccessTest measures latency for sequential memory access
func (m *MemTester) SequentialAccessTest() {
	// Create a large array
	data := make([]int64, m.Config.SizeInMB*1024*1024/8)

	// Fill with some values
	for i := range data {
		data[i] = int64(i)
	}

	// Sequential access
	fmt.Println("\nSequential Access Test:")
	fmt.Printf("Array size: %d MB\n", m.Config.SizeInMB)

	// Warm up
	for i := 0; i < 1000; i++ {
		_ = data[i%len(data)]
	}

	// Measure
	start := time.Now()
	sum := int64(0)

	for i := 0; i < m.Config.Iterations; i++ {
		idx := i % len(data)
		sum += data[idx]
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(m.Config.Iterations)

	fmt.Printf("Sequential access latency: %.2f ns (sum: %d)\n", nsPerAccess, sum)
}

// PointerChasingTest provides a more accurate latency measurement
// by creating a linked list with randomized pointers, then traversing it
func (m *MemTester) PointerChasingTest() {
	fmt.Println("\nPointer Chasing Test (Most Accurate for Latency):")

	// Create array of nodes
	nodeCount := m.Config.SizeInMB * 1024 * 1024 / 64 // 64 bytes per node
	nodes := make([]Node, nodeCount)

	// Create a random permutation for true random access pattern
	indices := rand.Perm(nodeCount)

	// Link nodes in a random order to force cache misses
	for i := 0; i < nodeCount-1; i++ {
		nodes[indices[i]].Next = &nodes[indices[i+1]]
	}
	// Connect the last node back to a random node (not the first)
	randomIdx := rand.Intn(nodeCount-2) + 1
	nodes[indices[nodeCount-1]].Next = &nodes[indices[randomIdx]]

	// Start at a random position
	current := &nodes[indices[0]]

	// Warm up
	for i := 0; i < 1000; i++ {
		current = current.Next
	}

	// Measure pointer chasing latency
	start := time.Now()
	node := current
	count := 0

	for i := 0; i < m.Config.Iterations; i++ {
		node = node.Next
		count++
	}

	elapsed := time.Since(start)
	nsPerAccess := float64(elapsed.Nanoseconds()) / float64(m.Config.Iterations)

	fmt.Printf("Array size: %d MB, Nodes: %d\n", m.Config.SizeInMB, nodeCount)
	fmt.Printf("Pointer chasing latency: %.2f ns (count: %d)\n", nsPerAccess, count)
	fmt.Printf("Memory address of last node: %p\n", unsafe.Pointer(node))
}
