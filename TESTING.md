# Тестирование перед production

Тестовый контур разделён по скорости и назначению. Тесты PostgreSQL работают только с базой, имя которой заканчивается на `_test`; это защищает рабочие данные от очистки.

| Уровень | Что проверяет | Команда |
| --- | --- | --- |
| Unit | конфигурацию, часовые пояса, YAML, scoring, сервисы, представление и внутреннее состояние бота | `make test` |
| Race | конкурентный доступ и гонки данных | `make test-race` |
| Smoke | создание Telegram-клиента и запрос `getMe` через in-memory transport | `make test-smoke` |
| Integration | миграции, PostgreSQL-репозитории, запросы уведомлений и rollback транзакции | `make test-integration` |
| E2E | путь `YAML → sync → user → wrong flag → solve → scoring → profile → rating → notification` | `make test-e2e` |
| Image smoke | сборку production Docker image и запуск под непривилегированным пользователем | выполняется в CI и `make test-full` |

## Быстрый цикл разработки

```bash
make test
make test-race
make test-smoke
```

## PostgreSQL-тесты

Поднять изолированную БД:

```bash
make test-db-up
make test-integration
make test-e2e
make test-db-down
```

Для уже запущенной тестовой БД можно передать собственный DSN:

```bash
TEST_DATABASE_DSN='postgres://user:password@localhost:5432/project_test?sslmode=disable' make test-integration test-e2e
```

Таблицы тестовой БД очищаются перед каждым integration/e2e набором. Не используйте общую БД для параллельного запуска этих двух команд.

## Полный production gate

Требуются Go и Docker с Compose:

```bash
make test-full
```

Команда последовательно проверяет форматирование, `go vet`, unit-тесты с race detector и coverage, smoke-тест Telegram, PostgreSQL integration, E2E и сборку production image. Тестовый контейнер и volume удаляются автоматически.

Перед выпуском дополнительно проверяются реальные production-секреты и доступность Telegram/PostgreSQL в целевом окружении. Автотесты намеренно не отправляют сообщения настоящим пользователям.
