package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aio-stack/aio-stack/internal/ai"
	"github.com/aio-stack/aio-stack/internal/codeintel"
	"github.com/aio-stack/aio-stack/internal/optimizer"
	"github.com/aio-stack/aio-stack/internal/platform"
	"github.com/aio-stack/aio-stack/internal/plugin"
)

func main() {
	ctx := context.Background()

	// Test platform detection
	fmt.Println("=== Platform Detection ===")
	pf, err := platform.Detect(&platform.Config{
		AutoDetect: true,
		EnableSIMD: true,
	})
	if err != nil {
		fmt.Printf("Platform detection failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OS: %s, Arch: %s, CPUs: %d\n", pf.OS, pf.Arch, pf.CPUs)
	fmt.Printf("SIMD Width: %d bytes\n", pf.GetSIMDWidth())
	fmt.Printf("Has AVX2: %v, Has NEON: %v\n", pf.HasFeature("AVX2"), pf.HasFeature("NEON"))

	// Test optimizer
	fmt.Println("\n=== Optimizer ===")
	opt, err := optimizer.NewOptimizer(&optimizer.Config{
		Enabled:         true,
		EnableSIMD:      true,
		EnablePooling:   true,
		EnableInterning: true,
	}, pf)
	if err != nil {
		fmt.Printf("Optimizer creation failed: %v\n", err)
		os.Exit(1)
	}
	if err := opt.Init(); err != nil {
		fmt.Printf("Optimizer init failed: %v\n", err)
		os.Exit(1)
	}

	buf := opt.GetBytes(1024)
	fmt.Printf("Got buffer from pool: %d bytes\n", len(buf))
	opt.PutBytes(buf)

	s := opt.Intern("hello world")
	fmt.Printf("Interned string: %s\n", s)

	fmt.Printf("Optimizer stats: %+v\n", opt.GetStats())

	// Test code intelligence
	fmt.Println("\n=== Code Intelligence ===")
	ci, err := codeintel.NewEngine(&codeintel.Config{
		Enabled: true,
	}, pf, opt)
	if err != nil {
		fmt.Printf("CodeIntel creation failed: %v\n", err)
		os.Exit(1)
	}
	if err := ci.Start(ctx); err != nil {
		fmt.Printf("CodeIntel start failed: %v\n", err)
		os.Exit(1)
	}

	langs := ci.SupportedLanguages()
	fmt.Printf("Supported languages: %v\n", langs)

	// Test AI engine
	fmt.Println("\n=== AI Engine ===")
	aiEngine, err := ai.NewEngine(&ai.Config{
		Enabled:      true,
		DefaultModel: "phi-3-mini-4k-instruct-q4_k_m.gguf",
		EnableWASM:   true,
	}, pf, opt)
	if err != nil {
		fmt.Printf("AI engine creation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("AI engine created successfully (skipping model load)")
	_ = aiEngine

	// Test plugin manager
	fmt.Println("\n=== Plugin Manager ===")
	pm, err := plugin.NewManager(&plugin.Config{
		Enabled:         true,
		EnableWASM:      true,
		EnableGoPlugins: true,
		Registry:        "https://registry.aio-stack.dev",
	}, pf, opt)
	if err != nil {
		fmt.Printf("Plugin manager creation failed: %v\n", err)
		os.Exit(1)
	}
	if err := pm.LoadAll(ctx); err != nil {
		fmt.Printf("Plugin load failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Loaded plugins: %d\n", len(pm.List()))

	fmt.Println("\n=== All tests passed! ===")
}