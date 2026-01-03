#!/bin/bash

# 🚀 Chaos API Proxy - Quick Start Script
# Este script te ayuda a empezar rápidamente con el Chaos API Proxy

set -e  # Exit on error

echo "🌪️  Chaos API Proxy - Quick Start"
echo "=================================="
echo ""

# Verificar si Docker está instalado
if ! command -v docker &> /dev/null; then
    echo "❌ Docker no está instalado."
    echo ""
    echo "Por favor, instala Docker primero:"
    echo "  Ubuntu/Debian: sudo apt-get install docker.io docker-compose-plugin"
    echo "  Fedora: sudo dnf install docker docker-compose-plugin"
    echo ""
    echo "Luego ejecuta: sudo systemctl start docker"
    exit 1
fi

echo "✅ Docker detectado"

# Verificar si Docker Compose está disponible
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    echo "❌ Docker Compose no está instalado."
    echo ""
    echo "Por favor, instala Docker Compose:"
    echo "  sudo apt-get install docker-compose-plugin"
    exit 1
fi

echo "✅ Docker Compose detectado"
echo ""

# Preguntar al usuario qué quiere hacer
echo "¿Qué quieres hacer?"
echo "1) Iniciar el Chaos API Proxy con Docker"
echo "2) Ejecutar el script de demostración"
echo "3) Ver logs del proxy"
echo "4) Detener el proxy"
echo "5) Ver ayuda"
echo ""
read -p "Selecciona una opción (1-5): " option

case $option in
    1)
        echo ""
        echo "🚀 Iniciando Chaos API Proxy..."
        echo ""
        
        # Detener contenedores existentes si los hay
        $DOCKER_COMPOSE down 2>/dev/null || true
        
        # Iniciar servicios
        $DOCKER_COMPOSE up -d
        
        echo ""
        echo "⏳ Esperando a que el proxy esté listo..."
        sleep 5
        
        # Verificar health
        if curl -s http://localhost:8081/health > /dev/null 2>&1; then
            echo "✅ Proxy iniciado correctamente!"
            echo ""
            echo "📍 El proxy está corriendo en: http://localhost:8081"
            echo ""
            echo "Próximos pasos:"
            echo "  1. Ver documentación: cat README.md"
            echo "  2. Ejecutar demo: ./quick-start.sh (opción 2)"
            echo "  3. Crear tu primera configuración:"
            echo ""
            echo "     curl -X POST http://localhost:8081/api/v1/configs \\"
            echo "       -H \"Content-Type: application/json\" \\"
            echo "       -d '{"
            echo "         \"name\": \"Mi Primera Config\","
            echo "         \"target\": \"https://jsonplaceholder.typicode.com\","
            echo "         \"enabled\": true,"
            echo "         \"rules\": {\"latency_ms\": 500}"
            echo "       }'"
            echo ""
        else
            echo "⚠️  El proxy no respondió al health check."
            echo "Ver logs con: $DOCKER_COMPOSE logs chaos-proxy"
        fi
        ;;
        
    2)
        echo ""
        echo "🎬 Ejecutando script de demostración..."
        echo ""
        
        # Verificar que el proxy esté corriendo
        if ! curl -s http://localhost:8081/health > /dev/null 2>&1; then
            echo "⚠️  El proxy no está corriendo."
            echo "Iniciándolo primero..."
            echo ""
            $DOCKER_COMPOSE up -d
            sleep 5
        fi
        
        # Ejecutar demo
        if [ -f examples/demo.sh ]; then
            chmod +x examples/demo.sh
            ./examples/demo.sh
        else
            echo "❌ Script de demo no encontrado: examples/demo.sh"
        fi
        ;;
        
    3)
        echo ""
        echo "📋 Logs del Chaos API Proxy:"
        echo ""
        $DOCKER_COMPOSE logs -f chaos-proxy
        ;;
        
    4)
        echo ""
        echo "🛑 Deteniendo Chaos API Proxy..."
        $DOCKER_COMPOSE down
        echo "✅ Proxy detenido"
        ;;
        
    5)
        echo ""
        echo "📚 Ayuda - Chaos API Proxy"
        echo "=========================="
        echo ""
        echo "Comandos útiles:"
        echo ""
        echo "# Iniciar servicios"
        echo "  $DOCKER_COMPOSE up -d"
        echo ""
        echo "# Ver logs"
        echo "  $DOCKER_COMPOSE logs -f chaos-proxy"
        echo ""
        echo "# Detener servicios"
        echo "  $DOCKER_COMPOSE down"
        echo ""
        echo "# Health check"
        echo "  curl http://localhost:8081/health"
        echo ""
        echo "# Listar configuraciones"
        echo "  curl http://localhost:8081/api/v1/configs"
        echo ""
        echo "# Crear configuración"
        echo "  curl -X POST http://localhost:8081/api/v1/configs \\"
        echo "    -H \"Content-Type: application/json\" \\"
        echo "    -d '{\"name\":\"Test\",\"target\":\"https://api.example.com\",\"rules\":{}}'"
        echo ""
        echo "Documentación completa:"
        echo "  - README.md"
        echo "  - docs/INSTALLATION.md"
        echo "  - docs/EXAMPLES.md"
        echo "  - docs/API.md"
        echo ""
        ;;
        
    *)
        echo ""
        echo "❌ Opción inválida. Por favor selecciona 1-5."
        exit 1
        ;;
esac

echo ""
