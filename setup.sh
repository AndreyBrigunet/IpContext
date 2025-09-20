#!/bin/bash

# IP API Service Setup Script
# This script helps you set up the IP API service quickly

set -e

echo "🚀 IP API Service Setup"
echo "======================="

# Check if .env file exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ .env file created. Please edit it with your credentials:"
    echo "   - GEOIPUPDATE_ACCOUNT_ID"
    echo "   - GEOIPUPDATE_LICENSE_KEY"
    echo "   - GEONAMES_USERNAME (optional, for neighbours/languages)"
    echo ""
    echo "⚠️  Please edit .env file before continuing!"
    read -p "Press Enter when you've configured .env file..."
fi

# Check if required environment variables are set
source .env

if [ -z "$GEOIPUPDATE_ACCOUNT_ID" ] || [ -z "$GEOIPUPDATE_LICENSE_KEY" ]; then
    echo "❌ Error: GEOIPUPDATE_ACCOUNT_ID and GEOIPUPDATE_LICENSE_KEY must be set in .env file"
    exit 1
fi

echo "✅ Environment configuration looks good!"

# Ask user for deployment type
echo ""
echo "🔧 Choose deployment type:"
echo "1) Production (use pre-built image from GitHub)"
echo "2) Development (build locally)"
read -p "Enter choice (1 or 2): " choice

case $choice in
    1)
        echo "🐳 Starting production deployment..."
        docker compose pull
        docker compose up -d
        ;;
    2)
        echo "🔨 Starting development deployment..."
        docker compose -f docker-compose.dev.yml build
        docker compose -f docker-compose.dev.yml up -d
        ;;
    *)
        echo "❌ Invalid choice. Exiting."
        exit 1
        ;;
esac

echo ""
echo "⏳ Waiting for services to start..."
sleep 10

# Health check
echo "🏥 Performing health check..."
if curl -f http://localhost:3280/health > /dev/null 2>&1; then
    echo "✅ Service is healthy!"
else
    echo "⚠️  Service might still be starting. Check logs with:"
    echo "   docker compose logs -f"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "📋 Available endpoints:"
echo "   • http://localhost:3280/ - Your IP info"
echo "   • http://localhost:3280/8.8.8.8 - Specific IP info"
echo "   • http://localhost:3280/health - Health check"
echo ""
echo "📊 Useful commands:"
echo "   • docker compose logs -f - View logs"
echo "   • docker compose down - Stop services"
echo "   • docker compose pull && docker compose up -d - Update to latest"
