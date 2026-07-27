// Package platform provides cross-platform detection and optimization utilities.
package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/klauspost/cpuid/v2"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// Info holds platform information and capabilities.
type Info struct {
	// OS is the operating system (linux, windows, darwin, freebsd, etc.)
	OS string

	// Arch is the CPU architecture (amd64, arm64, arm, 386, etc.)
	Arch string

	// Version is the OS version.
	Version string

	// Kernel is the kernel version.
	Kernel string

	// Hostname is the system hostname.
	Hostname string

	// IsContainer indicates if running in a container.
	IsContainer bool

	// IsVM indicates if running in a VM.
	IsVM bool

	// IsTermux indicates if running in Termux.
	IsTermux bool

	// IsWSL indicates if running in WSL.
	IsWSL bool

	// CPUs is the number of logical CPUs.
	CPUs int

	// CPUModel is the CPU model name.
	CPUModel string

	// CPUFeatures are the detected CPU features.
	CPUFeatures CPUFeatures

	// MemoryTotal is total system memory in bytes.
	MemoryTotal uint64

	// MemoryAvailable is available memory in bytes.
	MemoryAvailable uint64

	// PageSize is the system page size.
	PageSize int

	// TempDir is the temporary directory.
	TempDir string

	// HomeDir is the user's home directory.
	HomeDir string

	// ConfigDir is the configuration directory.
	ConfigDir string

	// DataDir is the data directory.
	DataDir string

	// CacheDir is the cache directory.
	CacheDir string

	// ExecutablePath is the path to the current executable.
	ExecutablePath string

	// WorkingDir is the current working directory.
	WorkingDir string

	// Env is a copy of relevant environment variables.
	Env map[string]string

	// Config holds platform-specific configuration.
	Config *Config

	// DetectedAt is when the platform was detected.
	DetectedAt int64

	mu sync.RWMutex
}

// CPUFeatures holds detected CPU capabilities.
type CPUFeatures struct {
	// HasAVX2 indicates AVX2 support.
	HasAVX2 bool

	// HasAVX512F indicates AVX-512F support.
	HasAVX512F bool

	// HasAVX512VL indicates AVX-512VL support.
	HasAVX512VL bool

	// HasAVX512BW indicates AVX-512BW support.
	HasAVX512BW bool

	// HasAVX512DQ indicates AVX-512DQ support.
	HasAVX512DQ bool

	// HasAVX512VBMI indicates AVX-512VBMI support.
	HasAVX512VBMI bool

	// HasAVX512VBMI2 indicates AVX-512VBMI2 support.
	HasAVX512VBMI2 bool

	// HasAVX512BITALG indicates AVX-512BITALG support.
	HasAVX512BITALG bool

	// HasAVX512VPOPCNTDQ indicates AVX-512VPOPCNTDQ support.
	HasAVX512VPOPCNTDQ bool

	// HasNEON indicates ARM NEON support.
	HasNEON bool

	// HasSVE indicates ARM SVE support.
	HasSVE bool

	// HasSVE2 indicates ARM SVE2 support.
	HasSVE2 bool

	// HasFMA indicates FMA support.
	HasFMA bool

	// HasBMI1 indicates BMI1 support.
	HasBMI1 bool

	// HasBMI2 indicates BMI2 support.
	HasBMI2 bool

	// HasLZCNT indicates LZCNT support.
	HasLZCNT bool

	// HasPOPCNT indicates POPCNT support.
	HasPOPCNT bool

	// HasAES indicates AES-NI support.
	HasAES bool

	// HasSHA indicates SHA extensions support.
	HasSHA bool

	// HasRDRAND indicates RDRAND support.
	HasRDRAND bool

	// HasRDSEED indicates RDSEED support.
	HasRDSEED bool

	// CacheLineSize is the CPU cache line size in bytes.
	CacheLineSize int

	// L1CacheSize is the L1 cache size in bytes.
	L1CacheSize int

	// L2CacheSize is the L2 cache size in bytes.
	L2CacheSize int

	// L3CacheSize is the L3 cache size in bytes.
	L3CacheSize int
}

// Config holds platform configuration.
type Config struct {
	AutoDetect          bool
	PreferredArch       string
	PreferredOS         string
	EnableSIMD          bool
	EnableWASM          bool
	WASMRuntime         string
	MaxProcs            int
	MemoryLimit         int64
	CacheDir            string
	ConfigDir           string
	DataDir             string
	TermuxMode          bool
	WindowsConsoleMode  string
}

// Detect performs platform detection with the given configuration.
func Detect(cfg *Config) (*Info, error) {
	info := &Info{
		Config: cfg,
		Env:    make(map[string]string),
	}

	// Basic Go runtime info
	info.OS = runtime.GOOS
	info.Arch = runtime.GOARCH
	info.CPUs = runtime.NumCPU()
	info.PageSize = os.Getpagesize()
	info.ExecutablePath, _ = os.Executable()
	info.WorkingDir, _ = os.Getwd()
	info.TempDir = os.TempDir()
	info.HomeDir, _ = os.UserHomeDir()

	// Apply config overrides
	if cfg.PreferredOS != "" {
		info.OS = cfg.PreferredOS
	}
	if cfg.PreferredArch != "" {
		info.Arch = cfg.PreferredArch
	}

	// Detect OS-specific info
	switch info.OS {
	case "linux":
		if err := detectLinux(info); err != nil {
			return nil, fmt.Errorf("linux detection: %w", err)
		}
	case "windows":
		if err := detectWindows(info); err != nil {
			return nil, fmt.Errorf("windows detection: %w", err)
		}
	case "darwin":
		if err := detectDarwin(info); err != nil {
			return nil, fmt.Errorf("darwin detection: %w", err)
		}
	case "freebsd", "openbsd", "netbsd":
		if err := detectBSD(info); err != nil {
			return nil, fmt.Errorf("bsd detection: %w", err)
		}
	case "android":
		if err := detectAndroid(info); err != nil {
			return nil, fmt.Errorf("android detection: %w", err)
		}
	}

	// Detect Termux
	info.IsTermux = detectTermux(info)

	// Detect WSL
	info.IsWSL = detectWSL(info)

	// Detect container
	info.IsContainer = detectContainer(info)

	// Detect VM
	info.IsVM = detectVM(info)

	// CPU features
	if cfg.EnableSIMD {
		info.CPUFeatures = detectCPUFeatures(info)
	}

	// Memory info
	if mem, err := getMemoryInfo(); err == nil {
		info.MemoryTotal = mem.Total
		info.MemoryAvailable = mem.Available
	}

	// Directories
	info.ConfigDir = getConfigDir(info)
	info.DataDir = getDataDir(info)
	info.CacheDir = getCacheDir(info)

	// Environment variables
	captureEnv(info)

	return info, nil
}

// Init initializes platform-specific settings.
func (i *Info) Init() error {
	// Set GOMAXPROCS
	if i.Config.MaxProcs > 0 {
		runtime.GOMAXPROCS(i.Config.MaxProcs)
	}

	// Windows console mode
	if i.OS == "windows" {
		if err := setupWindowsConsole(i.Config.WindowsConsoleMode); err != nil {
			return fmt.Errorf("windows console setup: %w", err)
		}
	}

	// Termux optimizations
	if i.IsTermux {
		if err := setupTermux(); err != nil {
			return fmt.Errorf("termux setup: %w", err)
		}
	}

	return nil
}

// Shutdown cleans up platform resources.
func (i *Info) Shutdown() error {
	return nil
}

// GetOptimalThreadCount returns the optimal thread count for the given workload.
func (i *Info) GetOptimalThreadCount(workload string) int {
	base := i.CPUs
	switch workload {
	case "cpu-intensive":
		return base
	case "io-intensive":
		return base * 2
	case "mixed":
		return base + base/2
	default:
		return base
	}
}

// GetOptimalBufferSize returns an optimal buffer size for I/O operations.
func (i *Info) GetOptimalBufferSize() int {
	base := i.PageSize
	if i.CPUFeatures.CacheLineSize > 0 {
		base = i.CPUFeatures.CacheLineSize * 64 // 64 cache lines
	}
	if base < 4096 {
		base = 4096
	}
	if base > 1048576 {
		base = 1048576
	}
	return base
}

// GetOptimalPoolSize returns an optimal pool size for the given object size.
func (i *Info) GetOptimalPoolSize(objectSize int) int {
	// Aim for ~1MB per pool shard
	shardSize := 1024 * 1024
	count := shardSize / objectSize
	if count < 16 {
		count = 16
	}
	if count > 10000 {
		count = 10000
	}
	return count * i.CPUs
}

// IsWindows returns true if running on Windows.
func (i *Info) IsWindows() bool {
	return i.OS == "windows"
}

// IsLinux returns true if running on Linux.
func (i *Info) IsLinux() bool {
	return i.OS == "linux"
}

// IsDarwin returns true if running on macOS.
func (i *Info) IsDarwin() bool {
	return i.OS == "darwin"
}

// IsBSD returns true if running on a BSD variant.
func (i *Info) IsBSD() bool {
	switch i.OS {
	case "freebsd", "openbsd", "netbsd", "dragonflybsd":
		return true
	default:
		return false
	}
}

// IsUnix returns true if running on a Unix-like system.
func (i *Info) IsUnix() bool {
	return i.IsLinux() || i.IsDarwin() || i.IsBSD()
}

// HasFeature checks if a CPU feature is available.
func (i *Info) HasFeature(feature string) bool {
	switch strings.ToUpper(feature) {
	case "AVX2":
		return i.CPUFeatures.HasAVX2
	case "AVX512F":
		return i.CPUFeatures.HasAVX512F
	case "NEON":
		return i.CPUFeatures.HasNEON
	case "SVE":
		return i.CPUFeatures.HasSVE
	case "FMA":
		return i.CPUFeatures.HasFMA
	case "BMI2":
		return i.CPUFeatures.HasBMI2
	case "POPCNT":
		return i.CPUFeatures.HasPOPCNT
	case "AES":
		return i.CPUFeatures.HasAES
	case "SHA":
		return i.CPUFeatures.HasSHA
	default:
		return false
	}
}

// GetSIMDWidth returns the SIMD register width in bytes.
func (i *Info) GetSIMDWidth() int {
	if i.CPUFeatures.HasAVX512F {
		return 64 // 512 bits
	}
	if i.CPUFeatures.HasAVX2 {
		return 32 // 256 bits
	}
	if i.CPUFeatures.HasNEON {
		return 16 // 128 bits
	}
	return 16 // Default SSE/NEON
}

// String returns a human-readable platform description.
func (i *Info) String() string {
	return fmt.Sprintf("%s/%s (CPUs: %d, SIMD: %d bytes)", i.OS, i.Arch, i.CPUs, i.GetSIMDWidth())
}

// detectLinux detects Linux-specific information.
func detectLinux(info *Info) error {
	// Read /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				info.Version = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
	}

	// Read kernel version
	if data, err := os.ReadFile("/proc/version"); err == nil {
		info.Kernel = strings.TrimSpace(string(data))
	}

	// Hostname
	if h, err := os.Hostname(); err == nil {
		info.Hostname = h
	}

	// CPU model
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.CPUModel = strings.TrimSpace(parts[1])
				}
				break
			}
		}
	}

	return nil
}

// detectWindows detects Windows-specific information.
func detectWindows(info *Info) error {
	hi, err := host.Info()
	if err == nil {
		info.Version = hi.PlatformVersion
		info.Kernel = hi.KernelVersion
		info.Hostname = hi.Hostname
	}
	return nil
}

// detectDarwin detects macOS-specific information.
func detectDarwin(info *Info) error {
	hi, err := host.Info()
	if err == nil {
		info.Version = hi.PlatformVersion
		info.Kernel = hi.KernelVersion
		info.Hostname = hi.Hostname
	}

	// CPU model via sysctl
	if out, err := runCommand("sysctl", "-n", "machdep.cpu.brand_string"); err == nil {
		info.CPUModel = strings.TrimSpace(out)
	}

	return nil
}

// detectBSD detects BSD-specific information.
func detectBSD(info *Info) error {
	hi, err := host.Info()
	if err == nil {
		info.Version = hi.PlatformVersion
		info.Kernel = hi.KernelVersion
		info.Hostname = hi.Hostname
	}
	return nil
}

// detectAndroid detects Android-specific information.
func detectAndroid(info *Info) error {
	info.IsTermux = detectTermux(info)
	return nil
}

// detectTermux checks if running in Termux.
func detectTermux(info *Info) bool {
	if os.Getenv("TERMUX_VERSION") != "" {
		return true
	}
	if os.Getenv("PREFIX") != "" && strings.Contains(os.Getenv("PREFIX"), "com.termux") {
		return true
	}
	prefix := os.Getenv("PREFIX")
	if prefix != "" && strings.HasPrefix(prefix, "/data/data/com.termux") {
		return true
	}
	return false
}

// detectWSL checks if running in WSL.
func detectWSL(info *Info) bool {
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		return true
	}
	if _, err := os.Stat("/mnt/wsl"); err == nil {
		return true
	}
	if data, err := os.ReadFile("/proc/version"); err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "microsoft") || strings.Contains(content, "wsl") {
			return true
		}
	}
	return false
}

// detectContainer checks if running in a container.
func detectContainer(info *Info) bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return true
	}
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") ||
			strings.Contains(content, "containerd") ||
			strings.Contains(content, "kubepods") {
			return true
		}
	}
	return false
}

// detectVM checks if running in a VM.
func detectVM(info *Info) bool {
	vmPaths := []string{
		"/sys/class/dmi/id/product_name",
		"/sys/class/dmi/id/sys_vendor",
		"/sys/class/dmi/id/board_vendor",
	}
	for _, path := range vmPaths {
		if data, err := os.ReadFile(path); err == nil {
			content := strings.ToLower(string(data))
			vmKeywords := []string{"vmware", "virtualbox", "qemu", "kvm", "xen", "hyper-v", "parallels", "innotek"}
			for _, kw := range vmKeywords {
				if strings.Contains(content, kw) {
					return true
				}
			}
		}
	}
	return false
}

// detectCPUFeatures detects CPU capabilities using cpuid.
func detectCPUFeatures(info *Info) CPUFeatures {
	var features CPUFeatures

	cpu := cpuid.CPU

	// x86 features
	if info.Arch == "amd64" || info.Arch == "386" {
		features.HasAVX2 = cpu.Has(cpuid.AVX2)
		features.HasAVX512F = cpu.Has(cpuid.AVX512F)
		features.HasAVX512VL = cpu.Has(cpuid.AVX512VL)
		features.HasAVX512BW = cpu.Has(cpuid.AVX512BW)
		features.HasAVX512DQ = cpu.Has(cpuid.AVX512DQ)
		features.HasAVX512VBMI = cpu.Has(cpuid.AVX512VBMI)
		features.HasAVX512VBMI2 = cpu.Has(cpuid.AVX512VBMI2)
		features.HasAVX512BITALG = cpu.Has(cpuid.AVX512BITALG)
		features.HasAVX512VPOPCNTDQ = cpu.Has(cpuid.AVX512VPOPCNTDQ)
		features.HasFMA = cpu.Has(cpuid.FMA3)
		features.HasBMI1 = cpu.Has(cpuid.BMI1)
		features.HasBMI2 = cpu.Has(cpuid.BMI2)
		features.HasLZCNT = cpu.Has(cpuid.LZCNT)
		features.HasPOPCNT = cpu.Has(cpuid.POPCNT)
		features.HasAES = cpu.Has(cpuid.AESNI)
		features.HasSHA = cpu.Has(cpuid.SHA)
		features.HasRDRAND = cpu.Has(cpuid.RDRAND)
		features.HasRDSEED = cpu.Has(cpuid.RDSEED)

		// Cache info
		features.CacheLineSize = cpu.CacheLine
		if cpu.Cache.L1D > 0 {
			features.L1CacheSize = cpu.Cache.L1D
		}
		if cpu.Cache.L2 > 0 {
			features.L2CacheSize = cpu.Cache.L2
		}
		if cpu.Cache.L3 > 0 {
			features.L3CacheSize = cpu.Cache.L3
		}
	}

	// ARM features
	if info.Arch == "arm64" || info.Arch == "arm" {
		features.HasNEON = cpu.Has(cpuid.ASIMD)
		features.HasSVE = cpu.Has(cpuid.SVE)
		// SVE2 not available in this cpuid version
		features.CacheLineSize = cpu.CacheLine
		if cpu.Cache.L1D > 0 {
			features.L1CacheSize = cpu.Cache.L1D
		}
		if cpu.Cache.L2 > 0 {
			features.L2CacheSize = cpu.Cache.L2
		}
		if cpu.Cache.L3 > 0 {
			features.L3CacheSize = cpu.Cache.L3
		}
	}

	return features
}

// getMemoryInfo retrieves system memory information.
func getMemoryInfo() (*MemoryInfo, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &MemoryInfo{
		Total:     v.Total,
		Available: v.Available,
	}, nil
}

// MemoryInfo holds memory statistics.
type MemoryInfo struct {
	Total     uint64
	Available uint64
}

// getConfigDir returns the configuration directory.
func getConfigDir(info *Info) string {
	if info.Config.ConfigDir != "" {
		return info.Config.ConfigDir
	}
	switch info.OS {
	case "windows":
		return os.Getenv("APPDATA")
	case "darwin":
		return info.HomeDir + "/Library/Application Support"
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return xdg
		}
		return info.HomeDir + "/.config"
	}
}

// getDataDir returns the data directory.
func getDataDir(info *Info) string {
	if info.Config.DataDir != "" {
		return info.Config.DataDir
	}
	switch info.OS {
	case "windows":
		return os.Getenv("LOCALAPPDATA")
	case "darwin":
		return info.HomeDir + "/Library/Application Support"
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return xdg
		}
		return info.HomeDir + "/.local/share"
	}
}

// getCacheDir returns the cache directory.
func getCacheDir(info *Info) string {
	if info.Config.CacheDir != "" {
		return info.Config.CacheDir
	}
	switch info.OS {
	case "windows":
		return os.Getenv("LOCALAPPDATA") + "/Cache"
	case "darwin":
		return info.HomeDir + "/Library/Caches"
	default:
		if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
			return xdg
		}
		return info.HomeDir + "/.cache"
	}
}

// captureEnv captures relevant environment variables.
func captureEnv(info *Info) {
	keys := []string{
		"PATH", "HOME", "USER", "SHELL", "TERM", "LANG", "LC_ALL",
		"GOOS", "GOARCH", "GOPATH", "GOROOT", "GOBIN",
		"CGO_ENABLED", "CC", "CXX", "PKG_CONFIG_PATH",
		"TERMUX_VERSION", "PREFIX", "ANDROID_ROOT",
		"WSL_DISTRO_NAME", "WSLENV",
		"CONTAINER", "KUBERNETES_SERVICE_HOST",
	}
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			info.Env[key] = val
		}
	}
}

// runCommand runs a command and returns its output.
func runCommand(name string, args ...string) (string, error) {
	// Placeholder - in production use exec.CommandContext
	return "", fmt.Errorf("not implemented")
}

// setupWindowsConsole configures Windows console mode.
func setupWindowsConsole(mode string) error {
	return nil
}

// setupTermux applies Termux-specific optimizations.
func setupTermux() error {
	return nil
}