package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"mpm-client/internal/client"
)

func main() {
	// Определяем параметры командной строки
	serverAddr := flag.String("server", "localhost:50051", "Адрес gRPC сервера")
	flag.Parse()

	// Получаем команду от пользователя
	args := flag.Args()
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	// Создаем клиент
	albumClient, err := client.NewAlbumClient(*serverAddr)
	if err != nil {
		log.Fatalf("Ошибка при создании клиента: %v", err)
	}
	defer albumClient.Close()

	// Создаем контекст для запросов
	ctx := context.Background()

	// Обрабатываем команду
	switch strings.ToLower(args[0]) {
	case "list":
		// Получаем список альбомов
		albums, err := albumClient.GetAlbums(ctx)
		if err != nil {
			log.Fatalf("Ошибка при получении альбомов: %v", err)
		}

		fmt.Println("Список альбомов:")
		if len(albums) == 0 {
			fmt.Println("Альбомы отсутствуют")
		} else {
			for _, album := range albums {
				fmt.Printf("ID: %d\nНазвание: %s\nОписание: %s\nДата создания: %s\n\n",
					album.Id, album.Name, album.Description, album.CreatedAt)
			}
		}

	case "create":
		if len(args) < 3 {
			fmt.Println("Недостаточно аргументов. Использование: create <название> <описание>")
			os.Exit(1)
		}

		name := args[1]
		description := args[2]

		album, err := albumClient.CreateAlbum(ctx, name, description)
		if err != nil {
			log.Fatalf("Ошибка при создании альбома: %v", err)
		}

		fmt.Printf("Альбом успешно создан:\nID: %d\nНазвание: %s\nОписание: %s\nДата создания: %s\n",
			album.Id, album.Name, album.Description, album.CreatedAt)

	case "delete":
		if len(args) < 2 {
			fmt.Println("Недостаточно аргументов. Использование: delete <id>")
			os.Exit(1)
		}

		id, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			fmt.Printf("Некорректный ID альбома: %v\n", err)
			os.Exit(1)
		}

		success, err := albumClient.DeleteAlbum(ctx, int32(id))
		if err != nil {
			log.Fatalf("Ошибка при удалении альбома: %v", err)
		}

		if success {
			fmt.Printf("Альбом с ID %d успешно удален\n", id)
		} else {
			fmt.Printf("Не удалось удалить альбом с ID %d\n", id)
		}

	default:
		fmt.Printf("Неизвестная команда: %s\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Использование: mpm-client [опции] <команда> [аргументы]")
	fmt.Println("\nКоманды:")
	fmt.Println("  list                         Получить список всех альбомов")
	fmt.Println("  create <название> <описание> Создать новый альбом")
	fmt.Println("  delete <id>                  Удалить альбом по ID")
	fmt.Println("\nОпции:")
	fmt.Println("  -server string               Адрес gRPC сервера (по умолчанию \"localhost:50051\")")
}
