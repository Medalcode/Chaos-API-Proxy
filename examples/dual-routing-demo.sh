#!/bin/bash

# 🔀 Dual Routing Mode Demo
# Demuestra ambos métodos de routing: Path-based y Header-based

BASE_URL="http://localhost:8081"

echo "🔀 Dual Routing Mode - Demo Comparativo"
echo "=========================================="
echo ""

# Check if server is running
echo "1️⃣  Verificando servidor..."
if ! curl -s "${BASE_URL}/health" > /dev/null; then
    echo "❌ Servidor no está corriendo en ${BASE_URL}"
    echo "   Ejecuta: docker compose up -d"
    exit 1
fi
echo "✅ Servidor activo"
echo ""

# Create configuration
echo "2️⃣  Creando configuración de prueba..."
CONFIG_RESPONSE=$(curl -s -X POST "${BASE_URL}/rules" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dual Routing Demo",
    "target": "https://jsonplaceholder.typicode.com",
    "enabled": true,
    "rules": {
      "latency_ms": 300,
      "jitter": 100
    }
  }')

CONFIG_ID=$(echo "$CONFIG_RESPONSE" | jq -r '.id')
echo "✅ Config ID: $CONFIG_ID"
echo ""

echo "═══════════════════════════════════════════"
echo "Método 1: PATH-BASED ROUTING"
echo "═══════════════════════════════════════════"
echo ""
echo "URL: ${BASE_URL}/proxy/${CONFIG_ID}/posts/1"
echo ""

for i in {1..5}; do
    START=$(date +%s%N)
    
    RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/proxy/${CONFIG_ID}/posts/1")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    END=$(date +%s%N)
    DURATION=$(( ($END - $START) / 1000000 ))
    
    echo "Request #$i: HTTP $HTTP_CODE - ${DURATION}ms"
done

echo ""
echo "═══════════════════════════════════════════"
echo "Método 2: HEADER-BASED ROUTING"
echo "═══════════════════════════════════════════"
echo ""
echo "URL: ${BASE_URL}/posts/1"
echo "Header: X-Chaos-Config-ID: ${CONFIG_ID}"
echo ""

for i in {1..5}; do
    START=$(date +%s%N)
    
    RESPONSE=$(curl -s -w "\n%{http_code}" \
      -H "X-Chaos-Config-ID: ${CONFIG_ID}" \
      "${BASE_URL}/posts/1")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    END=$(date +%s%N)
    DURATION=$(( ($END - $START) / 1000000 ))
    
    echo "Request #$i: HTTP $HTTP_CODE - ${DURATION}ms"
done

echo ""
echo "═══════════════════════════════════════════"
echo "Comparación de Métodos"
echo "═══════════════════════════════════════════"
echo ""
echo "📍 Path-Based:"
echo "   ✅ URL auto-documenta la configuración"
echo "   ✅ Fácil debugging en logs"
echo "   ✅ No requiere headers especiales"
echo "   ✅ RESTful y estándar"
echo ""
echo "📍 Header-Based:"
echo "   ✅ URLs limpias sin prefijo /proxy"
echo "   ✅ Compatible con spec original"
echo "   ✅ Flexible para proxies transparentes"
echo "   ✅ Fácil cambiar config dinámicamente"
echo ""

# Test alias endpoints
echo "═══════════════════════════════════════════"
echo "Bonus: Endpoints Alias /rules"
echo "═══════════════════════════════════════════"
echo ""
echo "Ambos retornan lo mismo:"
echo ""

echo "1. GET /api/v1/configs"
CONFIGS_V1=$(curl -s "${BASE_URL}/api/v1/configs" | jq -r '.count')
echo "   Configs: $CONFIGS_V1"
echo ""

echo "2. GET /rules (alias)"
CONFIGS_RULES=$(curl -s "${BASE_URL}/rules" | jq -r '.count')
echo "   Configs: $CONFIGS_RULES"
echo ""

if [ "$CONFIGS_V1" == "$CONFIGS_RULES" ]; then
    echo "✅ Ambos endpoints son equivalentes"
else
    echo "⚠️  Los endpoints difieren (esto no debería pasar)"
fi

echo ""

# Cleanup
echo "3️⃣  Limpieza..."
curl -s -X DELETE "${BASE_URL}/rules/${CONFIG_ID}" > /dev/null
echo "✅ Configuración eliminada"
echo ""

echo "╔══════════════════════════════════════════════╗"
echo "║  🎉 Demo completado!                         ║"
echo "║                                              ║"
echo "║  Ahora puedes usar cualquiera de los dos    ║"
echo "║  métodos según tus necesidades.             ║"
echo "║                                              ║"
echo "║  Ver docs/DUAL_ROUTING.md para más info.    ║"
echo "╚══════════════════════════════════════════════╝"
