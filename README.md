# Content Tracker

Rastreador de precios de juegos de Steam. Buscás un juego, lo agregás a tu lista personal, fijás un **precio objetivo** y la app te avisa cuando el juego llega a ese precio.

## Qué hace

- Registro e inicio de sesión con JWT.
- Búsqueda de juegos usando la tienda de Steam.
- Lista personal con precio objetivo editable por juego.
- Un **worker en segundo plano** que revisa los precios en Steam cada cierto tiempo y genera avisos cuando un precio baja hasta el objetivo.
- Panel de avisos con contador de no leídos.

## Stack

| Capa | Tecnología |
|---|---|
| Backend | Go, [chi](https://github.com/go-chi/chi) |
| Base de datos | PostgreSQL 16 (pgx / pgxpool) |
| Auth | JWT |
| Front | HTML + JavaScript sin framework (`web/index.html`) |
| Infra local | Docker Compose |

## Cómo funciona

```
Navegador ──► API (chi + JWT) ──► PostgreSQL
                  ▲
Worker (goroutine) ─┤──► Steam Store API
```

Cada fila de `games` es la relación *usuario ↔ juego* con su precio objetivo. El worker:

1. Toma los `steam_app_id` **únicos** de la tabla, para consultar a Steam una sola vez por juego aunque lo sigan muchos usuarios.
2. Compara el precio actual contra el guardado. Si **no cambió**, no hace nada, así no se repiten avisos en cada ciclo.
3. Si cambió y el nuevo precio es menor o igual al objetivo de un usuario, crea una notificación para ese usuario.
4. Guarda el precio nuevo.

## Correrlo en local

Requisitos: Go, Docker.

```bash
# 1. Variables de entorno
cp .env.example .env        # y completá los valores

# 2. Base de datos
docker compose up -d db
for f in db/migrations/*.sql; do
  docker compose exec -T db psql -U postgres -d content_tracker < "$f"
done

# 3. API
go run ./cmd/api

# 4. Front (en otra terminal, sirviendo solo la carpeta web, solo en tu máquina)
cd web && python3 -m http.server 5500 --bind 127.0.0.1
```

Abrí `http://localhost:5500`.

### Variables de entorno

| Variable | Descripción |
|---|---|
| `PORT` | Puerto de la API (ej. `8080`) |
| `DATABASE_URL` | Cadena de conexión a Postgres |
| `JWT_SECRET_KEY` | Secreto para firmar tokens. Generalo con `openssl rand -base64 32` |
| `JWT_EXPIRATION_HOURS` | Duración del token en horas |
| `ENV` | `dev` habilita CORS para `localhost:5500` y el endpoint de debug |

## API

Los endpoints marcados con 🔒 requieren `Authorization: Bearer <token>`.

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/users` | Crear cuenta |
| `POST` | `/users/login` | Iniciar sesión, devuelve el token |
| `GET` | `/games/search?term=` 🔒 | Buscar juegos en Steam |
| `POST` | `/games` 🔒 | Agregar un juego. Body: `{"steam_app_id": 367520, "target_price": 2.5}` |
| `GET` | `/games/user-games` 🔒 | Listar mis juegos |
| `PATCH` | `/games/update` 🔒 | Cambiar el objetivo. Body: `{"steam_app_id": 367520, "target_price": 3}` |
| `DELETE` | `/games/{steamAppID}` 🔒 | Quitar un juego de mi lista |
| `GET` | `/notifications` 🔒 | Mis avisos, del más nuevo al más viejo |
| `PATCH` | `/notificationS/read` 🔒 | Marcar mis avisos como leídos |

Con `ENV=dev` existe además `POST /debug/check-prices`, que dispara un chequeo de precios manual.

## Decisiones de diseño

- **Un solo pool de conexiones** (`pgxpool`) compartido por handlers y worker, en lugar de una conexión única.
- **Precios convertidos en un solo lugar:** Steam devuelve centavos y se pasan a unidades al leer la respuesta, para que el resto del código trabaje siempre en la misma unidad.
- **Autorización por fila:** las consultas de modificación y borrado filtran también por `user_id`, así un usuario no puede tocar juegos ajenos.
- **Un juego por usuario:** `UNIQUE (user_id, steam_app_id)` evita duplicados, y la API responde `409` si se intenta repetir.
- **Worker en el mismo proceso:** una goroutine con `time.Ticker`. Alcanza para esta escala; si crece, se puede mover a un proceso separado.

## Pendiente / ideas

- Historial de precios (`price_history`) y gráfico por juego.
- Notificaciones por email.
- Avisar al agregar un juego que ya está por debajo del objetivo.
- Aviso sin objetivo ("avisame de cualquier baja").
- Catálogo normalizado de juegos (separar datos del juego de la relación con el usuario).
- Recomendaciones personalizadas.

## Aviso sobre IA
- No se utilizaron agentes de IA para realizar este proyecto. SI se utilizo el LLM de Claude para debug, explicar conceptos y sugerir correcciones/mejorar y generar documentación.
