# RedButton-bot

Бот для решения задач от объединения Красная кнопка

## Запуск в Docker

Создайте рабочий файл окружения:

```shell
cp .env.example .env
```

Укажите в `.env` токен Telegram, Telegram ID администраторов, период работы бота и безопасный пароль PostgreSQL. Затем запустите приложение:

```shell
docker compose up --build -d
```

Просмотр логов:

```shell
docker compose logs -f bot
```

Остановка контейнеров:

```shell
docker compose down
```

Данные PostgreSQL сохраняются в volume `postgres-data`. Директория `tasks` монтируется в контейнер только для чтения. После изменения YAML-файлов перезапустите бот для повторной синхронизации:

```shell
docker compose restart bot
```
