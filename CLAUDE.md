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
