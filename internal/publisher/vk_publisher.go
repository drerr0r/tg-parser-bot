package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/drerr0r/tgparserbot/internal/models"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"go.uber.org/zap"
)

// VKPublisher публикатор в VK
type VKPublisher struct {
	vk      *api.VK
	groupID int
	logger  *zap.SugaredLogger
}

// NewVKPublisher создает новый публикатор для VK
func NewVKPublisher(accessToken string, groupID int, logger *zap.SugaredLogger) (*VKPublisher, error) {
	vk := api.NewVK(accessToken)

	// Более простая проверка токена - запрос информации о пользователе
	_, err := vk.UsersGet(api.Params{
		"user_ids": 1, // Запрашиваем информацию о пользователе с ID 1
		"fields":   "first_name,last_name",
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки токена VK: %v", err)
	}

	// Проверяем что группа существует (если указан groupID)
	if groupID != 0 {
		groups, err := vk.GroupsGetByID(api.Params{
			"group_ids": groupID,
		})
		if err != nil {
			logger.Warnf("Не удалось проверить группу %d: %v", groupID, err)
		} else if len(groups) > 0 {
			logger.Infof("Группа найдена: %s", groups[0].Name)
		}
	}

	logger.Info("Успешное подключение к VK API")

	return &VKPublisher{
		vk:      vk,
		groupID: groupID,
		logger:  logger,
	}, nil
}

// Publish публикует пост в VK группу
func (p *VKPublisher) Publish(ctx context.Context, post *models.Post) error {
	p.logger.Infof("Публикация поста %d в VK группу %d", post.ID, p.groupID)

	// Подготавливаем контент
	content := p.prepareContent(post)

	// Создаем параметры для поста
	b := params.NewWallPostBuilder()

	b.OwnerID(-p.groupID) // Для групп используем отрицательный ID
	b.Message(content)
	b.FromGroup(true)

	// Если есть медиа, добавляем его
	if post.MediaURL != "" {
		// Загружаем медиа и получаем attachment
		attachment, err := p.uploadMedia(post)
		if err != nil {
			return fmt.Errorf("ошибка загрузки медиа: %v", err)
		}
		if attachment != "" {
			b.Attachments(attachment)
		}
	}

	// Публикуем пост
	_, err := p.vk.WallPost(b.Params)
	if err != nil {
		return fmt.Errorf("ошибка публикации в VK: %v", err)
	}

	p.logger.Infof("Пост %d успешно опубликован в VK", post.ID)
	return nil
}

// TestConnection проверяет подключение к VK
func (p *VKPublisher) TestConnection(ctx context.Context) error {
	_, err := p.vk.GroupsGetByID(api.Params{
		"group_ids": p.groupID,
	})
	if err != nil {
		return fmt.Errorf("ошибка проверки подключения к VK: %v", err)
	}
	return nil
}

// prepareContent подготавливает контент для публикации
func (p *VKPublisher) prepareContent(post *models.Post) string {
	var content strings.Builder

	content.WriteString(post.Content)

	// Добавляем информацию об источнике
	if post.SourceChannel != "" {
		content.WriteString(fmt.Sprintf("\n\n📎 Источник: %s", post.SourceChannel))
	}

	return content.String()
}

// uploadMedia загружает медиа файл и возвращает attachment
func (p *VKPublisher) uploadMedia(post *models.Post) (string, error) {
	p.logger.Infof("Загрузка медиа в VK: %s (тип: %s)", post.MediaURL, post.MediaType)

	switch post.MediaType {
	case models.MediaPhoto:
		return p.uploadPhoto(post)
	case models.MediaVideo:
		return p.uploadVideo(post)
	case models.MediaDocument:
		return p.uploadDocument(post)
	default:
		p.logger.Warnf("Неподдерживаемый тип медиа для VK: %s", post.MediaType)
		return "", nil
	}
}

// uploadPhoto загружает фото и возвращает attachment
func (p *VKPublisher) uploadPhoto(post *models.Post) (string, error) {
	// 1. Получаем URL для загрузки
	uploadServer, err := p.vk.PhotosGetWallUploadServer(api.Params{
		"group_id": p.groupID,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка получения upload server: %v", err)
	}

	// 2. Загружаем файл на сервер
	photoData, err := p.uploadFileToVK(uploadServer.UploadURL, post.MediaURL, "photo")
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки фото: %v", err)
	}

	// 3. Сохраняем фото в альбом группы
	savedPhoto, err := p.vk.PhotosSaveWallPhoto(api.Params{
		"group_id": p.groupID,
		"photo":    photoData.Photo,
		"server":   photoData.Server,
		"hash":     photoData.Hash,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения фото: %v", err)
	}

	if len(savedPhoto) == 0 {
		return "", fmt.Errorf("фото не было сохранено")
	}

	// 4. Формируем attachment
	photo := savedPhoto[0]
	return fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID), nil
}

// uploadVideo загружает видео и возвращает attachment
func (p *VKPublisher) uploadVideo(post *models.Post) (string, error) {
	// 1. Получаем URL для загрузки
	uploadServer, err := p.vk.VideoSave(api.Params{
		"group_id":    p.groupID,
		"name":        "Video from parser",
		"description": fmt.Sprintf("Source: %s", post.SourceChannel),
	})
	if err != nil {
		return "", fmt.Errorf("ошибка получения upload server для видео: %v", err)
	}

	// 2. Загружаем файл на сервер
	_, err = p.uploadFileToVK(uploadServer.UploadURL, post.MediaURL, "video_file")
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки видео: %v", err)
	}

	// 3. Видео автоматически сохраняется после загрузки
	// Формируем attachment из ответа
	return fmt.Sprintf("video%d_%d", uploadServer.OwnerID, uploadServer.VideoID), nil
}

// uploadDocument загружает документ и возвращает attachment
func (p *VKPublisher) uploadDocument(post *models.Post) (string, error) {
	// 1. Получаем URL для загрузки
	uploadServer, err := p.vk.DocsGetWallUploadServer(api.Params{
		"group_id": p.groupID,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка получения upload server для документа: %v", err)
	}

	// 2. Загружаем файл на сервер
	docData, err := p.uploadFileToVK(uploadServer.UploadURL, post.MediaURL, "file")
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки документа: %v", err)
	}

	// 3. Сохраняем документ
	savedDoc, err := p.vk.DocsSave(api.Params{
		"file":  docData.File,
		"title": "Document from parser",
	})
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения документа: %v", err)
	}

	// Проверяем что документ сохранен
	if savedDoc.Doc.ID == 0 {
		return "", fmt.Errorf("документ не был сохранен")
	}

	// 4. Формируем attachment
	return fmt.Sprintf("doc%d_%d", savedDoc.Doc.OwnerID, savedDoc.Doc.ID), nil
}

// uploadFileToVK загружает файл по URL на сервер VK
func (p *VKPublisher) uploadFileToVK(uploadURL, fileURL, fieldName string) (*VKUploadResponse, error) {
	// 1. Скачиваем файл по URL
	downloadResp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка скачивания файла: %v", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка HTTP при скачивании: %s", downloadResp.Status)
	}

	// 2. Читаем содержимое файла
	fileData, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %v", err)
	}

	// 3. Проверяем размер файла
	if len(fileData) == 0 {
		return nil, fmt.Errorf("файл пустой или не загружен")
	}
	if len(fileData) > 50*1024*1024 { // 50MB limit
		return nil, fmt.Errorf("файл слишком большой: %d bytes", len(fileData))
	}

	// 4. Создаем multipart форму
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Создаем поле для файла
	part, err := writer.CreateFormFile(fieldName, "file")
	if err != nil {
		return nil, fmt.Errorf("ошибка создания формы: %v", err)
	}

	// Записываем данные файла
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("ошибка записи файла в форму: %v", err)
	}

	// Закрываем writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия формы: %v", err)
	}

	// 5. Отправляем файл на сервер VK
	req, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	uploadResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки файла: %v", err)
	}
	defer uploadResp.Body.Close() // Закрываем Body ответа

	// 6. Читаем ответ от сервера VK
	responseData, err := io.ReadAll(uploadResp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// 7. Парсим ответ
	var uploadResponse VKUploadResponse
	if err := json.Unmarshal(responseData, &uploadResponse); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %v", err)
	}

	if uploadResponse.Error != "" {
		return nil, fmt.Errorf("ошибка от сервера VK: %s", uploadResponse.Error)
	}

	return &uploadResponse, nil
}

// VKUploadResponse структура для ответа от VK upload server
type VKUploadResponse struct {
	Server int    `json:"server"`
	Photo  string `json:"photo"`
	Hash   string `json:"hash"`
	File   string `json:"file"`
	Error  string `json:"error"`
}
