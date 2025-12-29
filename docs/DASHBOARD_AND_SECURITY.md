# 🖥️ Web Dashboard & Seguridad

En la versión 1.2 del Chaos API Proxy hemos añadido una interfaz gráfica moderna y controles de seguridad.

---

## 🌪️ Web Dashboard

Una interfaz gráfica de una sola página (SPA) para gestionar las reglas de caos de forma visual.

**Acceso:** `http://localhost:8081/dashboard`

### Características

- **Listado Visual:** Ver todas las configuraciones activas y pausadas.
- **Control Rápido:** Activar/Pausar reglas con un click.
- **Creación Fácil:** Formulario para crear nuevas reglas sin lidiar con JSON manualmente.
- **Copia Rápida:** Click en el ID para copiarlo al portapapeles.
- **Modo Oscuro:** Diseño moderno "glassmorphism".

---

## 🔐 Seguridad (API Keys)

Protege tu Chaos Proxy para que solo usuarios autorizados puedan crear o borrar reglas.

### Configuración

Define la variable de entorno `CHAOS_API_KEYS` con una o más claves separadas por comas.

**En `docker-compose.yml`:**

```yaml
environment:
  - CHAOS_API_KEYS=secret-key-123,dev-team-key
```

**Si no defines esta variable, la autenticación estará DESACTIVADA (modo inseguro de desarrollo).**

### Uso con cURL

Debes incluir el header `X-API-Key` en tus peticiones a la API de administración (`/api/v1` o `/rules`).

```bash
curl -X POST http://localhost:8081/rules \
  -H "X-API-Key: secret-key-123" \
  ...
```

### Uso en el Dashboard

Si la autenticación está activada, verás un error al cargar el dashboard.
Introduce tu API Key en el campo **"🔑 API Key"** en la esquina superior derecha. El dashboard guardará la clave en tu navegador.

### Rutas Protegidas vs. Públicas

| Ruta         | Estado       | Descripción                                       |
| ------------ | ------------ | ------------------------------------------------- |
| `/api/v1/*`  | 🔒 Protegido | Gestión de reglas (CRUD)                          |
| `/rules`     | 🔒 Protegido | Alias de gestión de reglas                        |
| `/proxy/*`   | 🔓 Público   | Tráfico proxy (intencionalmente abierto)          |
| `/dashboard` | 🔓 Público\* | La UI carga, pero requiere Key para obtener datos |
| `/health`    | 🔓 Público   | Health check                                      |
| `/metrics`   | 🔓 Público   | Prometheus metrics                                |

---
