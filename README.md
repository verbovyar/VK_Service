# VK_Service PubSub

## Описание
VK_Service — это лёгкий Go-сервис с gRPC API, реализующий паттерн «публикация-подписка».  
Он позволяет клиентам публиковать текстовые сообщения по «темам» (keys) и подписываться на поток этих сообщений.  
Все публикации сохраняются в PostgreSQL (pgxpool) для обеспечения исторической «ленты», а новые и старые сообщения стримятся подписчикам.

## Возможности
- **gRPC-сервер** с методами
    - `Publish(key, data)` — сохраняет запись в Postgres и рассылает её всем подписчикам.
    - `Subscribe(key)` — стримит сначала все исторические сообщения по теме, затем новые в режиме реального времени.
- **PostgreSQL** для надёжного хранения и выдачи истории сообщений.
- **pgxpool** — пул соединений для устойчивой работы под нагрузкой.
- **In-memory брокер** (`pkg/subpub`)
    - Буферизованные каналы (FIFO) для каждого подписчика
    - Неблокирующая рассылка (медленный подписчик не тормозит остальных)
    - Корректный graceful shutdown через WaitGroup и каналы
- **Конфигурируемость** через env-файл и флаги командной строки.
- **Unit-тесты** для ядра брокера.

## Требования
- Go 1.20+
- PostgreSQL 12+
- protoc + плагины:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest  
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
- Установите goose:
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest

## Структура проекта

Проект реализован по принципу чисто архитектуры.
Слой работы с базой данных изолирован от слоя сервиса, чтобы можно было легко поменять тип базы без переписывания логики хендлеров сервиса.

При остановке сервиса применяется graceful shutdown, чтобы корректно завершить все активные соединения и горутины

## Быстрый старт

1. Клонировать репозиторий 
```bash
git clone https://github.com/verbovyar/VK_Service.git
cd VK_Service
```
2. Настроить конфиг
   Отредактируйте файл config/conf.env:
```env
CONNECTION_STRING="postgres://user:password@localhost:5432/dbname?sslmode=disable"
NETWORK_TYPE=tcp
PORT=:8080
```
3. Установить зависимости и сгенерировать protobuf
```bash
go mod tidy
cd proto
protoc --go_out=. --go-grpc_out=. proto/pubsub.proto
```
Из корня репозитория
```bash
goose -dir migrations postgres "postgres://user:password@localhost:8080/dbname?sslmode=disable" up
```
4. Собрать сервер
```bash
go build -o bin/server cmd/main.go
```
5. Запустить
```bash
go run cmd/main.go
```
