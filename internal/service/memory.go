package service

import (
	"context"
	"runtime"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// MemoryOptimizedComplianceService provides memory-optimized operations for Lambda
type MemoryOptimizedComplianceService struct {
	*ComplianceService
	pool *ClientPool
}

// ClientPool manages reusable AWS clients to reduce memory allocations
type ClientPool struct {
	logsClients map[string]*cloudwatchlogs.Client
	kmsClients  map[string]*kms.Client
	mu          sync.RWMutex
}

// NewMemoryOptimizedComplianceService creates a memory-optimized service
func NewMemoryOptimizedComplianceService(baseService *ComplianceService) *MemoryOptimizedComplianceService {
	return &MemoryOptimizedComplianceService{
		ComplianceService: baseService,
		pool: &ClientPool{
			logsClients: make(map[string]*cloudwatchlogs.Client),
			kmsClients:  make(map[string]*kms.Client),
		},
	}
}

// GetLogsClient returns a cached CloudWatch Logs client for the region
func (cp *ClientPool) GetLogsClient(region string, createFunc func() *cloudwatchlogs.Client) *cloudwatchlogs.Client {
	cp.mu.RLock()
	if client, exists := cp.logsClients[region]; exists {
		cp.mu.RUnlock()
		return client
	}
	cp.mu.RUnlock()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := cp.logsClients[region]; exists {
		return client
	}

	// Create new client
	client := createFunc()
	cp.logsClients[region] = client
	return client
}

// GetKMSClient returns a cached KMS client for the region
func (cp *ClientPool) GetKMSClient(region string, createFunc func() *kms.Client) *kms.Client {
	cp.mu.RLock()
	if client, exists := cp.kmsClients[region]; exists {
		cp.mu.RUnlock()
		return client
	}
	cp.mu.RUnlock()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := cp.kmsClients[region]; exists {
		return client
	}

	// Create new client
	client := createFunc()
	cp.kmsClients[region] = client
	return client
}

// Cleanup releases resources and triggers garbage collection
func (cp *ClientPool) Cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Clear client maps to allow GC
	cp.logsClients = make(map[string]*cloudwatchlogs.Client)
	cp.kmsClients = make(map[string]*kms.Client)

	// Force garbage collection to free memory
	runtime.GC()
}

// MemoryStats provides memory usage statistics
type MemoryStats struct {
	AllocMB      uint64 // Current memory allocation in MB
	TotalAllocMB uint64 // Total memory allocated in MB
	SysMB        uint64 // System memory obtained from OS in MB
	NumGCRuns    uint32 // Number of GC runs
	HeapObjects  uint64 // Number of objects in heap
}

// GetMemoryStats returns current memory usage statistics
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		AllocMB:      bToMB(m.Alloc),
		TotalAllocMB: bToMB(m.TotalAlloc),
		SysMB:        bToMB(m.Sys),
		NumGCRuns:    m.NumGC,
		HeapObjects:  m.HeapObjects,
	}
}

// OptimizeMemory performs memory optimization operations
func OptimizeMemory() {
	// Force garbage collection
	runtime.GC()

	// Return memory to OS if possible
	runtime.GC()
	runtime.GC() // Call twice for better effect
}

// WithMemoryOptimization wraps a function call with memory optimization
func WithMemoryOptimization(ctx context.Context, fn func(context.Context) error) error {
	// Get baseline memory stats
	initialStats := GetMemoryStats()

	// Execute function
	err := fn(ctx)

	// Optimize memory after execution
	OptimizeMemory()

	// Get final memory stats
	finalStats := GetMemoryStats()

	// Log memory usage if significant
	if finalStats.AllocMB > initialStats.AllocMB+10 || initialStats.AllocMB > finalStats.AllocMB+10 {
		// Memory usage changed significantly - could add logging here if needed for debugging
		_ = finalStats // Mark as used to avoid unused variable warning
	}

	return err
}

// bToMB converts bytes to megabytes
func bToMB(b uint64) uint64 {
	return b / 1024 / 1024
}

// StringPool provides memory-efficient string operations
type StringPool struct {
	pool sync.Pool
}

// NewStringPool creates a new string pool
func NewStringPool() *StringPool {
	return &StringPool{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 1024) // 1KB initial capacity
				return &buf
			},
		},
	}
}

// GetBuffer gets a buffer from the pool
func (sp *StringPool) GetBuffer() []byte {
	bufPtr := sp.pool.Get().(*[]byte)
	buf := *bufPtr
	return buf[:0] // Reset length but keep capacity
}

// PutBuffer returns a buffer to the pool
func (sp *StringPool) PutBuffer(buf []byte) {
	if cap(buf) < 64*1024 { // Don't pool buffers larger than 64KB
		// Reset buffer length and return to pool
		buf = buf[:0]
		sp.pool.Put(&buf)
	}
}

// Global string pool for efficient string operations
var globalStringPool = NewStringPool()

// GetSharedBuffer gets a shared buffer for string operations
func GetSharedBuffer() []byte {
	return globalStringPool.GetBuffer()
}

// PutSharedBuffer returns a shared buffer
func PutSharedBuffer(buf []byte) {
	globalStringPool.PutBuffer(buf)
}
