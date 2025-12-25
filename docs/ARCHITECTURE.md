# Chaos API Proxy - Project Structure

```
chaos-api-proxy/
│
├── cmd/                          # Comandos principales
│   └── server/
│       └── main.go              # Punto de entrada del servidor
│
├── internal/                     # Código privado de la aplicación
│   ├── chaos/                   # Motor de caos
│   │   ├── engine.go            # Lógica de decisiones de chaos
│   │   └── engine_test.go       # Tests del motor
│   │
│   ├── config/                  # Configuración
│   │   └── config.go            # Carga de env vars
│   │
│   ├── handler/                 # HTTP handlers
│   │   ├── config.go            # API de configuraciones (CRUD)
│   │   └── proxy.go             # Proxy inverso con chaos
│   │
│   ├── models/                  # Modelos de datos
│   │   ├── chaos_config.go      # Estructura ChaosConfig
│   │   └── chaos_config_test.go # Tests de modelos
│   │
│   └── storage/                 # Capa de persistencia
│       └── redis.go             # Implementación Redis
│
├── examples/                     # Ejemplos de uso
│   ├── configs/                 # Configuraciones de ejemplo
│   │   ├── high-latency.json
│   │   ├── rate-limit.json
│   │   ├── network-flakiness.json
│   │   └── slow-download.json
│   └── demo.sh                  # Script de demostración
│
├── docs/                        # Documentación
│   ├── API.md                   # Referencia de API
│   ├── EXAMPLES.md              # Guía de ejemplos
│   └── INSTALLATION.md          # Guía de instalación
│
├── .env.example                 # Variables de entorno de ejemplo
├── .gitignore                   # Archivos ignorados por Git
├── docker-compose.yml           # Orquestación Docker
├── Dockerfile                   # Imagen Docker del proxy
├── go.mod                       # Dependencias de Go
├── go.sum                       # Checksums de dependencias
├── LICENSE                      # Licencia MIT
├── Makefile                     # Comandos de desarrollo
└── README.md                    # Documentación principal

```

## Descripción de Directorios

### `/cmd`

Contiene los puntos de entrada de la aplicación. Cada subdirectorio es un ejecutable diferente.

### `/internal`

Código privado de la aplicación que no puede ser importado por otros proyectos Go.

- **chaos/**: Motor de decisiones (qué chaos inyectar y cuándo)
- **config/**: Carga de configuración desde environment
- **handler/**: HTTP request handlers
- **models/**: Estructuras de datos y lógica de negocio
- **storage/**: Abstracción de persistencia (actualmente Redis)

### `/examples`

Ejemplos listos para usar y script de demostración.

### `/docs`

Documentación detallada del proyecto.

## Flujo de Datos

```
[Cliente] → [Proxy Handler] → [Get Config from Redis]
                ↓
          [Chaos Engine]
                ↓
     [Decision: Inject Chaos?]
                ↓
        ┌───────┴────────┐
        ↓                ↓
    [Inject]         [No Inject]
        ↓                ↓
   [Return Error]   [ReverseProxy]
                         ↓
                    [Target API]
                         ↓
                    [Response]
                         ↓
                    [Cliente]
```

## Componentes Clave

### 1. Main Server (`cmd/server/main.go`)

- Inicializa logger
- Carga configuración
- Conecta a Redis
- Setup del router (Gorilla Mux)
- Graceful shutdown

### 2. Proxy Handler (`internal/handler/proxy.go`)

- Recibe petición con config ID
- Obtiene configuración desde Redis
- Consulta al Chaos Engine
- Aplica chaos o proxea la petición
- Maneja bandwidth limiting

### 3. Chaos Engine (`internal/chaos/engine.go`)

- Toma decisiones probabilísticas
- Calcula latencias con jitter
- Decide si inyectar errores
- Gestiona drop connections

### 4. Config Handler (`internal/handler/config.go`)

- CRUD de configuraciones
- Validación
- Serialización JSON

### 5. Storage (`internal/storage/redis.go`)

- Interfaz con Redis
- Operaciones CRUD
- Keys: `chaos:config:{id}`
- Set: `chaos:configs`

## Decisiones de Diseño

### ¿Por qué Go?

- Excelente para proxies (httputil.ReverseProxy)
- Concurrencia nativa (goroutines)
- Rendimiento superior
- Typed y compilado

### ¿Por qué Redis?

- Latencia ultra-baja (< 1ms)
- Acceso en memoria
- Simple para este caso de uso
- Fácil de escalar

### ¿Por qué Gorilla Mux?

- Path variables fáciles (`/proxy/{configID}`)
- Middleware support
- Estable y probado

## Extensibilidad

### Añadir Nuevo Tipo de Chaos

1. Añadir campo a `ChaosRules` en `models/chaos_config.go`
2. Actualizar `engine.MakeDecision()` en `chaos/engine.go`
3. Implementar lógica en `proxy.go`
4. Añadir tests

Ejemplo: Chaos de "Respuesta Parcial"

```go
// models/chaos_config.go
type ChaosRules struct {
    // ... existing fields
    TruncateResponseRate float64 `json:"truncate_response_rate,omitempty"`
}

// chaos/engine.go
func (e *Engine) MakeDecision(rules ChaosRules) *Decision {
    // ... existing logic

    if rules.TruncateResponseRate > 0 && e.rng.Float64() < rules.TruncateResponseRate {
        decision.ShouldTruncateResponse = true
    }

    return decision
}

// handler/proxy.go
if decision.ShouldTruncateResponse {
    // Implementar lógica de truncado
}
```

### Añadir Nuevo Storage Backend

Crear interfaz en `storage/`:

```go
// storage/interface.go
type Storage interface {
    SaveConfig(ctx context.Context, config *models.ChaosConfig) error
    GetConfig(ctx context.Context, id string) (*models.ChaosConfig, error)
    ListConfigs(ctx context.Context) ([]*models.ChaosConfig, error)
    DeleteConfig(ctx context.Context, id string) error
    UpdateConfig(ctx context.Context, config *models.ChaosConfig) error
}

// storage/postgres.go
type PostgresStorage struct { ... }
func (s *PostgresStorage) SaveConfig(...) { ... }
```

### Añadir Autenticación

Crear middleware en `internal/middleware/`:

```go
// middleware/auth.go
func APIKeyAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validar API key
        next.ServeHTTP(w, r)
    })
}

// cmd/server/main.go
api := router.PathPrefix("/api/v1").Subrouter()
api.Use(middleware.APIKeyAuth)
```

## Testing

```bash
# Unit tests
go test ./internal/...

# Con coverage
go test -coverprofile=coverage.txt ./internal/...
go tool cover -html=coverage.txt

# Benchmark
go test -bench=. ./internal/chaos/
```

## Performance

### Benchmarks Esperados

- **Latencia del proxy (sin chaos)**: < 5ms overhead
- **Throughput**: > 1000 req/s (hardware moderno)
- **Memoria**: ~50MB en idle

### Optimizaciones

1. **Connection pooling** en ReverseProxy (automático en Go)
2. **Redis pipeline** para múltiples gets
3. **Config caching** en memoria (TTL)
4. **Streaming** para responses grandes

---

**Contribuye** mejorando la arquitectura en [GitHub](https://github.com/medalcode/chaos-api-proxy) 🚀
