#!/bin/bash

set -e

echo "🚀 Starting Nova Test Environment..."

if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose not found. Please install docker-compose first."
    exit 1
fi

case "$1" in
    up)
        echo "📦 Starting services..."
        docker-compose -f docker-compose.test.yml up -d
        echo "⏳ Waiting for services to be ready..."
        sleep 5
        docker-compose -f docker-compose.test.yml ps
        echo "✅ Services started successfully!"
        echo "🌐 Server: http://localhost:8080"
        echo "🗄️  PostgreSQL: localhost:5432"
        echo "💾 Redis: localhost:6379"
        ;;
    down)
        echo "🛑 Stopping services..."
        docker-compose -f docker-compose.test.yml down
        echo "✅ Services stopped!"
        ;;
    restart)
        echo "🔄 Restarting services..."
        docker-compose -f docker-compose.test.yml restart
        echo "✅ Services restarted!"
        ;;
    logs)
        docker-compose -f docker-compose.test.yml logs -f
        ;;
    clean)
        echo "🧹 Cleaning up..."
        docker-compose -f docker-compose.test.yml down -v
        docker volume rm nova_postgres_data nova_redis_data 2>/dev/null || true
        echo "✅ Cleanup complete!"
        ;;
    status)
        docker-compose -f docker-compose.test.yml ps
        ;;
    build)
        echo "🔨 Building images..."
        docker-compose -f docker-compose.test.yml build --no-cache
        echo "✅ Build complete!"
        ;;
    rebuild)
        echo "🔨 Rebuilding and restarting..."
        docker-compose -f docker-compose.test.yml down
        docker-compose -f docker-compose.test.yml build --no-cache
        docker-compose -f docker-compose.test.yml up -d
        echo "✅ Rebuild complete!"
        ;;
    *)
        echo "Usage: $0 {up|down|restart|logs|clean|status|build|rebuild}"
        echo ""
        echo "Commands:"
        echo "  up       - Start all services"
        echo "  down     - Stop all services"
        echo "  restart  - Restart all services"
        echo "  logs     - View logs (follow mode)"
        echo "  clean    - Stop services and remove volumes"
        echo "  status   - Show service status"
        echo "  build    - Build Docker images"
        echo "  rebuild  - Rebuild and restart services"
        exit 1
        ;;
esac
