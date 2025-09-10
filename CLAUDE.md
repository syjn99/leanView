# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 📋 Project Overview

**Lean View** is a PQ Devnet Visualizer Web Application - a lightweight monitoring tool for visualizing the health and status of the PQ Devnet.

📚 **Documentation**: See the PQ devnet specifications at [`docs/lean-specs/`](./docs/lean-specs/)

## 🏗️ Project Structure

```
leanView/
├── backend/     # Go backend services
├── frontend/    # React frontend application  
├── proto/       # Protocol buffer definitions
└── docs/        # Documentation and specifications
```

## 🎯 Project Goal

Create a lightweight, standalone visualization tool inspired by:
- [**Forkmon**](https://github.com/ethereum/nodemonitor) - Ethereum network monitor
- [**Dora**](https://github.com/ethpandaops/dora) - Ethereum beacon chain explorer

## ✅ Core Requirements

### Monitoring & Visualization
- **Multiple Client Support**: Connect to and display data from various client endpoints (Rust, Zig, etc.)
- **Chain State Visualization**: 
  - 3-Slot Finality model tracking
  - Head state monitoring
  - Recent blocks/slots display

### Architecture Requirements  
- **Standalone Deployment**: Single Docker image with no external cloud dependencies
- **Configuration**: YAML-based configuration (`config.yaml`)
- **Kurtosis Integration**: Deployable as an additional service within Kurtosis framework

## 🛠️ Technology Stack

### Backend
- **Language**: Go
- **Database**: SQLite
- **DB Access**: sqlc
- **API**: Protobuf + Connect RPC

### Frontend
- **Framework**: React
- **Build Tool**: Vite
- **API Client**: Connect RPC

### Infrastructure
- **Containerization**: Docker & Docker Compose
- **Configuration**: YAML-based config files

## 📡 Backend API Usage

The backend provides Connect RPC APIs for accessing blockchain and monitoring data. The server runs on `http://localhost:8080` by default.

### Available Services

#### 1. BlockService
- **GetLatestBlockHeader**: Returns the current head block from the in-memory cache

```bash
# Get latest block header
curl -X POST http://localhost:8080/api.v1.BlockService/GetLatestBlockHeader \
  -H "Content-Type: application/json" \
  -d '{}' | jq
```

#### 2. MonitoringService  
- **GetAllClientsHeads**: Returns the current head block from all connected clients with their health status

```bash
# Get all clients' head blocks
curl -X POST http://localhost:8080/api.v1.MonitoringService/GetAllClientsHeads \
  -H "Content-Type: application/json" \
  -d '{}' | jq
```

Sample response:
```json
{
  "clientHeads": [
    {
      "clientLabel": "local",
      "endpointUrl": "http://127.0.0.1:5052",
      "isHealthy": true,
      "blockHeader": {
        "slot": "90",
        "proposerIndex": "2",
        "parentRoot": "0x08d024...",
        "stateRoot": "0xcd2412...",
        "bodyRoot": "0x263448..."
      },
      "blockRoot": "0x52ad43...",
      "lastUpdateMs": "1757474205827"
    }
  ],
  "totalClients": 1,
  "healthyClients": 1
}
```

### Testing the API

1. **Check server status**:
```bash
curl http://localhost:8080/
```

2. **Run the backend server**:
```bash
cd backend
go run cmd/main.go
```

3. **Test endpoints**: Use the curl commands above or any Connect RPC client
