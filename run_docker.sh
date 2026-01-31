#!/bin/bash
echo "Starting JJMC in Docker..."
docker-compose up -d --build
echo "----------------------------------------"
echo "JJMC is running at http://localhost:3001"
echo "To stop it, run: docker-compose down"
