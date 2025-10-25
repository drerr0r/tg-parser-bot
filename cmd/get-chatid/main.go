// cmd/get-chatid/main.go
package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("8471622405:AAGofvJO-tvYDFtY8TlWXW0qUkYMCNaJNjs")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🤖 Бот авторизован как:", bot.Self.UserName)
	fmt.Println("📝 Отправьте любое сообщение в ваш канал @drerr0r_test_parser_channel")
	fmt.Println("⏳ Получаем обновления...")

	// Получаем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.ChannelPost != nil {
			chat := update.ChannelPost.Chat
			fmt.Printf("\n🎯 НАЙДЕН КАНАЛ:\n")
			fmt.Printf("   ID: %d\n", chat.ID)
			fmt.Printf("   Username: @%s\n", chat.UserName)
			fmt.Printf("   Title: %s\n", chat.Title)
			fmt.Printf("   Type: %s\n", chat.Type)
			fmt.Printf("\n💡 Используйте этот ID в конфиге: %d\n", chat.ID)
			return
		}

		if update.Message != nil {
			chat := update.Message.Chat
			fmt.Printf("\n💬 НАЙДЕН ЧАТ:\n")
			fmt.Printf("   ID: %d\n", chat.ID)
			fmt.Printf("   Username: @%s\n", chat.UserName)
			fmt.Printf("   Title: %s\n", chat.Title)
			fmt.Printf("   Type: %s\n", chat.Type)
		}
	}
}
