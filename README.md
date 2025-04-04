# MangaCage Service Bot

MangaCage Service Bot — это телеграм-бот, предназначенный для модерации манги на сайте MangaCage. Бот предоставляет модераторам и администраторам удобный интерфейс для управления тайтлами, томами, главами и пользователями.

## Статус проекта

Проект находится в стадии разработки. Некоторые функции могут быть недоступны или работать некорректно.

## Функциональность

Бот поддерживает следующие команды:

### Для модерации:
- `/approve_title <id>` — Одобрить тайтл (или изменения в нём).
- `/reject_title <id>` — Отклонить обращение на модерацию тайтла.
- `/approve_volume <id>` — Одобрить том (или изменения в нём).
- `/reject_volume <id>` — Отклонить обращение на модерацию тома.
- `/approve_chapter <id>` — Одобрить главу (или изменения в ней).
- `/reject_chapter <id>` — Отклонить обращение на модерацию главы.
- `/approve_user <id>` — Верифицировать аккаунт пользователя.
- `/reject_user <id>` — Отклонить обращение на модерацию аккаунта.

### Для просмотра:
- `/get_new_titles_on_moderation` — Получить список новых тайтлов на модерации.
- `/get_edited_titles_on_moderation` — Получить список отредактированных тайтлов на модерации.
- `/get_new_volumes_on_moderation` — Получить список новых томов на модерации.
- `/get_edited_volumes_on_moderation` — Получить список отредактированных томов на модерации.
- `/get_new_chapters_on_moderation` — Получить список новых глав на модерации.
- `/get_edited_chapters_on_moderation` — Получить список отредактированных глав на модерации.
- `/get_new_users_on_moderation` — Получить список новых пользователей на модерации.
- `/get_edited_users_on_moderation` — Получить список отредактированных пользователей на модерации.

### Для просмотра одного объекта:
- `/review_title <id>` — Просмотреть подробности тайтла (с обложкой, по необходимости новой и старой, изменениями).
- `/review_volume <id>` — Просмотреть подробности тома (по необходимости с изменениями).
- `/review_chapter <id>` — Просмотреть подробности главы (по необходимости со страницами или изменениями).
- `/review_user <id>` — Просмотреть подробности пользователя (по необходимости с изменениями, аватаркой).

### Для работы с аккаунтом:
- `/login <username> <password>` — Войти в аккаунт модератора или администратора.

## Используемые технологии

- **Go** — основной язык программирования.
- **PostgreSQL** — для хранения данных о тайтлах, томах, главах и пользователях.
- **MongoDB** — для хранения больших бинарных данных (например, обложек и страниц).
- **Telegram Bot API** — для взаимодействия с пользователями через Telegram.
- **[go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)** — библиотека для работы с Telegram Bot API.

## Планы на будущее

- Настройка списка доверенных пользователей через Redis. Что позволит улучшить файловую структуру проекта за счёт более чёткого разделения логики.

## Обратная связь

Если у вас есть вопросы, предложения или вы нашли ошибку, пожалуйста, свяжитесь со мной:

- Через [GitHub Issues](https://github.com/Araks1255/mangacage_service_bot/issues)
- Через Telegram: [@Araks4621](https://t.me/Araks4621)