package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(f string) (*Config, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	//	fmt.Println("Config file:", string(data))

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func PrintConfig(c *Config) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Printf("Error printing config: %v\n", err)
		return
	}
	fmt.Println("--- Aegis Configuration ---")
	fmt.Println(string(data))
	fmt.Println("---------------------------")
}
func PrintRTConfig(c *RuntimeConfig) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Printf("Error printing config: %v\n", err)
		return
	}
	fmt.Println("--- Aegis Configuration ---")
	fmt.Println(string(data))
	fmt.Println("---------------------------")
}

func PrintSummary(c *Config) {
	fmt.Println("🛡️  Aegis Configuration Loaded")
	fmt.Println("======================================")

	fmt.Printf("🌐 SERVER:  %s:%d (Max Conns: %d)\n", c.Server.Host, c.Server.Port, c.Server.MaxConnections)
	fmt.Printf("📦 REDIS:   %s (Pool: %d)\n", c.Redis.Address, c.Redis.PoolSize)

	fmt.Println("\n⚙️  FEATURES:")
	fmt.Printf("   - HotKeys:     %v\n", c.Aegis.HotKeys)
	fmt.Printf("   - Singleflight: %v\n", c.Aegis.Singleflight)
	fmt.Printf("   - Tags:         %v\n", c.Aegis.Tags)

	if c.HotKeys != nil {
		fmt.Println("\n🔥 GLOBAL HOTKEY SETTINGS:")
		fmt.Printf("   - Max Tracked: %d\n", c.HotKeys.MaxTracked)
		fmt.Printf("   - Cleanup:     %s\n", c.HotKeys.CleanupInterval)
	}

	fmt.Println("\n📑 POLICIES:")
	for _, p := range c.Policies {
		fmt.Printf("   - [%s] Pattern: %-15s | TTL: %-5s | Singleflight: %v\n",
			p.Name, p.Match.Pattern, p.Config.TTL, p.Config.Singleflight)

		if p.Config.HotKeys != nil && p.Config.HotKeys.Enabled {
			fmt.Printf("     └─ HotKey: Threshold=%d, Multiplier=%.1fx\n",
				p.Config.HotKeys.Threshold, p.Config.HotKeys.TTLMultiplier)
		}
	}
	fmt.Println("======================================")
}
