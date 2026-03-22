
**Distributed Order Management System** 
Высокопроизводительная микросервисная система управления заказами на языке Go. 
**Архитектура системы**
Система состоит из 4-х микросервисов, объединенных в единую экосистему через gRPC и Kafka:

1.API Gateway (Gin): REST-интерфейс для внешних клиентов. Транслирует HTTP/JSON в gRPC и обеспечивает централизованную проверку JWT.

2.User Service (gRPC): Центр управления пользователями. Реализует Stateless Auth (JWT) и высокопроизводительное кэширование профилей в Redis для мгновенного доступа к данным.

3.Order Service (gRPC): Управляет заказами и транзакциями. Гарантирует целостность данных в PostgreSQL и инициирует бизнес-события через Kafka.

4.Notification Service: Асинхронный обработчик событий. Интегрирован с Telegram Bot API для мгновенного уведомления о новых заказах.

**Технологический стек**
Язык: Go 
Базы данных: PostgreSQL (pgxpool), Redis (Caching)
Транспорт: gRPC, HTTP (Gin), Kafka (Event Streaming)
Инструментарий: Docker, Docker Compose, Goose (Migrations), Protobuf (Buf)

**Ключевые особенности и оптимизации**
High-Performance Caching: Реализован паттерн Cache-Aside. Профили пользователей кэшируются в Redis, что снижает нагрузку на PostgreSQL и ускоряет типичные запросы в 10-20 раз.
Cache Stability:
Singleflight: Защита от Cache Stampede — база данных не перегружается при одновременном обращении тысяч пользователей к одному ресурсу.
Jitter: Размазывание времени жизни кэша (TTL) для предотвращения одновременного истечения срока действия записей.
Transactional Integrity (ACID): Атомарное создание заказа и его состава с использованием транзакций и метода Bulk Insert (CopyFrom).
Event-Driven Communication: Полная развязка сервисов заказов и уведомлений через Kafka.

**Запуск программы и настройка окружения**
1. Настройка окружения
Создайте .env из шаблона:
```bash
cp .env.example .env
```
Заполните переменные: JWT_KEY, ADMIN_PASSWORD, CHAT_ID.
2. Запуск системы

1.  **Клонируйте репозиторий:**
    ```bash
    git clone https://github.com/Chaice1/Order-Service
    ```

2.  **Перейдите в директорию проекта:**
    ```bash
    cd Order-Service
    ```
3.  **Сборка и запуск программы:**
    ```bash
    make build
    ```
4.  **Просто запуск программы:**
    ```bash
    make run
    ```
5.  **Остановка работы программы:**
    ```
    make stop
    ```

**API Endpoints**
Метод	Эндпоинт	Описание
POST /register	Регистрация пользователя
POST /login	Вход и получение токена
GET	 /auth/get_user	Профиль из кэша Redis/БД
POST /auth/create_order	Оформление заказа
GET	 /auth/get_order	Просмотр деталей заказа
