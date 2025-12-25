# 📘 Ejemplos de Uso del Chaos API Proxy

## Índice

1. [Conceptos Básicos](#conceptos-básicos)
2. [Ejemplos por Caso de Uso](#ejemplos-por-caso-de-uso)
3. [Frontend Testing](#frontend-testing)
4. [Backend Testing](#backend-testing)
5. [Testing Automatizado](#testing-automatizado)
6. [Patrones Avanzados](#patrones-avanzados)

## Conceptos Básicos

### Flujo de Trabajo

```
1. Crear configuración → 2. Obtener Config ID → 3. Usar en tu app → 4. Observar comportamiento
```

### Anatomía de una Petición

```bash
# URL Original (sin proxy)
https://api.stripe.com/v1/charges

# URL Con Proxy
http://localhost:8081/proxy/{CONFIG_ID}/v1/charges
                              ↑
                        Tu Config ID
```

## Ejemplos por Caso de Uso

### 1. Testing de Estados de Carga (Loading States)

**Problema**: ¿Tu UI muestra correctamente un spinner cuando la API tarda?

```bash
# Crear configuración con latencia alta
curl -X POST http://localhost:8081/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Latency Test",
    "target": "https://api.github.com",
    "enabled": true,
    "rules": {
      "latency_ms": 3000,
      "jitter": 1000
    }
  }'

# Guardar el ID retornado
CONFIG_ID="abc-123-def"
```

**En tu aplicación React/Vue/Angular:**

```javascript
const fetchWithChaos = async () => {
  setLoading(true);

  try {
    // Apuntar al proxy en lugar de la API real
    const response = await fetch(
      `http://localhost:8081/proxy/${CONFIG_ID}/users`
    );

    // Tu app debería mostrar loading durante 2-4 segundos
    const data = await response.json();
    setData(data);
  } catch (error) {
    setError(error);
  } finally {
    setLoading(false);
  }
};
```

### 2. Testing de Manejo de Errores

**Problema**: ¿Tu app muestra mensajes de error apropiados cuando la API falla?

```bash
# Configuración que falla el 50% de las veces
curl -X POST http://localhost:8081/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Error Injection Test",
    "target": "https://api.example.com",
    "enabled": true,
    "rules": {
      "inject_failure_rate": 0.5,
      "error_code": 503,
      "error_body": "{\"error\": \"Service temporarily unavailable\"}"
    }
  }'
```

**Testing en tu app:**

```javascript
// Hacer 10 peticiones y verificar manejo de errores
for (let i = 0; i < 10; i++) {
  try {
    const res = await fetch(`http://localhost:8081/proxy/${CONFIG_ID}/data`);
    if (!res.ok) {
      console.log(`❌ Request ${i + 1}: Error ${res.status}`);
      // ¿Tu UI muestra el error correctamente?
    } else {
      console.log(`✅ Request ${i + 1}: Success`);
    }
  } catch (error) {
    console.log(`💥 Request ${i + 1}: Network error`);
    // ¿Manejas errores de red?
  }
}
```

### 3. Testing de Timeouts

**Problema**: ¿Tu app tiene timeouts configurados? ¿Funcionan?

```bash
# Configuración que cierra conexiones aleatoriamente
curl -X POST http://localhost:8081/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Connection Drop Test",
    "target": "https://api.example.com",
    "enabled": true,
    "rules": {
      "drop_connection_rate": 0.3,
      "latency_ms": 5000
    }
  }'
```

**Con Axios (timeout):**

```javascript
import axios from "axios";

const api = axios.create({
  baseURL: `http://localhost:8081/proxy/${CONFIG_ID}`,
  timeout: 3000, // 3 segundos
});

try {
  const response = await api.get("/data");
  console.log("✅ Success:", response.data);
} catch (error) {
  if (error.code === "ECONNABORTED") {
    console.log("⏱️ Request timed out - ¿Tu UI lo maneja?");
  } else if (error.response) {
    console.log("❌ Server error:", error.response.status);
  } else {
    console.log("💥 Network error:", error.message);
  }
}
```

### 4. Testing de Rate Limiting

**Problema**: ¿Tu app respeta headers de rate limiting?

```bash
curl -X POST http://localhost:8081/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Rate Limit Simulation",
    "target": "https://api.github.com",
    "enabled": true,
    "rules": {
      "inject_failure_rate": 0.2,
      "error_code": 429,
      "error_body": "{\"message\": \"API rate limit exceeded\", \"retry_after\": 60}",
      "modify_headers": {
        "X-RateLimit-Limit": "60",
        "X-RateLimit-Remaining": "0",
        "Retry-After": "60"
      }
    }
  }'
```

**Implementar retry con backoff:**

```javascript
async function fetchWithRetry(url, options = {}, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(url, options);

      if (response.status === 429) {
        const retryAfter = response.headers.get("Retry-After") || 60;
        console.log(`⏳ Rate limited, waiting ${retryAfter}s...`);
        await sleep(retryAfter * 1000);
        continue;
      }

      return response;
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await sleep(Math.pow(2, i) * 1000); // Exponential backoff
    }
  }
}
```

### 5. Testing de Conexiones Lentas (Mobile)

**Problema**: ¿Tu app funciona bien en 3G/4G?

```bash
# Simular conexión 3G (100 KB/s)
curl -X POST http://localhost:8081/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "3G Speed Simulation",
    "target": "https://cdn.example.com",
    "enabled": true,
    "rules": {
      "bandwidth_limit_kbps": 100,
      "latency_ms": 200,
      "jitter": 100
    }
  }'
```

**Testing de progressive loading:**

```javascript
async function downloadImage(url) {
  const response = await fetch(
    `http://localhost:8081/proxy/${CONFIG_ID}${url}`
  );
  const reader = response.body.getReader();
  const contentLength = response.headers.get("Content-Length");

  let receivedLength = 0;
  let chunks = [];

  while (true) {
    const { done, value } = await reader.read();

    if (done) break;

    chunks.push(value);
    receivedLength += value.length;

    // Actualizar barra de progreso
    const progress = (receivedLength / contentLength) * 100;
    console.log(`Downloading: ${progress.toFixed(2)}%`);
    updateProgressBar(progress);
  }

  return new Blob(chunks);
}
```

## Frontend Testing

### React Example

```javascript
import { useState, useEffect } from "react";

function UserList() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const CONFIG_ID = "your-config-id";

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await fetch(
          `http://localhost:8081/proxy/${CONFIG_ID}/users`
        );

        // Verificar header de chaos
        if (response.headers.get("X-Chaos-Proxy-Injected")) {
          console.warn("⚠️ Esta respuesta fue afectada por caos");
        }

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }

        const data = await response.json();
        setUsers(data);
      } catch (err) {
        setError(err.message);
        // ¿Tu componente muestra este error?
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  if (loading) return <Spinner />; // ¿Se muestra?
  if (error) return <ErrorMessage error={error} />; // ¿Se muestra?

  return (
    <ul>
      {users.map((u) => (
        <li key={u.id}>{u.name}</li>
      ))}
    </ul>
  );
}
```

### Vue Example

```vue
<template>
  <div>
    <div v-if="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <ul v-else>
      <li v-for="user in users" :key="user.id">{{ user.name }}</li>
    </ul>
  </div>
</template>

<script>
export default {
  data() {
    return {
      users: [],
      loading: false,
      error: null,
      configId: "your-config-id",
    };
  },
  async mounted() {
    this.loading = true;

    try {
      const response = await fetch(
        `http://localhost:8081/proxy/${this.configId}/users`
      );

      if (!response.ok) {
        throw new Error(`Error ${response.status}`);
      }

      this.users = await response.json();
    } catch (err) {
      this.error = err.message;
    } finally {
      this.loading = false;
    }
  },
};
</script>
```

## Backend Testing

### Node.js/Express

```javascript
const express = require("express");
const axios = require("axios");

const app = express();
const CONFIG_ID = "your-config-id";

// Circuit Breaker Pattern
class CircuitBreaker {
  constructor(threshold = 5, timeout = 60000) {
    this.failureCount = 0;
    this.threshold = threshold;
    this.timeout = timeout;
    this.state = "CLOSED"; // CLOSED, OPEN, HALF_OPEN
    this.nextAttempt = Date.now();
  }

  async execute(fn) {
    if (this.state === "OPEN") {
      if (Date.now() < this.nextAttempt) {
        throw new Error("Circuit breaker is OPEN");
      }
      this.state = "HALF_OPEN";
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  onSuccess() {
    this.failureCount = 0;
    this.state = "CLOSED";
  }

  onFailure() {
    this.failureCount++;
    if (this.failureCount >= this.threshold) {
      this.state = "OPEN";
      this.nextAttempt = Date.now() + this.timeout;
      console.log("🔴 Circuit breaker opened!");
    }
  }
}

const breaker = new CircuitBreaker();

app.get("/api/users", async (req, res) => {
  try {
    const data = await breaker.execute(async () => {
      const response = await axios.get(
        `http://localhost:8081/proxy/${CONFIG_ID}/users`,
        { timeout: 5000 }
      );
      return response.data;
    });

    res.json(data);
  } catch (error) {
    if (error.message === "Circuit breaker is OPEN") {
      res.status(503).json({ error: "Service unavailable" });
    } else {
      res.status(500).json({ error: "Internal server error" });
    }
  }
});

app.listen(3000);
```

### Python/Flask

```python
import requests
from functools import wraps
import time

CONFIG_ID = "your-config-id"
BASE_URL = f"http://localhost:8081/proxy/{CONFIG_ID}"

# Retry decorator con exponential backoff
def retry_with_backoff(retries=3, backoff_in_seconds=1):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            x = 0
            while True:
                try:
                    return func(*args, **kwargs)
                except requests.exceptions.RequestException as e:
                    if x == retries:
                        raise

                    wait = backoff_in_seconds * 2 ** x
                    print(f"Retry {x+1}/{retries} after {wait}s...")
                    time.sleep(wait)
                    x += 1
        return wrapper
    return decorator

@retry_with_backoff(retries=3)
def fetch_users():
    response = requests.get(
        f"{BASE_URL}/users",
        timeout=5
    )
    response.raise_for_status()
    return response.json()

# Usar en tu app
try:
    users = fetch_users()
    print(f"✅ Got {len(users)} users")
except Exception as e:
    print(f"❌ Failed after retries: {e}")
```

## Testing Automatizado

### Jest/Vitest Tests

```javascript
import { describe, it, expect, beforeAll } from "vitest";

describe("API Resilience Tests", () => {
  let configId;

  beforeAll(async () => {
    // Crear configuración para tests
    const response = await fetch("http://localhost:8081/api/v1/configs", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: "Test Config",
        target: "https://jsonplaceholder.typicode.com",
        enabled: true,
        rules: {
          inject_failure_rate: 0.5,
          error_code: 500,
        },
      }),
    });

    const config = await response.json();
    configId = config.id;
  });

  it("should handle errors gracefully", async () => {
    let successCount = 0;
    let errorCount = 0;
    const iterations = 20;

    for (let i = 0; i < iterations; i++) {
      try {
        const response = await fetch(
          `http://localhost:8081/proxy/${configId}/posts/1`
        );

        if (response.ok) {
          successCount++;
        } else {
          errorCount++;
        }
      } catch (error) {
        errorCount++;
      }
    }

    // Con 50% failure rate, esperamos ~10 errores
    expect(errorCount).toBeGreaterThan(5);
    expect(errorCount).toBeLessThan(15);

    console.log(`Results: ${successCount} success, ${errorCount} errors`);
  });
});
```

### Cypress E2E Tests

```javascript
describe("Chaos Testing", () => {
  it("shows loading spinner during slow requests", () => {
    // Visitar tu app
    cy.visit("/users");

    // Verificar que aparece loading
    cy.get('[data-testid="loading-spinner"]').should("be.visible");

    // Esperar a que termine (con latencia inyectada)
    cy.get('[data-testid="user-list"]', { timeout: 10000 }).should(
      "be.visible"
    );
  });

  it("shows error message on API failure", () => {
    // Interceptar y forzar error
    cy.intercept("GET", "/api/users", {
      statusCode: 503,
      body: { error: "Service unavailable" },
    });

    cy.visit("/users");

    cy.get('[data-testid="error-message"]')
      .should("be.visible")
      .and("contain", "unavailable");
  });
});
```

## Patrones Avanzados

### Configuraciones Dinámicas

```javascript
// Cambiar configuración on-the-fly para diferentes tests
async function createTestConfig(rules) {
  const response = await fetch("http://localhost:8081/api/v1/configs", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      name: `Test-${Date.now()}`,
      target: "https://api.example.com",
      enabled: true,
      rules,
    }),
  });

  const config = await response.json();
  return config.id;
}

// Test con latencia baja
const lowLatencyConfig = await createTestConfig({ latency_ms: 100 });
await runTests(lowLatencyConfig);

// Test con latencia alta
const highLatencyConfig = await createTestConfig({ latency_ms: 3000 });
await runTests(highLatencyConfig);

// Test con errores
const errorConfig = await createTestConfig({ inject_failure_rate: 0.8 });
await runTests(errorConfig);
```

### Métricas y Observabilidad

```javascript
class ChaosMetrics {
  constructor() {
    this.totalRequests = 0;
    this.chaosInjected = 0;
    this.errors = 0;
    this.latencies = [];
  }

  async fetch(url) {
    this.totalRequests++;
    const startTime = Date.now();

    try {
      const response = await fetch(url);
      const latency = Date.now() - startTime;
      this.latencies.push(latency);

      if (response.headers.get("X-Chaos-Proxy-Injected")) {
        this.chaosInjected++;
      }

      if (!response.ok) {
        this.errors++;
      }

      return response;
    } catch (error) {
      this.errors++;
      throw error;
    }
  }

  getStats() {
    const avgLatency =
      this.latencies.reduce((a, b) => a + b, 0) / this.latencies.length;

    return {
      totalRequests: this.totalRequests,
      chaosInjected: this.chaosInjected,
      chaosRate:
        ((this.chaosInjected / this.totalRequests) * 100).toFixed(2) + "%",
      errorRate: ((this.errors / this.totalRequests) * 100).toFixed(2) + "%",
      avgLatency: avgLatency.toFixed(2) + "ms",
    };
  }
}

// Usar
const metrics = new ChaosMetrics();

for (let i = 0; i < 100; i++) {
  await metrics.fetch(`http://localhost:8081/proxy/${CONFIG_ID}/data`);
}

console.log(metrics.getStats());
// {
//   totalRequests: 100,
//   chaosInjected: 12,
//   chaosRate: '12.00%',
//   errorRate: '10.00%',
//   avgLatency: '523.45ms'
// }
```

---

¿Más ejemplos? Contribuye en [GitHub](https://github.com/medalcode/chaos-api-proxy) 🚀
