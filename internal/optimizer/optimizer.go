// Package optimizer provides performance optimization utilities for AIO Stack.
package optimizer

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/aio-stack/aio-stack/internal/platform"
	"github.com/klauspost/cpuid/v2"
)

// Optimizer manages performance optimizations.
type Optimizer struct {
	config   *Config
	platform *platform.Info
	pools    map[string]*sync.Pool
	interned map[string]string
	mu       sync.RWMutex
	enabled  bool
	stats    Stats
}

// Config holds optimizer configuration.
type Config struct {
	Enabled         bool              `toml:"enabled" json:"enabled"`
	ProfileDir      string            `toml:"profile_dir" json:"profile_dir"`
	EnablePGO       bool              `toml:"enable_pgo" json:"enable_pgo"`
	EnableSIMD      bool              `toml:"enable_simd" json:"enable_simd"`
	EnablePooling   bool              `toml:"enable_pooling" json:"enable_pooling"`
	EnableInterning bool              `toml:"enable_interning" json:"enable_interning"`
	PoolSizes       map[string]int    `toml:"pool_sizes" json:"pool_sizes"`
	GCPercent       int               `toml:"gc_percent" json:"gc_percent"`
	MemoryBallast   int64             `toml:"memory_ballast" json:"memory_ballast"`
}

// Stats holds optimizer statistics.
type Stats struct {
	PoolHits     uint64
	PoolMisses   uint64
	InternHits   uint64
	InternMisses uint64
	AllocCount   uint64
	AllocBytes   uint64
}

// NewOptimizer creates a new optimizer.
func NewOptimizer(cfg *Config, pf *platform.Info) (*Optimizer, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	opt := &Optimizer{
		config:   cfg,
		platform: pf,
		pools:    make(map[string]*sync.Pool),
		interned: make(map[string]string),
		enabled:  cfg.Enabled,
	}

	// Set GC percent
	if cfg.GCPercent > 0 {
		debug.SetGCPercent(cfg.GCPercent)
	}

	// Set memory ballast
	if cfg.MemoryBallast > 0 {
		opt.setMemoryBallast(cfg.MemoryBallast)
	}

	// Initialize CPU features
	// cpuid.CPU is automatically populated on import

	return opt, nil
}

// DefaultConfig returns default optimizer configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		EnablePGO:       true,
		EnableSIMD:      true,
		EnablePooling:   true,
		EnableInterning: true,
		GCPercent:       100,
		MemoryBallast:   0,
		PoolSizes: map[string]int{
			"bytes":    1024,
			"strings": 512,
			"buffers":  256,
			"contexts": 128,
		},
	}
}

// Init initializes the optimizer.
func (o *Optimizer) Init() error {
	if !o.enabled {
		return nil
	}

	// Pre-create common pools
	o.GetPool("bytes", 4096)
	o.GetPool("strings", 256)
	o.GetPool("buffers", 8192)

	return nil
}

// Shutdown shuts down the optimizer.
func (o *Optimizer) Shutdown() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Clear pools
	o.pools = make(map[string]*sync.Pool)
	o.interned = make(map[string]string)

	return nil
}

// GetPool returns a sync.Pool for the given type and size.
func (o *Optimizer) GetPool(name string, size int) *sync.Pool {
	if !o.enabled || !o.config.EnablePooling {
		return nil
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if pool, ok := o.pools[name]; ok {
		return pool
	}

	pool := &sync.Pool{
		New: func() interface{} {
			atomic.AddUint64(&o.stats.PoolMisses, 1)
			switch name {
			case "bytes":
				return make([]byte, size)
			case "strings":
				return new(string)
			case "buffers":
				return make([]byte, size)
			default:
				return make([]byte, size)
			}
		},
	}

	o.pools[name] = pool
	return pool
}

// GetBytes returns a byte slice from the pool.
func (o *Optimizer) GetBytes(size int) []byte {
	if pool := o.GetPool("bytes", size); pool != nil {
		atomic.AddUint64(&o.stats.PoolHits, 1)
		buf := pool.Get().([]byte)
		if cap(buf) < size {
			return make([]byte, size)
		}
		return buf[:size]
	}
	return make([]byte, size)
}

// PutBytes returns a byte slice to the pool.
func (o *Optimizer) PutBytes(buf []byte) {
	if pool := o.GetPool("bytes", cap(buf)); pool != nil {
		pool.Put(buf)
	}
}

// GetString returns a string pointer from the pool.
func (o *Optimizer) GetString() *string {
	if pool := o.GetPool("strings", 0); pool != nil {
		atomic.AddUint64(&o.stats.PoolHits, 1)
		return pool.Get().(*string)
	}
	s := ""
	return &s
}

// PutString returns a string pointer to the pool.
func (o *Optimizer) PutString(s *string) {
	if pool := o.GetPool("strings", 0); pool != nil {
		*s = ""
		pool.Put(s)
	}
}

// Intern interns a string for memory efficiency.
func (o *Optimizer) Intern(s string) string {
	if !o.enabled || !o.config.EnableInterning {
		return s
	}

	o.mu.RLock()
	if interned, ok := o.interned[s]; ok {
		atomic.AddUint64(&o.stats.InternHits, 1)
		o.mu.RUnlock()
		return interned
	}
	o.mu.RUnlock()

	o.mu.Lock()
	defer o.mu.Unlock()

	// Double-check after acquiring write lock
	if interned, ok := o.interned[s]; ok {
		atomic.AddUint64(&o.stats.InternHits, 1)
		return interned
	}

	atomic.AddUint64(&o.stats.InternMisses, 1)
	o.interned[s] = s
	return s
}

// SIMDWidth returns the optimal SIMD width for the platform.
func (o *Optimizer) SIMDWidth() int {
	if !o.enabled || !o.config.EnableSIMD {
		return 0
	}

	if o.platform != nil {
		return o.platform.GetSIMDWidth()
	}

	// Fallback based on architecture
	switch runtime.GOARCH {
	case "amd64":
		if cpuid.CPU.Has(cpuid.AVX512F) {
			return 64 // 512 bits
		}
		if cpuid.CPU.Has(cpuid.AVX2) {
			return 32 // 256 bits
		}
		return 16 // SSE
	case "arm64":
		return 16 // NEON
	default:
		return 16
	}
}

// HasSIMD returns true if SIMD is available.
func (o *Optimizer) HasSIMD() bool {
	if !o.enabled || !o.config.EnableSIMD {
		return false
	}
	if o.platform != nil {
		return o.platform.GetSIMDWidth() > 16
	}
	return runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64"
}

// GetOptimalBatchSize returns the optimal batch size for the given operation.
func (o *Optimizer) GetOptimalBatchSize(operation string, elementSize int) int {
	// Base batch size on cache line and SIMD width
	simdWidth := o.SIMDWidth()
	if simdWidth == 0 {
		simdWidth = 16
	}

	// Aim for L1 cache friendly batches
	l1Size := 32768 // 32KB typical L1
	if o.platform != nil && o.platform.CPUFeatures.L1CacheSize > 0 {
		l1Size = o.platform.CPUFeatures.L1CacheSize
	}

	batchSize := l1Size / (elementSize * 2)
	if batchSize < simdWidth {
		batchSize = simdWidth
	}

	// Round to SIMD width
	batchSize = (batchSize / simdWidth) * simdWidth
	if batchSize == 0 {
		batchSize = simdWidth
	}

	// Cap at reasonable maximum
	if batchSize > 10000 {
		batchSize = 10000
	}

	return batchSize
}

// GetStats returns optimizer statistics.
func (o *Optimizer) GetStats() Stats {
	return Stats{
		PoolHits:     atomic.LoadUint64(&o.stats.PoolHits),
		PoolMisses:   atomic.LoadUint64(&o.stats.PoolMisses),
		InternHits:   atomic.LoadUint64(&o.stats.InternHits),
		InternMisses: atomic.LoadUint64(&o.stats.InternMisses),
	}
}

// ResetStats resets optimizer statistics.
func (o *Optimizer) ResetStats() {
	atomic.StoreUint64(&o.stats.PoolHits, 0)
	atomic.StoreUint64(&o.stats.PoolMisses, 0)
	atomic.StoreUint64(&o.stats.InternHits, 0)
	atomic.StoreUint64(&o.stats.InternMisses, 0)
}

// setMemoryBallast sets a memory ballast to reduce GC pressure.
func (o *Optimizer) setMemoryBallast(bytes int64) {
	if bytes <= 0 {
		return
	}
	// Allocate and hold memory to act as ballast
	ballast := make([]byte, bytes)
	_ = ballast // Prevent optimization
}

// String returns a string representation of the optimizer.
func (o *Optimizer) String() string {
	return fmt.Sprintf("Optimizer(enabled=%v, simd=%v, pooling=%v, interning=%v)",
		o.enabled, o.config.EnableSIMD, o.config.EnablePooling, o.config.EnableInterning)
}