# Asperitas

Backend на Go для фронтенда [Asperitas](https://github.com/d11z/asperitas). Хранилище — в памяти
(после перезапуска данные пропадают). Контракт API — `api/openapi.yaml`.

## Запуск

```sh
cp .env.example .env        # впишите свой секрет: openssl rand -hex 32
make run                    # http://localhost:8080
make test
```

## Структура

```
cmd/asperitas                       точка входа: конфиг, http.Server, graceful shutdown
internal/app                        сборка зависимостей и маршрутов (используется и в тестах)
internal/domain                     сущности и бизнес-правила (голоса, права, ошибки)
internal/usecase                    сценарии: AuthUseCase, PostUseCase
internal/repository/localstorage    in-memory репозитории (PostRepo, UserRepo)
internal/controller/http/v1         HTTP-хендлеры, валидация, формат ошибок
internal/controller/http/middleware JWT-авторизация, логирование
internal/controller/http/web        index.html для клиентских маршрутов, favicon, manifest
```

## API

| Метод | Путь | Доступ |
| --- | --- | --- |
| POST | `/api/register`, `/api/login` | публичный |
| GET | `/api/posts`, `/api/posts/{category}`, `/api/user/{username}`, `/api/post/{postId}` | публичный |
| POST | `/api/posts` | токен |
| DELETE | `/api/post/{postId}` | автор поста или admin |
| POST | `/api/post/{postId}` | токен (комментарий) |
| DELETE | `/api/post/{postId}/{commentId}` | автор комментария, автор поста или admin |
| GET | `/api/post/{postId}/upvote \| downvote \| unvote` | токен |
