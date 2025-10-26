package main

import (
	"context"
	"fmt"
	"log"

	"github.com/drerr0r/tgparserbot/internal/config"
	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"github.com/drerr0r/tgparserbot/internal/utils"
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

	userRepo := storage.NewUserRepository(db)

	// Создаем правильный хэш для пароля "admin123"
	passwordHash, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatal("Ошибка хэширования пароля:", err)
	}

	user := &models.User{
		Username:     "admin",
		PasswordHash: passwordHash,
		Email:        "admin@example.com",
		Role:         "admin",
		IsActive:     true,
	}

	// Удаляем старого пользователя если есть
	ctx := context.Background()
	existingUser, _ := userRepo.GetByUsername(ctx, "admin")
	if existingUser != nil {
		fmt.Println("🗑️ Удаляем существующего пользователя...")
		// Здесь нужно добавить метод Delete в UserRepository или выполнить SQL
		_, err := db.Pool.Exec(ctx, "DELETE FROM users WHERE username = $1", "admin")
		if err != nil {
			log.Fatal("Ошибка удаления пользователя:", err)
		}
	}

	// Создаем нового пользователя
	fmt.Printf("🔐 Создаем пользователя:\n")
	fmt.Printf("   Username: %s\n", user.Username)
	fmt.Printf("   Password: admin123\n")
	fmt.Printf("   Hash: %s\n", user.PasswordHash)

	err = userRepo.Create(ctx, user)
	if err != nil {
		log.Fatal("Ошибка создания пользователя:", err)
	}

	fmt.Println("✅ Пользователь успешно создан!")
	fmt.Println("📝 Для тестирования используйте:")
	fmt.Println("   Username: admin")
	fmt.Println("   Password: admin123")
}
