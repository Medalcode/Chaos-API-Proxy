# 🚀 Guía de Instalación y Despliegue

## Requisitos Previos

- **Go 1.21+** (para desarrollo local)
- **Docker & Docker Compose** (para despliegue con contenedores)
- **Redis** (si no usas Docker)
- **jq** (opcional, para ejemplos con JSON)

## Opción 1: Docker Compose (Recomendado)

### Instalación de Docker

#### Ubuntu/Debian

```bash
# Actualizar repositorios
sudo apt-get update

# Instalar Docker
sudo apt-get install -y docker.io docker-compose-plugin

# Añadir usuario al grupo docker
sudo usermod -aG docker $USER
newgrp docker

# Verificar instalación
docker --version
docker compose version
```

#### Fedora/RHEL

```bash
sudo dnf install -y docker docker-compose-plugin
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
```

### Iniciar el Proxy

```bash
# 1. Clonar el repositorio
git clone https://github.com/medalcode/chaos-api-proxy.git
cd chaos-api-proxy

# 2. Iniciar servicios
docker compose up -d

# 3. Verificar que está corriendo
docker compose ps

# 4. Ver logs
docker compose logs -f chaos-proxy

# 5. Probar health check
curl http://localhost:8080/health
```

### Comandos Útiles

```bash
# Ver logs en tiempo real
docker compose logs -f

# Reiniciar servicios
docker compose restart

# Detener servicios
docker compose down

# Reconstruir imagen (después de cambios)
docker compose build
docker compose up -d

# Ver estadísticas de recursos
docker stats
```

## Opción 2: Desarrollo Local (Sin Docker)

### Instalación de Go

#### Ubuntu/Debian

```bash
# Descargar Go 1.21
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz

# Extraer
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Añadir al PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verificar
go version
```

### Instalación de Redis

#### Ubuntu/Debian

```bash
sudo apt-get install -y redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Verificar
redis-cli ping
# Debería devolver: PONG
```

### Ejecutar el Proxy

```bash
# 1. Instalar dependencias
make deps

# 2. Ejecutar en modo desarrollo
make run

# O manualmente:
go run ./cmd/server/main.go
```

## Opción 3: Build Manual

```bash
# Compilar binario
make build

# El binario estará en bin/chaos-api-proxy
./bin/chaos-api-proxy
```

## Configuración

### Variables de Entorno

Crea un archivo `.env`:

```bash
cp .env.example .env
```

Edita `.env`:

```env
PORT=8080
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Para Docker Compose

Edita `docker-compose.yml`:

```yaml
chaos-proxy:
  environment:
    - PORT=8080
    - REDIS_ADDR=redis:6379
    - REDIS_PASSWORD=mi_password_seguro
```

## Verificación de Instalación

### 1. Health Check

```bash
curl http://localhost:8080/health
# Esperado: {"status":"healthy"}
```

### 2. Crear una Configuración de Prueba

```bash
curl -X POST http://localhost:8080/api/v1/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Config",
    "target": "https://jsonplaceholder.typicode.com",
    "enabled": true,
    "rules": {"latency_ms": 100}
  }'
```

### 3. Ejecutar Demo Completo

```bash
./examples/demo.sh
```

## Despliegue en Producción

### ⚠️ IMPORTANTE: Consideraciones de Seguridad

**Este proxy NO debe usarse en producción sin medidas de seguridad adicionales:**

1. **Autenticación**: Implementar API keys o JWT
2. **Rate Limiting**: Evitar abuso
3. **TLS/HTTPS**: Encriptar tráfico
4. **Logging**: Registrar todas las peticiones
5. **Network Policies**: Restringir acceso

### Ejemplo con NGINX + TLS

```nginx
server {
    listen 443 ssl http2;
    server_name chaos-proxy.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chaos-proxy
spec:
  replicas: 2
  selector:
    matchLabels:
      app: chaos-proxy
  template:
    metadata:
      labels:
        app: chaos-proxy
    spec:
      containers:
        - name: chaos-proxy
          image: chaos-api-proxy:latest
          ports:
            - containerPort: 8080
          env:
            - name: REDIS_ADDR
              value: "redis-service:6379"
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: chaos-proxy-service
spec:
  selector:
    app: chaos-proxy
  ports:
    - port: 80
      targetPort: 8080
  type: LoadBalancer
```

## Monitorización

### Logs Estructurados (JSON)

El proxy usa logrus con formato JSON:

```bash
# Ver solo errores
docker compose logs chaos-proxy | grep '"level":"error"'

# Ver peticiones proxy
docker compose logs chaos-proxy | grep 'Processing proxy request'
```

### Métricas con Prometheus (TODO)

En el futuro, se pueden añadir métricas Prometheus:

- Total de peticiones
- Tasa de errores inyectados
- Latencia promedio
- Configuraciones activas

## Troubleshooting

### El servidor no inicia

```bash
# Verificar que el puerto no esté en uso
sudo lsof -i :8080

# Verificar logs
docker compose logs chaos-proxy
```

### No se puede conectar a Redis

```bash
# Verificar que Redis esté corriendo
docker compose ps redis

# Probar conexión manualmente
docker exec -it chaos-proxy-redis redis-cli ping
```

### Errores de proxy

```bash
# Ver logs detallados
docker compose logs -f chaos-proxy

# Verificar que la URL target sea accesible
curl https://api.stripe.com  # Debe funcionar
```

## Desarrollo

### Tests

```bash
# Ejecutar todos los tests
make test

# Con coverage
make test-coverage

# Abrir reporte HTML
xdg-open coverage.html
```

### Linting

```bash
# Instalar golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Ejecutar linter
make lint
```

### Hot Reload

```bash
# Instalar air
go install github.com/cosmtrek/air@latest

# Ejecutar con hot reload
make dev
```

## Próximos Pasos

1. ✅ **Probar con APIs reales**: Configura el proxy para tus APIs favoritas
2. ✅ **Experimentar con parámetros**: Prueba diferentes combinaciones de caos
3. ✅ **Medir la resiliencia**: Observa cómo reacciona tu app
4. ✅ **Iterar**: Mejora el manejo de errores en tu aplicación

## Recursos Adicionales

- [Documentación de Go httputil.ReverseProxy](https://pkg.go.dev/net/http/httputil#ReverseProxy)
- [Redis Documentation](https://redis.io/documentation)
- [Chaos Engineering Principles](https://principlesofchaos.org/)
- [Site Reliability Engineering Book](https://sre.google/books/)

---

¿Problemas? [Abre un issue](https://github.com/medalcode/chaos-api-proxy/issues) 🚀
