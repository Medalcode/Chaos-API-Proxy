# 🤝 Contributing to Chaos API Proxy

¡Gracias por tu interés en contribuir! Este documento te guiará en el proceso.

## 📋 Tabla de Contenidos

- [Código de Conducta](#código-de-conducta)
- [¿Cómo Puedo Contribuir?](#cómo-puedo-contribuir)
- [Configuración del Entorno](#configuración-del-entorno)
- [Proceso de Desarrollo](#proceso-de-desarrollo)
- [Guía de Estilo](#guía-de-estilo)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)

## Código de Conducta

Este proyecto se adhiere a un código de conducta. Al participar, se espera que mantengas un comportamiento respetuoso y profesional.

## ¿Cómo Puedo Contribuir?

### 🐛 Reportar Bugs

Antes de crear un bug report, verifica que no exista ya. Si creas uno nuevo, incluye:

- **Descripción clara** del problema
- **Pasos para reproducir** el comportamiento
- **Comportamiento esperado** vs **comportamiento actual**
- **Screenshots** si aplica
- **Environment**: OS, Go version, Docker version

**Ejemplo de Bug Report:**

```markdown
### Descripción

El proxy no aplica latencia cuando jitter es mayor que latency_ms

### Pasos para Reproducir

1. Crear config con `latency_ms: 100, jitter: 200`
2. Hacer petición a través del proxy
3. Observar que a veces la latencia es negativa

### Comportamiento Esperado

La latencia nunca debería ser negativa

### Comportamiento Actual

Se obtienen latencias negativas que causan errores

### Environment

- OS: Ubuntu 22.04
- Go: 1.21.6
- Docker: 24.0.7
```

### 💡 Sugerir Features

Las sugerencias de features son bienvenidas! Incluye:

- **Caso de uso** detallado
- **Por qué sería útil** para otros usuarios
- **Ejemplos** de cómo se usaría

### 🔧 Pull Requests

Contribuciones de código son muy apreciadas! Ver [Pull Request Process](#pull-request-process).

## Configuración del Entorno

### 1. Fork y Clone

```bash
# Fork en GitHub, luego:
git clone https://github.com/TU_USUARIO/chaos-api-proxy.git
cd chaos-api-proxy

# Añadir upstream
git remote add upstream https://github.com/medalcode/chaos-api-proxy.git
```

### 2. Instalar Dependencias

```bash
# Go 1.21+
go version

# Instalar dependencias
make deps

# Redis para desarrollo
docker run -d -p 6379:6379 redis:7-alpine
```

### 3. Ejecutar Tests

```bash
# Todos los tests
make test

# Con coverage
make test-coverage
```

### 4. Ejecutar el Proxy

```bash
# Modo desarrollo
make run

# O con hot reload
make dev
```

## Proceso de Desarrollo

### 1. Crear una Rama

```bash
git checkout -b feature/nombre-descriptivo

# O para bugs
git checkout -b fix/nombre-del-bug
```

**Nombres de ramas:**

- `feature/` - Nuevas funcionalidades
- `fix/` - Bug fixes
- `docs/` - Cambios en documentación
- `refactor/` - Refactorización
- `test/` - Añadir tests

### 2. Hacer Cambios

Sigue la [Guía de Estilo](#guía-de-estilo).

### 3. Commit

Mensajes de commit claros y descriptivos:

```bash
git commit -m "feat: add circuit breaker pattern support"
git commit -m "fix: prevent negative latency values"
git commit -m "docs: update API reference for bandwidth limiting"
```

**Formato de commits:**

```
type(scope): subject

body (optional)

footer (optional)
```

**Types:**

- `feat`: Nueva funcionalidad
- `fix`: Bug fix
- `docs`: Documentación
- `style`: Formato, punto y coma, etc
- `refactor`: Refactorización
- `test`: Tests
- `chore`: Mantenimiento

**Ejemplos:**

```
feat(chaos): add response truncation chaos type

Implements new chaos type that truncates response bodies
at a random position to simulate network issues.

Closes #42
```

```
fix(proxy): handle nil response body gracefully

Previously, nil response bodies caused panic. Now they are
handled correctly with proper error messages.

Fixes #38
```

### 4. Push

```bash
git push origin feature/nombre-descriptivo
```

### 5. Abrir Pull Request

- Título descriptivo
- Descripción detallada de cambios
- Referenciar issues relacionados
- Screenshots si aplica

## Guía de Estilo

### Go Code Style

Seguimos [Effective Go](https://go.dev/doc/effective_go) y [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Principales reglas:**

1. **Nombres**: CamelCase o camelCase (no snake_case)
2. **Comentarios**: Toda función exportada debe tener comentario
3. **Error handling**: Siempre manejar errores explícitamente
4. **Formato**: Usar `gofmt` (automático con `make lint`)

**Ejemplo de buen código:**

```go
// ProcessRequest handles incoming HTTP requests and applies chaos rules.
// It returns an error if the configuration cannot be retrieved from storage.
func ProcessRequest(ctx context.Context, configID string) error {
    config, err := storage.GetConfig(ctx, configID)
    if err != nil {
        log.WithError(err).Error("Failed to get config")
        return fmt.Errorf("get config: %w", err)
    }

    if !config.Enabled {
        return ErrConfigDisabled
    }

    return nil
}
```

### Logging

Usar `logrus` con campos estructurados:

```go
log.WithFields(log.Fields{
    "config_id": configID,
    "latency":   decision.LatencyDuration,
    "injected":  decision.ShouldInjectError,
}).Info("Processing request")
```

Niveles:

- `Debug`: Información detallada para debugging
- `Info`: Eventos normales importantes
- `Warn`: Situaciones inusuales pero manejables
- `Error`: Errores que necesitan atención

### Testing

1. **Coverage**: Mantener > 70% de coverage
2. **Table-driven tests**: Preferir este patrón
3. **Nombres descriptivos**: `TestEngineMakeDecision_WithHighLatency`

**Ejemplo de table-driven test:**

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  ChaosConfig
        wantErr bool
        errType error
    }{
        {
            name:    "valid config",
            config:  ChaosConfig{Target: "https://api.example.com"},
            wantErr: false,
        },
        {
            name:    "empty target",
            config:  ChaosConfig{Target: ""},
            wantErr: true,
            errType: ErrInvalidTarget,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Testing

### Unit Tests

```bash
# Ejecutar todos los tests
make test

# Un paquete específico
go test ./internal/chaos/

# Un test específico
go test -run TestEngineMakeDecision ./internal/chaos/

# Verbose
go test -v ./internal/...
```

### Coverage

```bash
# Generar coverage
make test-coverage

# Ver en browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Integration Tests

```bash
# TODO: Añadir tests de integración
# make test-integration
```

## Pull Request Process

### Checklist antes de PR

- [ ] Tests pasan: `make test`
- [ ] Linter pasa: `make lint`
- [ ] Coverage >= 70%
- [ ] Documentación actualizada si aplica
- [ ] Commits tienen mensajes descriptivos
- [ ] Branch actualizado con `main`

### Template de PR

```markdown
## Descripción

[Descripción clara de qué cambia este PR]

## Tipo de cambio

- [ ] Bug fix (non-breaking change)
- [ ] New feature (non-breaking change)
- [ ] Breaking change (fix o feature que causa cambios incompatibles)
- [ ] Documentation update

## ¿Cómo se ha probado?

[Describe las pruebas que ejecutaste]

## Checklist

- [ ] Mi código sigue el style guide del proyecto
- [ ] He revisado mi propio código
- [ ] He comentado mi código donde es necesario
- [ ] He actualizado la documentación
- [ ] Mis cambios no generan nuevas advertencias
- [ ] He añadido tests que prueban mi fix/feature
- [ ] Tests nuevos y existentes pasan localmente

## Screenshots (si aplica)

[Screenshots o GIFs]

## Issues relacionados

Closes #[issue number]
Related to #[issue number]
```

### Proceso de Review

1. **Abres PR** → GitHub Actions corre tests automáticamente
2. **Maintainer revisa** → Puede pedir cambios
3. **Haces cambios** → Push a la misma rama
4. **Aprobación** → Maintainer hace merge
5. **Cleanup** → Borra tu rama after merge

### Responder a Feedback

```bash
# Hacer cambios solicitados
git add .
git commit -m "address review comments"
git push origin feature/nombre-descriptivo
```

## Ideas de Contribución

### Para Empezar (Good First Issues)

- [ ] Añadir más ejemplos de configuración en `/examples`
- [ ] Mejorar documentación con más casos de uso
- [ ] Añadir badges al README (build status, coverage, etc)
- [ ] Crear logo para el proyecto

### Features Medios

- [ ] Implementar métricas Prometheus
- [ ] Añadir soporte para WebSockets
- [ ] UI web para gestionar configuraciones
- [ ] CLI tool para interactuar con el API

### Features Avanzados

- [ ] Sistema de plugins para chaos personalizados
- [ ] Distributed tracing con OpenTelemetry
- [ ] Kubernetes operator
- [ ] gRPC proxy support

## Recursos

### Go

- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Proverbs](https://go-proverbs.github.io/)

### Testing

- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Testing in Go](https://quii.gitbook.io/learn-go-with-tests/)

### Chaos Engineering

- [Principles of Chaos Engineering](https://principlesofchaos.org/)
- [Chaos Engineering Book](https://www.oreilly.com/library/view/chaos-engineering/9781491988459/)

## Preguntas

¿Tienes dudas?

- **GitHub Issues**: Para preguntas técnicas
- **GitHub Discussions**: Para ideas y conversaciones generales

---

¡Gracias por contribuir! 🎉
