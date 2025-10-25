// cmd/check-migrations/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/drerr0r/tgparserbot/internal/config"
	"github.com/drerr0r/tgparserbot/internal/storage"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	db, err := storage.New(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("🔍 Проверка миграций...")
	fmt.Printf("Применены ли миграции: %v\n", db.CheckMigrationsApplied())

	// Проверяем существование путей
	paths := []string{
		"./migrations",
		"migrations",
		"../migrations",
		"internal/storage/migrations",
	}

	fmt.Println("\n📁 Проверка путей к миграциям:")
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			files, _ := filepath.Glob(filepath.Join(path, "*.sql"))
			fmt.Printf("✅ %s: существует (%d файлов)\n", path, len(files))
		} else {
			fmt.Printf("❌ %s: не существует\n", path)
		}
	}
}
