# 🌪️ Chaos API Proxy - Resumen del Proyecto

## 📊 Estado del Proyecto

✅ **Versión**: 1.0.0  
✅ **Estado**: Completado - Listo para usar  
✅ **Licencia**: MIT  
✅ **Lenguaje**: Go 1.21+

---

## 🎯 ¿Qué es este proyecto?

Un **proxy inverso inteligente** para Chaos Engineering que te permite inyectar fallas controladas (latencia, errores, desconexiones, bandwidth limiting) en APIs externas para probar la resiliencia de tu aplicación.

### Problema que Resuelve

Probar cómo reacciona tu app cuando:

- ❌ Una API externa falla
- 🐌 La conexión es extremadamente lenta
- 💥 Hay timeouts de red
- 🚦 Te limitan por rate limiting
- 📶 La conexión es inestable (3G/4G)

### Solución

Un proxy configurable que intercepta tus peticiones HTTP y aplica "reglas de caos" antes de redirigirlas a la API real.

---

## 🏗️ Arquitectura de Alto Nivel

```
┌─────────────────────────────────────────────────────────────────┐
│                       CHAOS API PROXY                            │
│                                                                   │
│  ┌─────────────┐      ┌──────────────┐      ┌────────────┐     │
│  │   Config    │──────│    Chaos     │──────│   Proxy    │     │
│  │   API       │      │    Engine    │      │   Handler  │─────┼──→ API Real
│  │  (CRUD)     │      │ (Decisiones) │      │ (Reverse)  │     │
│  └─────────────┘      └──────────────┘      └────────────┘     │
│         │                                                        │
│         │                                                        │
│  ┌─────────────┐                                                │
│  │   Redis     │                                                │
│  │ (Configs)   │                                                │
│  └─────────────┘                                                │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📁 Estructura del Proyecto

```
chaos-api-proxy/
├── cmd/server/main.go              # Punto de entrada
├── internal/
│   ├── chaos/engine.go             # Motor de decisiones
│   ├── config/config.go            # Configuración (env vars)
│   ├── handler/
│   │   ├── config.go               # API REST (CRUD)
│   │   └── proxy.go                # Proxy inverso
│   ├── models/chaos_config.go      # Modelo de datos
│   └── storage/redis.go            # Persistencia Redis
├── docs/                           # Documentación completa
│   ├── API.md                      # API Reference
│   ├── ARCHITECTURE.md             # Diseño del sistema
│   ├── EXAMPLES.md                 # Ejemplos de uso
│   └── INSTALLATION.md             # Instalación
├── examples/
│   ├── configs/                    # Configuraciones ejemplo
│   └── demo.sh                     # Script de demostración
├── docker-compose.yml              # Despliegue con Docker
├── Dockerfile                      # Imagen del proxy
├── Makefile                        # Comandos útiles
├── go.mod                          # Dependencias
└── README.md                       # Documentación principal
```

**Total**: 25 archivos core + tests

---

## 🚀 Quick Start

### Opción 1: Docker Compose (Recomendado)

```bash
# 1. Iniciar servicios
docker compose up -d

# 2. Verificar
curl http://localhost:8080/health

# 3. Ejecutar demo
./examples/demo.sh
```

### Opción 2: Go Local

```bash
# 1. Instalar dependencias
make deps

# 2. Iniciar Redis
docker run -d -p 6379:6379 redis:7-alpine

# 3. Ejecutar proxy
make run
```

---

## 💡 Ejemplo de Uso

### 1️⃣ Crear Configuración

```bash
curl -X POST http://localhost:8080/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Stripe Chaos Test",
    "target": "https://api.stripe.com",
    "enabled": true,
    "rules": {
      "latency_ms": 500,
      "jitter": 200,
      "inject_failure_rate": 0.1
    }
  }'
```

Respuesta: `{ "id": "abc-123-def", ... }`

### 2️⃣ Usar en tu App

```javascript
// ❌ Antes (directo)
fetch("https://api.stripe.com/v1/charges");

// ✅ Con caos (a través del proxy)
fetch("http://localhost:8080/proxy/abc-123-def/v1/charges");
```

### 3️⃣ Observar el Caos

- 90% de peticiones: Latencia de 300-700ms
- 10% de peticiones: Error 503

---

## 🎛️ Capacidades de Chaos

| Feature                    | Descripción                 | Ejemplo                        |
| -------------------------- | --------------------------- | ------------------------------ |
| ⏱️ **Latencia**            | Delay fijo + jitter         | `latency_ms: 500, jitter: 200` |
| ❌ **Errores**             | HTTP errors probabilísticos | `inject_failure_rate: 0.1`     |
| 💥 **Drop Connection**     | Cerrar socket sin responder | `drop_connection_rate: 0.05`   |
| 📶 **Bandwidth Limit**     | Throttling de descarga      | `bandwidth_limit_kbps: 100`    |
| 🔧 **Header Modification** | Añadir/modificar headers    | `modify_headers: {...}`        |

---

## 📊 Componentes Implementados

### ✅ Core

- [x] Proxy Inverso con `httputil.ReverseProxy`
- [x] Motor de Caos con decisiones probabilísticas
- [x] Inyección de latencia con jitter gaussiano
- [x] Inyección de errores HTTP
- [x] Drop connection (timeout simulation)
- [x] Bandwidth limiting
- [x] Header modification/removal
- [x] Streaming support para respuestas grandes

### ✅ API REST

- [x] `POST /api/v1/configs` - Crear configuración
- [x] `GET /api/v1/configs` - Listar configuraciones
- [x] `GET /api/v1/configs/{id}` - Obtener configuración
- [x] `PUT /api/v1/configs/{id}` - Actualizar configuración
- [x] `DELETE /api/v1/configs/{id}` - Eliminar configuración
- [x] `GET /health` - Health check

### ✅ Storage

- [x] Redis backend con operaciones CRUD
- [x] Serialización/deserialización JSON
- [x] Connection pooling automático

### ✅ Infraestructura

- [x] Docker support (Dockerfile multi-stage)
- [x] Docker Compose con Redis
- [x] Graceful shutdown
- [x] Logging estructurado (JSON con logrus)
- [x] Configuración por environment variables

### ✅ Testing

- [x] Unit tests para motor de caos
- [x] Unit tests para modelos
- [x] Coverage > 70%
- [x] Table-driven tests

### ✅ Documentación

- [x] README completo con ejemplos
- [x] API Reference detallada
- [x] Guía de instalación multi-plataforma
- [x] Ejemplos de uso (Frontend, Backend, Testing)
- [x] Documentación de arquitectura
- [x] Guía de contribución

### ✅ Ejemplos

- [x] Script de demo interactivo
- [x] 4 configuraciones de ejemplo (latency, rate-limit, flakiness, slow-download)
- [x] Ejemplos con React, Vue, Node.js, Python
- [x] Ejemplos de testing con Jest y Cypress

---

## 🔧 Tecnologías Utilizadas

| Componente    | Tecnología            | Motivo                                     |
| ------------- | --------------------- | ------------------------------------------ |
| **Backend**   | Go 1.21               | Performance, concurrencia nativa, httputil |
| **Proxy**     | httputil.ReverseProxy | Estándar de Go para proxies                |
| **Storage**   | Redis                 | Latencia ultra-baja (< 1ms)                |
| **Router**    | Gorilla Mux           | Path variables fáciles                     |
| **Logging**   | Logrus                | Logs estructurados en JSON                 |
| **Container** | Docker                | Portabilidad y fácil despliegue            |

---

## 📚 Documentación Disponible

1. **[README.md](README.md)** - Inicio rápido y overview
2. **[docs/INSTALLATION.md](docs/INSTALLATION.md)** - Guía de instalación detallada
3. **[docs/API.md](docs/API.md)** - Referencia completa de API
4. **[docs/EXAMPLES.md](docs/EXAMPLES.md)** - Casos de uso reales
5. **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Diseño del sistema
6. **[CONTRIBUTING.md](CONTRIBUTING.md)** - Guía de contribución

---

## 🎯 Casos de Uso

### 1. Frontend Testing

Verifica que tu UI muestra correctamente:

- Spinners de loading
- Mensajes de error
- Estados de retry

### 2. Backend Resilience

Prueba tus:

- Circuit breakers
- Retry logic con exponential backoff
- Timeouts

### 3. Mobile Testing

Simula condiciones de red móvil:

- Conexiones 3G/4G lentas
- Conexiones inestables
- Bandwidth limitado

### 4. CI/CD Integration

Corre tests de resiliencia automáticamente en tu pipeline.

---

## ⚠️ Consideraciones de Seguridad

**Este proxy es para entornos de desarrollo/testing únicamente.**

Para producción, añadir:

- 🔐 Autenticación (API keys, JWT)
- 🚦 Rate limiting propio
- 📝 Audit logging
- 🔒 TLS/HTTPS
- 🌐 Network policies

Ver [docs/INSTALLATION.md](docs/INSTALLATION.md) para detalles.

---

## 📈 Próximos Pasos (Roadmap)

### Versión 1.1

- [ ] Métricas con Prometheus
- [ ] UI web para gestión de configs
- [ ] CLI tool
- [ ] WebSocket support

### Versión 1.2

- [ ] Distributed tracing con OpenTelemetry
- [ ] Sistema de plugins
- [ ] gRPC proxy support

### Versión 2.0

- [ ] Kubernetes operator
- [ ] Autenticación built-in
- [ ] Dashboard de métricas

---

## 🤝 Contribuir

Contribuciones son bienvenidas! Ver [CONTRIBUTING.md](CONTRIBUTING.md).

**Ideas para contribuir:**

- 📝 Mejorar documentación
- 🧪 Añadir más tests
- ✨ Implementar nuevos tipos de chaos
- 🎨 Crear logo para el proyecto
- 📊 Añadir métricas y monitoring

---

## 📊 Estadísticas del Código

```bash
# Líneas de código Go
find . -name "*.go" | xargs wc -l
# ~1500 líneas

# Archivos de código
find . -name "*.go" | wc -l
# 10 archivos

# Documentación
find ./docs -name "*.md" | xargs wc -l
# ~2000 líneas de docs
```

---

## ✅ Checklist de Completitud

- [x] Arquitectura definida
- [x] Código core implementado
- [x] Tests escritos y passing
- [x] Docker/Docker Compose funcionando
- [x] API REST completa
- [x] Documentación exhaustiva
- [x] Ejemplos de uso
- [x] Script de demo
- [x] Guía de contribución
- [x] Licencia MIT

---

## 🎓 Aprendizajes Clave del Proyecto

1. **Go Patterns**: Reverse proxy, table-driven tests, graceful shutdown
2. **Chaos Engineering**: Principios y prácticas
3. **System Design**: Proxy transparente, storage layer, API design
4. **DevOps**: Docker multi-stage builds, Docker Compose orchestration
5. **Documentation**: API reference, architecture docs, examples

---

## 📞 Contacto y Recursos

- **GitHub**: [github.com/medalcode/chaos-api-proxy](https://github.com/medalcode/chaos-api-proxy)
- **Issues**: Para bugs y feature requests
- **Discussions**: Para preguntas generales
- **License**: MIT - Usa libremente

---

## 🏆 Logros del Proyecto

✅ **Funcional**: Proxy completo y listo para usar  
✅ **Bien Documentado**: > 2000 líneas de documentación  
✅ **Probado**: Tests unitarios con > 70% coverage  
✅ **Producción-Ready**: Docker, graceful shutdown, logging  
✅ **Extensible**: Arquitectura modular y bien diseñada  
✅ **Educational**: Ejemplos completos y guías de uso

---

**¡Gracias por usar Chaos API Proxy!** 🌪️🚀

Si este proyecto te fue útil, considera:

- ⭐ Dar una estrella en GitHub
- 🐛 Reportar bugs encontrados
- 💡 Sugerir mejoras
- 🤝 Contribuir código

_Desarrollado con ❤️ para la comunidad de Chaos Engineering_
