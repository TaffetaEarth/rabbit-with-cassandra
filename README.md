Поднять контейнеры rabbitmq и cassandra командой docker compose up -d

Запустить консьюмер
cd consumer 
go mod download
go run .

Запустить продюсер
cd producer 
go mod download
go run .

Пока запросы не будут распознаваться как DDOS (3-5 запросов), данные пользователей будут записаны в бд
