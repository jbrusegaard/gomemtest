# GoMemTest - RAM Benchmarking Suite

GoMemTest is a comprehensive memory benchmarking suite written in Go that provides detailed information about your system's RAM performance characteristics. It's designed to measure various aspects of memory performance similar to professional tools like AIDA64.

## Features

### Test1: RAM Latency Measurement Suite
- **Random Access Latency**: Measures true random access latency
- **Sequential vs. Random Access**: Compares sequential and random access patterns
- **Multi-threaded Testing**: Evaluates memory performance under multi-threaded loads
- **Detailed Size Tests**: Tests different memory block sizes to analyze cache effects

### Test2: Memory and Cache Analysis Suite
- **Basic Memory Tests**: Simple sequential and random access tests
- **Pointer Chasing**: Advanced tests that defeat CPU prefetching for accurate latency measurement
- **Cache Size Detection**: Automatically detects and measures L1, L2, and L3 cache sizes
- **Cache Performance**: Evaluates bandwidth and latency for each cache level
- **Prefetcher Detection**: Tests for CPU prefetcher effectiveness

## Installation

### Prerequisites
- Go 1.18 or higher

### Getting the Code
```bash
git clone https://github.com/yourusername/gomemtest.git
cd gomemtest
```

## Usage

### Using the Command Line Tools

#### Test1: RAM Latency Suite
```bash
make test1

# Or directly:
cd test1 && go run .

# With options:
go run test1/main.go -size=512 -iter=5000000 -threads=4
```

Options:
- `-size`: Size of array to allocate in elements (default: 33,554,432)
- `-iter`: Number of iterations for memory tests (default: 10,000,000)
- `-threads`: Number of threads to use for multi-threaded test (default: CPU count)
- `-verbose`: Enable verbose output
- `-skip-large`: Skip large memory tests
- `-chart-width`: Width of ASCII charts (default: 40)
- `-test-seq`: Run sequential vs random access test (default: true)
- `-test-threaded`: Run multi-threaded test (default: true)
- `-test-sizes`: Run detailed size tests (default: true)

#### Test2: Cache Analysis Suite
```bash
make test2

# Or directly:
cd test2 && go run .

# With options:
go run test2/main.go -size=512 -iter=2000000 -cache=true -basic=true -advanced=true
```

Options:
- `-size`: Size of memory to test in MB (default: 256)
- `-iter`: Number of iterations for tests (default: 1,000,000)
- `-basic`: Run basic memory tests (default: true)
- `-advanced`: Run advanced latency tests (default: true)
- `-cache`: Run cache detection and testing (default: true)
- `-help`: Show help message

### Using as a Library

GoMemTest can also be imported and used in your own Go programs:

```go
import "app/pkg/test1"

func main() {
    // Create a config with default settings
    config := test1.NewDefaultConfig()
    
    // Customize configuration
    config.ArraySize = 512 * 1024 * 1024 / 8
    config.Iterations = 5000000
    
    // Create tester and run tests
    tester := test1.NewMemTester(config)
    tester.RunAll()
}
```

```go
import "app/pkg/test2"

func main() {
    config := test2.NewDefaultConfig()
    config.SizeInMB = 512
    
    tester := test2.NewMemTester(config)
    tester.RunAll()
    
    // Or run specific tests
    tester.PointerChasingTest()
    cacheSizes := tester.EstimateCacheSizes()
    tester.RunCacheTests(cacheSizes)
}
```

## Understanding the Results

### Memory Latency
Memory latency is the time it takes for a CPU to request and receive data from RAM. Lower values are better, typically measured in nanoseconds (ns).

### Sequential vs. Random Access
- **Sequential Access**: Measures performance when accessing memory locations in order
- **Random Access**: Measures performance when accessing memory locations randomly
- The difference between these two is a good indicator of CPU prefetcher efficiency

### Cache Performance
Modern CPUs have multiple levels of cache:
- **L1 Cache**: Smallest but fastest cache (typically 32-64KB)
- **L2 Cache**: Mid-sized cache (typically 256KB-1MB)
- **L3 Cache**: Largest but slowest cache (typically 4-32MB)

The tests detect these cache sizes and measure their performance characteristics.

## System Requirements
- Operating System: Windows, macOS, or Linux
- Memory: At least 512MB free RAM
- For the full test suite, 1GB or more free RAM is recommended

## Interpreting Output

### ASCII Charts
Test results include ASCII bar charts for easy visualization of comparative results:
```
==== Memory Latency by Block Size ====
4 KB    | ████████████████████████████████████████ 1.56 ns
64 KB   | ██████████████████████████████████ 1.41 ns
1024 KB | ████████████████████ 0.94 ns
8192 KB | ██████████ 0.47 ns
```

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the LICENSE file for details.
