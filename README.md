# n-backoffice-api

API REST en Go para administrar agentes, workflows, pasos y notificaciones. Usa Gin como framework HTTP y GORM como ORM, con conexión MySQL configurada por variables de entorno.

## Requisitos
- Go 1.22+
- MySQL accesible (local o remoto)

## Configuración
1) Copiá los valores de entorno (se cargan con `godotenv` si existe un archivo `.env`):
```
DB_USER=nuser
DB_PASS=npass
DB_HOST=localhost
DB_PORT=3306
DB_NAME=ndb
```
2) Ejecutá la aplicación:
```
go run ./main.go
```
El servidor levanta en `:8080` y realiza `AutoMigrate` de los modelos `Agent`, `Workflow`, `Step` y `N`.

## Endpoints principales
- `POST /agents`, `GET /agents`, `PUT /agents/:id`, `DELETE /agents/:id`
- `POST /workflows`, `GET /workflows`, `PUT /workflows/:id`, `DELETE /workflows/:id`
- `POST /steps`, `GET /steps`, `PUT /steps/:id`, `DELETE /steps/:id`
- `POST /n`, `GET /n`, `PUT /n/:id`, `DELETE /n/:id`

Notas:
- Paginación opcional en los listados vía query params `page` y `size` (por defecto 1 y 10).
- Los modelos usan IDs autoincrementales (`uint64`).

## Pruebas
Ejecutá todos los tests (se usa SQLite en memoria, no requiere MySQL):
```
go test ./...
```

## Estructura del proyecto
- `config/`: carga de configuración y variables de entorno.
- `client/`: inicialización de la conexión a base de datos.
- `handler/`: definiciones de rutas y controladores HTTP.
- `model/`: modelos GORM.
- `service/`: lógica auxiliar (paginación, etc.).

