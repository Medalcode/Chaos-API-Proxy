# 🔌 API Reference

## Base URL

```
http://localhost:8081
```

## Endpoints

### Health Check

#### `GET /health`

Verifica el estado del servidor.

**Response**

```json
{
  "status": "healthy"
}
```

**Status Codes**

- `200 OK` - Servidor funcionando correctamente

---

## Configuration Management API

### Create Configuration

#### `POST /api/v1/configs`

Crea una nueva configuración de chaos engineering.

**Request Headers**

```
Content-Type: application/json
```

**Request Body**

```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "target": "string (required) - URL completa de la API target",
  "enabled": "boolean (optional, default: true)",
  "rules": {
    "latency_ms": "integer (optional)",
    "jitter": "integer (optional)",
    "inject_failure_rate": "float (optional, 0.0-1.0)",
    "error_code": "integer (optional, HTTP status code)",
    "error_body": "string (optional, JSON string)",
    "drop_connection": "boolean (optional)",
    "drop_connection_rate": "float (optional, 0.0-1.0)",
    "bandwidth_limit_kbps": "integer (optional)",
    "modify_headers": "object (optional)",
    "remove_headers": "array of strings (optional)"
  }
}
```

**Example Request**

```bash
curl -X POST http://localhost:8081/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Stripe API Test",
    "description": "Simula latencia para pagos",
    "target": "https://api.stripe.com",
    "enabled": true,
    "rules": {
      "latency_ms": 500,
      "jitter": 200,
      "inject_failure_rate": 0.1,
      "error_code": 503
    }
  }'
```

**Response** (`201 Created`)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Stripe API Test",
  "description": "Simula latencia para pagos",
  "target": "https://api.stripe.com",
  "enabled": true,
  "created_at": "2025-12-25T17:00:00Z",
  "updated_at": "2025-12-25T17:00:00Z",
  "rules": {
    "latency_ms": 500,
    "jitter": 200,
    "inject_failure_rate": 0.1,
    "error_code": 503,
    "error_body": ""
  }
}
```

**Error Responses**

`400 Bad Request` - Validación fallida

```json
{
  "error": "target URL is required"
}
```

`500 Internal Server Error` - Error del servidor

```json
{
  "error": "Failed to save configuration"
}
```

---

### List Configurations

#### `GET /api/v1/configs`

Obtiene todas las configuraciones almacenadas.

**Example Request**

```bash
curl http://localhost:8081/api/v1/configs
```

**Response** (`200 OK`)

```json
{
  "configs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Stripe API Test",
      "target": "https://api.stripe.com",
      "enabled": true,
      "created_at": "2025-12-25T17:00:00Z",
      "updated_at": "2025-12-25T17:00:00Z",
      "rules": { ... }
    },
    {
      "id": "660f9511-f30c-52e5-b827-557766551111",
      "name": "GitHub API Test",
      "target": "https://api.github.com",
      "enabled": false,
      "created_at": "2025-12-25T18:00:00Z",
      "updated_at": "2025-12-25T18:00:00Z",
      "rules": { ... }
    }
  ],
  "count": 2
}
```

---

### Get Configuration

#### `GET /api/v1/configs/{id}`

Obtiene una configuración específica por ID.

**Path Parameters**

- `id` (string, required) - UUID de la configuración

**Example Request**

```bash
curl http://localhost:8081/api/v1/configs/550e8400-e29b-41d4-a716-446655440000
```

**Response** (`200 OK`)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Stripe API Test",
  "description": "Simula latencia para pagos",
  "target": "https://api.stripe.com",
  "enabled": true,
  "created_at": "2025-12-25T17:00:00Z",
  "updated_at": "2025-12-25T17:00:00Z",
  "rules": {
    "latency_ms": 500,
    "jitter": 200,
    "inject_failure_rate": 0.1,
    "error_code": 503
  }
}
```

**Error Responses**

`404 Not Found` - Configuración no existe

```json
{
  "error": "Configuration not found"
}
```

---

### Update Configuration

#### `PUT /api/v1/configs/{id}`

Actualiza una configuración existente.

**Path Parameters**

- `id` (string, required) - UUID de la configuración

**Request Body**

```json
{
  "name": "string",
  "description": "string",
  "target": "string",
  "enabled": "boolean",
  "rules": { ... }
}
```

**Example Request**

```bash
curl -X PUT http://localhost:8081/api/v1/configs/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Stripe API Test",
    "target": "https://api.stripe.com",
    "enabled": false,
    "rules": {
      "latency_ms": 1000
    }
  }'
```

**Response** (`200 OK`)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Stripe API Test",
  "target": "https://api.stripe.com",
  "enabled": false,
  "updated_at": "2025-12-25T18:30:00Z",
  "rules": {
    "latency_ms": 1000
  }
}
```

---

### Delete Configuration

#### `DELETE /api/v1/configs/{id}`

Elimina una configuración.

**Path Parameters**

- `id` (string, required) - UUID de la configuración

**Example Request**

```bash
curl -X DELETE http://localhost:8081/api/v1/configs/550e8400-e29b-41d4-a716-446655440000
```

**Response** (`200 OK`)

```json
{
  "message": "Configuration deleted successfully",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## Proxy Endpoint

### Proxy Request

#### `ANY /proxy/{configID}/*`

Proxy transparente con inyección de chaos. Acepta cualquier método HTTP (GET, POST, PUT, DELETE, etc.)

**Path Parameters**

- `configID` (string, required) - UUID de la configuración a usar
- `*` - Path completo de la API target

**Headers**

- Todos los headers de la petición original se reenvían

**Body**

- El body de la petición original se reenvía sin modificar

**Example Requests**

```bash
# GET request
curl http://localhost:8081/proxy/550e8400-e29b-41d4-a716-446655440000/v1/charges

# POST request
curl -X POST http://localhost:8081/proxy/550e8400-e29b-41d4-a716-446655440000/v1/charges \
  -H "Authorization: Bearer sk_test_..." \
  -H "Content-Type: application/json" \
  -d '{"amount": 1000, "currency": "usd"}'
```

**URL Mapping**

```
Proxy URL:
http://localhost:8081/proxy/{configID}/v1/charges

Se mapea a:
{target}/v1/charges

Ejemplo:
http://localhost:8081/proxy/550e.../v1/charges
→ https://api.stripe.com/v1/charges
```

**Response Headers Added**

El proxy añade headers de trazabilidad:

```
X-Chaos-Proxy: true
X-Chaos-Proxy-Config-ID: 550e8400-e29b-41d4-a716-446655440000
```

Si se inyectó chaos:

```
X-Chaos-Proxy-Injected: true
X-Chaos-Proxy-Type: error | latency
X-Chaos-Proxy-Latency-Ms: 523ms
```

**Error Responses**

`404 Not Found` - Configuración no existe

```json
{
  "error": "Configuration not found"
}
```

`403 Forbidden` - Configuración deshabilitada

```json
{
  "error": "Configuration is disabled"
}
```

`502 Bad Gateway` - Error al contactar API target

```
X-Chaos-Proxy-Error: true
Proxy error
```

`503 Service Unavailable` - Conexión cerrada (drop_connection)

---

## Chaos Rules Reference

### Rules Object

```json
{
  "rules": {
    "latency_ms": 500,
    "jitter": 200,
    "inject_failure_rate": 0.1,
    "error_code": 503,
    "error_body": "{\"error\": \"Custom error message\"}",
    "drop_connection": false,
    "drop_connection_rate": 0.05,
    "bandwidth_limit_kbps": 100,
    "modify_headers": {
      "X-Custom-Header": "value"
    },
    "remove_headers": ["Authorization"]
  }
}
```

### Rule Parameters

| Parameter              | Type    | Range   | Description                                 |
| ---------------------- | ------- | ------- | ------------------------------------------- |
| `latency_ms`           | integer | 0+      | Latencia fija en milisegundos               |
| `jitter`               | integer | 0+      | Variación aleatoria de latencia (±ms)       |
| `inject_failure_rate`  | float   | 0.0-1.0 | Probabilidad de error (0.1 = 10%)           |
| `error_code`           | integer | 100-599 | Código HTTP del error                       |
| `error_body`           | string  | -       | Cuerpo de respuesta del error (JSON string) |
| `drop_connection`      | boolean | -       | Cerrar socket sin responder                 |
| `drop_connection_rate` | float   | 0.0-1.0 | Probabilidad de desconexión                 |
| `bandwidth_limit_kbps` | integer | 1+      | Límite de ancho de banda (KB/s)             |
| `modify_headers`       | object  | -       | Headers a añadir/modificar                  |
| `remove_headers`       | array   | -       | Headers a eliminar                          |

### Rule Behavior

#### Latency Calculation

```
final_latency = latency_ms + random(-jitter, +jitter)
```

Ejemplo:

- `latency_ms: 500, jitter: 200`
- Resultado: entre 300ms y 700ms

#### Error Injection

1. Genera número aleatorio [0.0, 1.0)
2. Si < `inject_failure_rate`, inyecta error
3. Devuelve `error_code` con `error_body`

#### Connection Drop

Si `drop_connection: true` o random < `drop_connection_rate`:

1. Cierra conexión inmediatamente
2. No devuelve respuesta
3. Cliente recibe error de red/timeout

#### Bandwidth Limiting

```
delay_per_chunk = (chunk_size_bytes / (bandwidth_limit_kbps * 1024)) seconds
```

Simula descarga lenta añadiendo delays entre chunks.

---

## Rate Limiting

Actualmente el proxy **NO** tiene rate limiting propio. En producción se recomienda:

```nginx
# NGINX rate limiting
limit_req_zone $binary_remote_addr zone=chaos:10m rate=10r/s;

location /api/v1/ {
    limit_req zone=chaos burst=20;
    proxy_pass http://chaos-proxy:8080;
}
```

---

## Authentication

⚠️ **El proxy actual NO tiene autenticación.**

Para producción, implementar uno de estos:

### API Key (Header)

```go
func AuthMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    apiKey := r.Header.Get("X-API-Key")
    if apiKey != os.Getenv("EXPECTED_API_KEY") {
      http.Error(w, "Unauthorized", http.StatusUnauthorized)
      return
    }
    next.ServeHTTP(w, r)
  })
}
```

### JWT

```go
import "github.com/golang-jwt/jwt/v5"

func JWTMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tokenString := extractToken(r)
    token, err := jwt.Parse(tokenString, keyFunc)

    if err != nil || !token.Valid {
      http.Error(w, "Unauthorized", http.StatusUnauthorized)
      return
    }

    next.ServeHTTP(w, r)
  })
}
```

---

## WebSocket Support

**Status**: ❌ No soportado actualmente

Para proxying de WebSockets, se necesitaría implementar:

```go
import "github.com/gorilla/websocket"

// TODO: WebSocket proxy handler
```

---

## Examples

Ver [EXAMPLES.md](./EXAMPLES.md) para ejemplos completos de uso.

---

## Changelog

### v1.0.0 (2025-12-25)

- ✅ Proxy inverso básico
- ✅ Inyección de latencia con jitter
- ✅ Inyección de errores
- ✅ Drop connection
- ✅ Bandwidth limiting
- ✅ Header modification
- ✅ API REST para configuraciones
- ✅ Storage en Redis
- ✅ Logging estructurado
- ✅ Docker support

---

**Contribuciones bienvenidas** 🚀
