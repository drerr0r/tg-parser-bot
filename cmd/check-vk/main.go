// cmd/check-vk/main.go
package main

import (
	"fmt"
	"log"

	"github.com/SevereCloud/vksdk/v2/api"
)

func main() {
	// Ваш токен
	token := "dddfef13dddfef13dddfef13bddee3e118ddddfdddfef13b52495badff4791dd026cb95"

	vk := api.NewVK(token)

	// Простая проверка токена
	users, err := vk.UsersGet(api.Params{
		"user_ids": 1,
	})

	if err != nil {
		log.Fatalf("❌ Ошибка проверки токена: %v", err)
	}

	fmt.Printf("✅ Токен валиден! Получен пользователь: %s %s\n",
		users[0].FirstName, users[0].LastName)

	// Проверка прав
	resp, err := vk.AccountGetAppPermissions(nil)
	if err != nil {
		log.Printf("⚠️ Не удалось проверить права: %v", err)
	} else {
		fmt.Printf("Права токена: %d\n", resp)
		fmt.Printf("Расшифровка прав:\n")
		if resp&1 != 0 {
			fmt.Println(" - notify")
		}
		if resp&2 != 0 {
			fmt.Println(" - friends")
		}
		if resp&4 != 0 {
			fmt.Println(" - photos")
		}
		if resp&8 != 0 {
			fmt.Println(" - audio")
		}
		if resp&16 != 0 {
			fmt.Println(" - video")
		}
		if resp&128 != 0 {
			fmt.Println(" - pages")
		}
		if resp&1024 != 0 {
			fmt.Println(" - status")
		}
		if resp&2048 != 0 {
			fmt.Println(" - notes")
		}
		if resp&4096 != 0 {
			fmt.Println(" - messages")
		}
		if resp&8192 != 0 {
			fmt.Println(" - wall")
		}
		if resp&32768 != 0 {
			fmt.Println(" - ads")
		}
		if resp&65536 != 0 {
			fmt.Println(" - offline")
		}
		if resp&131072 != 0 {
			fmt.Println(" - docs")
		}
		if resp&262144 != 0 {
			fmt.Println(" - groups")
		}
		if resp&1048576 != 0 {
			fmt.Println(" - notifications")
		}
		if resp&134217728 != 0 {
			fmt.Println(" - stories")
		}
		if resp&268435456 != 0 {
			fmt.Println(" - stats")
		}
	}
}
