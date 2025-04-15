# go-basic

## Project MPM(Masterplan Photo Manager)
Краткое описание проекта: Приложение для работы с фотографиями и альбомами. Пользователь может создавать альбомы, загружать фотографии в альбомы, удалять фотографии и альбомы. Приложение поддерживает работу с локальным и облачным хранилищем. Пользователь может управлять фотографиями и альбомами через Telegram бота. Информация о фотографиях и альбомах будет хранится в базе данных.
Архитектура проекта:
1. Backend: Go
    - API для управления альбомами и фотографиями
    - Аутентификация и авторизация
2. Database: PostgresSQL
3. Storage: 
    - Local (Локальное файловое хранилище)
    - Облачное хранилище (Google Photos)
4. Telegram Bot
    - Управление фотографиями и альбомами
    - Оповещения о новых фотографиях
5. Развертывание в Docker

### HW - 09
1. Хранение данных в JSON-файлах для каждого слайса реализовано  в *json_storage.go*
2. Реализованы методы для каждого слайса: сохранение новых структур, получение структур.
3. Реализована загрузка данных при старте программы
4. Флаг *dirtyFlag* для отслеживания изменений
5. Отслеживание новых сущностей с помощью индексов


### HW - 08

1. Модифицирована функция monitorEntities
   - Добавлен параметр ctx context.Context для отслеживания сигнала завершения
   - Изменена структура цикла на select с обработкой ctx.Done() для корректного завершения
   - Добавлено логирование остановки мониторинга
2. Добавлен метод StartMonitoring
   - Публичный метод, запускающий мониторинг с поддержкой контекста
   - Заменяет старый метод startEntityMonitoring(), который больше не нужен
3. Обработка контекста в main.go
   - Создается корневой контекст с возможностью отмены
   - Настроена обработка сигналов ОС (SIGINT, SIGTERM)
   - Вызов cancel() при получении сигнала завершения
4. Модификация горутин
   - Горутина с тикером для генерации сущностей теперь проверяет ctx.Done()
   - Запуск мониторинга через entityService.StartMonitoring(ctx)
   - Все горутины корректно завершаются при отмене контекста
5. Удалены устаревшие функци
   - Удален метод startEntityMonitoring(), так как он не поддерживал контекст
   - Удален вызов s.startEntityMonitoring() из метода GenerateAndSaveEntities()

### HW - 07
1. Создана структура **EntiryJob** для передачи данных через канал
2. В **repository.go** добавлены mutex для защиты доступа к общим данным
3. Функция **GenerateAndSaveEntities** декомпозирована на две функции работающие в разных горутинах **generateEntities** для генерации сущностей и **saveEntities**
4. Добавлены функции для мониторинга слайсов **GetEntitiesCounts** и **GetNewEntities**
5. Добавлен старт мониторинга каждые 200 миллисекунд в **scheduler.go**

### HW - 06
1. Реализован CD для деплоя приложения в облако
2. Создан interface **provider.go** для работы с хранилищами и реализованный в **local_storage.go**
3. Добавлены слайсы **tags** и **meta** в структуры
4. Пакеты  **storage** и **handlers** вызываются из **main.go** 

### HW - 05
1. Описаны модели в пакете **models** 
2. Созданы методы для работы с пользователями в **storage**

### HW - 04
1. Создана ветка homework4
2. Создан модуль mpm
3. Описана архитектура проекта в README.md
4. Начата работа над структурой проекта


### HW - 03
1. Создана ветка homework3
2. Создан модуль homework03
3. Создана функция **chessBoard** для рисования шахматной доски
4. Создан pipeline **go.yml** для сборки проекта с помощью GitHub Actions

### HW - 02
1. Создана ветка dev
````bash
git checkout -b dev
````
2. Создана директория **homework02** c примерами кода.
````go
package main

import "fmt"

func main() {
	fmt.Println("Hello, GitHub")
}
````
3. Сделан коммит по в ветку **dev**
4. Создан PR в ветку **main** - https://github.com/Tyatyushkin/go-basic/pull/1
5. Обновлена локальная базовая ветка с помощью 
````
git pull
````

### HW - 01
1. Создан репозиторий и ветка main на github - https://github.com/Tyatyushkin/go-basic.git
2. Создана директория homework01 с примером кода.