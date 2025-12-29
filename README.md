# 🌪️ Chaos API Proxy

> **Simulación de Caos y Latencia para Chaos Engineering**

Un proxy inverso inteligente que actúa como "man-in-the-middle" entre tu aplicación y APIs externas, permitiéndote inyectar fallas controladas, latencia, y otras condiciones adversas para probar la resiliencia de tu sistema.

## 🎯 ¿Qué Hace?

El Chaos API Proxy intercepta peticiones HTTP/HTTPS hacia APIs externas y aplica "reglas de caos" configurables:

- ✅ **Inyección de Latencia**: Simula conexiones lentas con latencia fija o variable (jitter)
- ✅ **Inyección de Errores**: Devuelve errores HTTP configurables (500, 503, 429, etc.)
- ✅ **Desconexión de Socket**: Cierra conexiones sin responder para simular timeouts
- ✅ **Limitación de Ancho de Banda**: Simula conexiones lentas con throttling configurable
- ✅ **Modificación de Headers**: Añade, modifica o elimina headers HTTP
- ✅ **Control Probabilístico**: Configura tasas de falla (ej: 10% de peticiones fallan)

## 🏗️ Arquitectura

```
┌─────────────┐      ┌──────────────────┐      ┌─────────────┐
│   Tu App    │─────▶│  Chaos Proxy     │─────▶│  API Real   │
│  (Cliente)  │      │  (Man-in-Middle) │      │ (Stripe,    │
└─────────────┘      │                  │      │  etc.)      │
                     │  ┌────────────┐  │      └─────────────┘
                     │  │ Motor de   │  │
                     │  │ Caos       │  │
                     │  └────────────┘  │
                     │  ┌────────────┐  │
                     │  │   Redis    │  │
                     │  │ (Configs)  │  │
                     │  └────────────┘  │
                     └──────────────────┘
```

### Componentes:

- **Proxy Inverso**: Basado en `httputil.ReverseProxy` de Go
- **Motor de Caos**: Toma decisiones probabilísticas sobre qué inyectar
- **Storage (Redis)**: Configuraciones de reglas con acceso ultrarrápido
- **API REST**: Gestión de configuraciones (CRUD)

## 🚀 Quick Start

### Con Docker Compose (Recomendado)

```bash
# 1. Clonar el repositorio
git clone https://github.com/medalcode/chaos-api-proxy.git
cd chaos-api-proxy

# 2. Iniciar servicios (Redis + Proxy)
docker-compose up -d

# 3. Verificar que está corriendo
curl http://localhost:8081/health
```

### Desarrollo Local

```bash
# 1. Instalar dependencias
make deps

# 2. Iniciar Redis
docker run -d -p 6379:6379 redis:7-alpine

# 3. Ejecutar el proxy
make run
```

## 📖 Uso

### 1. Crear una Configuración de Caos

```bash
curl -X POST http://localhost:8081/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Stripe API Chaos Test",
    "description": "Simula latencia y errores en Stripe",
    "target": "https://api.stripe.com",
    "enabled": true,
    "rules": {
      "latency_ms": 500,
      "jitter": 200,
      "inject_failure_rate": 0.10,
      "error_code": 503,
      "error_body": "{\"error\": \"Service temporarily unavailable\"}"
    }
  }'
```

Respuesta:

```json
{
  "id": "abc123-def456-ghi789",
  "name": "Stripe API Chaos Test",
  "target": "https://api.stripe.com",
  "enabled": true,
  "created_at": "2025-12-25T17:00:00Z",
  "rules": { ... }
}
```

### 2. Usar el Proxy en tu Aplicación

En lugar de llamar directamente a la API:

```javascript
// ❌ Antes (directo)
fetch("https://api.stripe.com/v1/charges", {
  method: "POST",
  headers: { Authorization: "Bearer sk_test_..." },
});
```

Apunta al proxy (dos métodos disponibles):

**Método 1: Path-Based (Recomendado)**

```javascript
// ✅ Path-based: Config ID en la URL
fetch("http://localhost:8081/proxy/abc123-def456-ghi789/v1/charges", {
  method: "POST",
  headers: { Authorization: "Bearer sk_test_..." },
});
```

**Método 2: Header-Based**

```javascript
// ✅ Header-based: Config ID en header
fetch("http://localhost:8081/v1/charges", {
  method: "POST",
  headers: {
    "X-Chaos-Config-ID": "abc123-def456-ghi789",
    Authorization: "Bearer sk_test_...",
  },
});
```

> 📘 **Dual Routing Mode**: Ver [docs/DUAL_ROUTING.md](docs/DUAL_ROUTING.md) para más detalles sobre ambos métodos.

El proxy:

1. Recibe tu petición
2. Aplica las reglas de caos (latencia, errores, etc.)
3. Si no inyecta error, redirige a `https://api.stripe.com/v1/charges`
4. Devuelve la respuesta (con headers de trazabilidad)

### 3. Gestionar Configuraciones

```bash
# Listar todas las configuraciones
curl http://localhost:8081/api/v1/configs

# Obtener una configuración específica
curl http://localhost:8081/api/v1/configs/abc123-def456-ghi789

# Actualizar configuración
curl -X PUT http://localhost:8081/api/v1/configs/abc123-def456-ghi789 \
  -H "Content-Type: application/json" \
  -d '{ "enabled": false }'

# Eliminar configuración
curl -X DELETE http://localhost:8081/api/v1/configs/abc123-def456-ghi789
```

## 🎛️ Matriz de Variables de Caos

| Parámetro              | Tipo   | Descripción                           | Ejemplo de Uso                    |
| ---------------------- | ------ | ------------------------------------- | --------------------------------- |
| `latency_ms`           | int    | Latencia fija en milisegundos         | Simular conexión 3G lenta (500ms) |
| `jitter`               | int    | Variación aleatoria de latencia (±ms) | Latencia entre 300ms y 700ms      |
| `inject_failure_rate`  | float  | Probabilidad de error (0.0-1.0)       | 10% de peticiones fallan          |
| `error_code`           | int    | Código HTTP del error                 | 429 (Too Many Requests)           |
| `error_body`           | string | Cuerpo de respuesta del error         | JSON personalizado                |
| `drop_connection`      | bool   | Cerrar socket sin responder           | Simular timeout de red            |
| `drop_connection_rate` | float  | Probabilidad de desconexión           | 5% de peticiones se cierran       |
| `bandwidth_limit_kbps` | int    | Límite de ancho de banda (KB/s)       | Simular descarga lenta (100 KB/s) |
| `modify_headers`       | map    | Headers a añadir/modificar            | `{"X-Custom": "value"}`           |
| `remove_headers`       | array  | Headers a eliminar                    | `["Authorization"]`               |

## 📊 Ejemplos de Configuraciones

### Latencia Variable con Jitter

```json
{
  "name": "Conexión 3G Inestable",
  "target": "https://api.example.com",
  "rules": {
    "latency_ms": 500,
    "jitter": 300
  }
}
```

Resultado: Latencia entre 200ms y 800ms

### Rate Limiting Simulation

```json
{
  "name": "Simular Rate Limit",
  "target": "https://api.example.com",
  "rules": {
    "inject_failure_rate": 0.2,
    "error_code": 429,
    "error_body": "{\"error\": \"Too many requests\"}"
  }
}
```

### Timeout por Desconexión

```json
{
  "name": "Network Flakiness",
  "target": "https://api.example.com",
  "rules": {
    "drop_connection_rate": 0.05
  }
}
```

### Descarga Lenta

```json
{
  "name": "Ancho de Banda Limitado",
  "target": "https://cdn.example.com",
  "rules": {
    "bandwidth_limit_kbps": 100,
    "latency_ms": 200
  }
}
```

## 🔍 Headers de Trazabilidad

El proxy añade headers para identificar peticiones procesadas:

```http
X-Chaos-Proxy: true
X-Chaos-Proxy-Config-ID: abc123-def456-ghi789
X-Chaos-Proxy-Injected: true (si se inyectó caos)
X-Chaos-Proxy-Type: error | latency
X-Chaos-Proxy-Latency-Ms: 523ms
```

Úsalos para debugging en tu aplicación:

```javascript
fetch(url).then((res) => {
  if (res.headers.get("X-Chaos-Proxy-Injected")) {
    console.log(
      "⚠️ Esta respuesta fue afectada por caos:",
      res.headers.get("X-Chaos-Proxy-Type")
    );
  }
});
```

## 🖥️ Web Dashboard

Gestiona tus reglas visualmente en: `http://localhost:8081/dashboard`

![Dashboard Preview](docs/assets/dashboard-preview.png) (Ver `docs/DASHBOARD_AND_SECURITY.md`)

## 🔐 Seguridad

Puedes proteger la API de administración con API Keys usando `CHAOS_API_KEYS`.
Ver guía detallada en [docs/DASHBOARD_AND_SECURITY.md](docs/DASHBOARD_AND_SECURITY.md).

## 📊 Observabilidad (Prometheus)

El proxy expone métricas detalladas en `/metrics`.
Consulta la [Guía de Métricas](docs/METRICS.md).

## 🛠️ Development

```bash
# Instalar dependencias
make deps

# Ejecutar tests
make test

# Ver cobertura
make test-coverage

# Linter
make lint

# Hot reload en desarrollo
make dev

# Ver logs de Docker
make docker-logs
```

## 🏗️ Estructura del Proyecto

```
chaos-api-proxy/
├── cmd/
│   └── server/          # Punto de entrada
│       └── main.go
├── internal/
│   ├── chaos/           # Motor de caos
│   │   └── engine.go
│   ├── config/          # Configuración
│   │   └── config.go
│   ├── handler/         # HTTP handlers
│   │   ├── config.go   # API de configuraciones
│   │   └── proxy.go    # Proxy inverso
│   ├── models/          # Modelos de datos
│   │   └── chaos_config.go
│   └── storage/         # Persistencia
│       └── redis.go
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

## 🌐 API Reference

### Configurations API

#### `POST /api/v1/configs`

Crear una nueva configuración de caos.

**Request Body:**

```json
{
  "name": "string",
  "description": "string (optional)",
  "target": "string (required)",
  "enabled": "boolean",
  "rules": { ... }
}
```

**Response:** `201 Created`

```json
{
  "id": "uuid",
  "name": "string",
  "target": "string",
  "enabled": true,
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "rules": { ... }
}
```

#### `GET /api/v1/configs`

Listar todas las configuraciones.

**Response:** `200 OK`

```json
{
  "configs": [ ... ],
  "count": 5
}
```

#### `GET /api/v1/configs/{id}`

Obtener una configuración específica.

**Response:** `200 OK`

#### `PUT /api/v1/configs/{id}`

Actualizar configuración.

**Response:** `200 OK`

#### `DELETE /api/v1/configs/{id}`

Eliminar configuración.

**Response:** `200 OK`

### Proxy Endpoint

#### `ANY /proxy/{configID}/*path`

Proxy transparente con inyección de caos.

**Ejemplo:**

```
GET /proxy/abc123/v1/users
  → GET https://api.example.com/v1/users (con caos aplicado)
```

## 🔒 Seguridad

⚠️ **IMPORTANTE**: Este proxy es para **entornos de desarrollo y testing únicamente**.

- No tiene autenticación por defecto
- Expondrá las APIs target si es accesible públicamente
- No usar en producción sin añadir:
  - Autenticación (API keys, JWT, etc.)
  - Rate limiting propio
  - Logging de auditoría
  - TLS/HTTPS

## 📝 Casos de Uso

### Frontend Testing

```javascript
// Probar estados de carga
const response = await fetch(chaosProxyURL);
// El 50% de las veces tardará 2 segundos
```

### Circuit Breaker Testing

```java
// Verificar que tu circuit breaker se activa
// cuando el 30% de peticiones fallan
```

### Mobile App Testing

```swift
// Simular conexión lenta en WiFi público
// bandwidth_limit_kbps: 50
```

### Microservices Resilience

```go
// Probar retry logic con errores intermitentes
// inject_failure_rate: 0.15
```

## 🤝 Contribuir

1. Fork el repositorio
2. Crea una rama: `git checkout -b feature/nueva-funcionalidad`
3. Commit: `git commit -am 'Add nueva funcionalidad'`
4. Push: `git push origin feature/nueva-funcionalidad`
5. Abre un Pull Request

## 📄 Licencia

MIT License - Ver archivo [LICENSE](LICENSE)

## 🙏 Créditos

Desarrollado con ❤️ por [MedalCode](https://github.com/medalcode)

---

**¿Preguntas?** Abre un [issue](https://github.com/medalcode/chaos-api-proxy/issues) 🚀
