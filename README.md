# 🌪️ Chaos API Proxy (Titanium/Node Edition)

> **Un Web Proxy diseñado para introducir caos, latencia y fallos en tus APIs.**  
> Ahora reescrito en **TypeScript (Node.js)** para máxima flexibilidad y programación dinámica.

---

## 🚀 Características Principales

- **🛡️ Intercepción Transparente**: Funciona como un proxy inverso entre tus clientes y tu API real.
- **⏱️ Inyección de Latencia**: Fija o con _jitter_ (variable).
- **💥 Inyección de Errores**: Retorna 500, 503, 404 a voluntad.
- **🧬 Response Fuzzing**: Muta JSONs para probar robustez de clientes.
- **📜 Dynamic Scripting (NUEVO)**: Escribe lógica JS personalizada para decidir cuándo y cómo aplicar caos.
- **📊 Métricas Prometheus**: Dashboards listos para consumir.
- **🚦 Live Logs**: Monitor de tráfico en tiempo real.
- **💻 Web Dashboard**: UI intuitiva para gestionar reglas y ver logs.

---

## 🛠️ Instalación y Uso

### Opción 1: Docker (Recomendado)

```bash
# 1. Clonar
git clone https://github.com/Medalcode/Chaos-API-Proxy.git
cd Chaos-API-Proxy

# 2. Arrancar (Redis + Proxy + Prometheus)
docker-compose up --build
```

El dashboard estará disponible en: [http://localhost:8081/dashboard](http://localhost:8081/dashboard)

### Opción 2: Local (Node.js)

Requisitos: Node.js 18+, Redis.

```bash
# Instalar dependencias
npm install

# Configurar entorno
cp .env.example .env # (Opcional, defaults funcionan)

# Arrancar en modo desarrollo
npm run dev
```

---

## 📜 Scripting Dinámico

Ahora puedes escribir scripts JavaScript para controlar el caos con precisión quirúrgica.

**Contexto disponible:**

- `req`: `{ method, path, headers, query, body }`
- `decision`: `{ shouldLatency, latencyMs, shouldError, errorCode, ... }`

**Ejemplo 1: Caos solo para iPhones**

```javascript
if (req.headers["user-agent"] && req.headers["user-agent"].includes("iPhone")) {
  decision.shouldLatency = true;
  decision.latencyMs = 2000;
}
```

**Ejemplo 2: Error 1 de cada 10 peticiones POST**

```javascript
if (req.method === "POST" && Math.random() < 0.1) {
  decision.shouldError = true;
  decision.errorCode = 503;
}
```

---

## 🔒 Seguridad

Configura `CHAOS_API_KEYS` en tu `.env` o `docker-compose.yml` para proteger el dashboard y la API de administración.

```env
CHAOS_API_KEYS=mi-clave-secreta-123
```

---

_Hecho con ❤️ y ☕ por el equipo de Chaos Engineering._
