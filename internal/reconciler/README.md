# Internal Reconciler Package

This package implements the core reconciliation loop engine for the Eph application.

## Purpose

The reconciler package provides the generic reconciliation control loop that drives Eph's reconciliation-first architecture.

## Responsibilities

- Implement the generic reconciliation control loop (every 30s)
- Handle reconciliation mechanics, timing, and error handling
- Coordinate between informers and controllers
- Provide reusable reconciliation framework across different resource types

## Architecture

This implements the Kubernetes controller pattern where:
- Reconciliation is primary (level-based)
- Webhooks are optimization (edge-based)
- External systems are source of truth
- Crash-only design with automatic recovery

## Key Principles

- **Level-based primary**: Poll external sources every 30s for eventual consistency
- **Edge-based optimization**: Webhooks trigger immediate reconciliation but aren't required
- **No internal source of truth**: GitHub defines what should exist, providers report what does exist
- **Stateless**: No persistent state, recovery is just normal startup

## Implementation Status

🚧 **Placeholder** - Not yet implemented
