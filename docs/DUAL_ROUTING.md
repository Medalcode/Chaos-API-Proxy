# 🔀 Dual Routing Mode - Usage Guide

El Chaos API Proxy ahora soporta **dos métodos** de identificación de configuración:

## 1️⃣ Path-Based Routing (Recomendado)

### Uso

```bash
# Formato: /proxy/{configID}/path/to/endpoint
curl http://localhost:8081/proxy/abc-123-def/v1/users
```

### Ventajas

✅ **URL auto-documenta** qué configuración se usa  
✅ **Fácil debugging** en logs  
✅ **RESTful** y estándar  
✅ **No requiere headers especiales**

### Ejemplo con JavaScript

```javascript
const CONFIG_ID = "abc-123-def";
const BASE_URL = `http://localhost:8081/proxy/${CONFIG_ID}`;

// Solo cambia la base URL, nada más
fetch(`${BASE_URL}/v1/users`)
  .then((res) => res.json())
  .then((data) => console.log(data));
```

---

## 2️⃣ Header-Based Routing (Compatible con Spec Original)

### Uso

```bash
# Envía header X-Chaos-Config-ID
curl http://localhost:8081/v1/users \
  -H "X-Chaos-Config-ID: abc-123-def"
```

### Ventajas

✅ **URL limpia** sin prefijo /proxy  
✅ **Compatible** con spec original  
✅ **Flexible** para proxies transparentes

### Ejemplo con JavaScript

```javascript
const CONFIG_ID = "abc-123-def";

fetch("http://localhost:8081/v1/users", {
  headers: {
    "X-Chaos-Config-ID": CONFIG_ID,
  },
})
  .then((res) => res.json())
  .then((data) => console.log(data));
```

### Ejemplo con Axios

```javascript
import axios from "axios";

const api = axios.create({
  baseURL: "http://localhost:8081",
  headers: {
    "X-Chaos-Config-ID": "abc-123-def",
  },
});

// Todas las peticiones usan el mismo config ID
api.get("/v1/users");
api.post("/v1/charges", { amount: 1000 });
```

---

## 🔄 Endpoints Alias `/rules`

Ambos endpoints funcionan idénticamente:

### API Versionada (Recomendado)

```bash
POST   /api/v1/configs
GET    /api/v1/configs
GET    /api/v1/configs/{id}
PUT    /api/v1/configs/{id}
DELETE /api/v1/configs/{id}
```

### Alias Compatibilidad

```bash
POST   /rules
GET    /rules
GET    /rules/{id}
PUT    /rules/{id}
DELETE /rules/{id}
```

### Ejemplo

```bash
# Ambos son equivalentes:
curl http://localhost:8081/api/v1/configs
curl http://localhost:8081/rules
```

---

## 🎯 ¿Cuál Usar?

### Usar **Path-Based** si:

- ✅ Quieres debugging fácil
- ✅ Estás en desarrollo/testing
- ✅ Quieres URLs auto-documentadas

### Usar **Header-Based** si:

- ✅ Necesitas URLs limpias
- ✅ Tienes un proxy transparente
- ✅ Quieres ocultar la configuración en la URL

---

## 📝 Ejemplo Completo: Ambos Modos

### Setup

```bash
# 1. Crear configuración
CONFIG_ID=$(curl -s -X POST http://localhost:8081/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Config",
    "target": "https://jsonplaceholder.typicode.com",
    "enabled": true,
    "rules": {
      "latency_ms": 500,
      "inject_failure_rate": 0.1
    }
  }' | jq -r '.id')

echo "Config ID: $CONFIG_ID"
```

### Opción A: Path-Based

```bash
# Hacer 10 requests con path-based
for i in {1..10}; do
  curl -s http://localhost:8081/proxy/$CONFIG_ID/posts/1 | head -3
  echo "---"
done
```

### Opción B: Header-Based

```bash
# Hacer 10 requests con header-based
for i in {1..10}; do
  curl -s http://localhost:8081/posts/1 \
    -H "X-Chaos-Config-ID: $CONFIG_ID" | head -3
  echo "---"
done
```

---

## 🔍 Verificar Modo en Logs

El proxy registra qué método se usó:

```json
{
  "level": "info",
  "msg": "Processing proxy request",
  "config_id": "abc-123-def",
  "target": "https://api.example.com",
  "routing_mode": "header-based" // o "path-based"
}
```

---

## ⚠️ Errores Comunes

### Error: "Missing configuration ID"

```bash
# ❌ Mal: No configID en path ni header
curl http://localhost:8081/v1/users

# ✅ Bien: Usar uno de los dos métodos
curl http://localhost:8081/proxy/abc-123/v1/users
# O
curl http://localhost:8081/v1/users -H "X-Chaos-Config-ID: abc-123"
```

### Error: "Configuration not found"

```bash
# ❌ ConfigID inválido
curl http://localhost:8081/proxy/invalid-id/v1/users

# ✅ Verificar configs disponibles
curl http://localhost:8081/rules
```

---

## 🎓 Casos de Uso

### Testing A/B de Resiliencia

```javascript
// Probar dos configuraciones diferentes
const configs = {
  "low-latency": "config-id-1",
  "high-latency": "config-id-2",
};

async function testResilience(scenario) {
  const configId = configs[scenario];

  const response = await fetch("http://localhost:8081/api/data", {
    headers: {
      "X-Chaos-Config-ID": configId,
    },
  });

  console.log(`Scenario ${scenario}:`, response.status);
}

await testResilience("low-latency");
await testResilience("high-latency");
```

### Cambiar Configuración Dinámicamente

```javascript
class ChaosClient {
  constructor(baseURL) {
    this.baseURL = baseURL;
    this.configId = null;
  }

  setChaosConfig(configId) {
    this.configId = configId;
  }

  async fetch(path, options = {}) {
    // Soporta ambos modos
    if (this.configId) {
      // Opción 1: Header-based
      options.headers = {
        ...options.headers,
        "X-Chaos-Config-ID": this.configId,
      };
      return fetch(`${this.baseURL}${path}`, options);

      // O Opción 2: Path-based
      // return fetch(`${this.baseURL}/proxy/${this.configId}${path}`, options);
    }

    // Sin caos
    return fetch(`${this.baseURL}${path}`, options);
  }
}

// Uso
const client = new ChaosClient("http://localhost:8081");

// Sin caos
await client.fetch("/api/users");

// Con caos
client.setChaosConfig("abc-123-def");
await client.fetch("/api/users"); // Ahora tiene caos inyectado
```

---

**¡Ahora tienes la flexibilidad de usar ambos métodos según tus necesidades!** 🚀
