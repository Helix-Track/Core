#!/bin/bash
# HelixTrack Core - Benchmark Runner
# Runs all performance benchmarks

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}=============================================${NC}"
echo -e "${BLUE}  Performance Benchmarks                    ${NC}"
echo -e "${BLUE}=============================================${NC}"
echo ""

cd "$PROJECT_ROOT"

mkdir -p coverage/benchmarks

echo -e "${YELLOW}Running all benchmarks...${NC}"
echo ""

# Run benchmarks with memory statistics
go test ./... -bench=. -benchmem -benchtime=1s -run=^$ | tee coverage/benchmarks/results.txt

echo ""
echo -e "${GREEN}✓ Benchmarks completed${NC}"
echo -e "${BLUE}Results saved to: coverage/benchmarks/results.txt${NC}"
echo ""

# Extract key metrics
echo -e "${BLUE}Key Performance Metrics:${NC}"
echo ""
grep -E "Benchmark(Cache|Database|Security|JWT|Permission)" coverage/benchmarks/results.txt | head -20 || echo "No benchmark results found"
