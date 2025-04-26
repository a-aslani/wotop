
## Overview

WOTOP is an open‑source Go framework designed to accelerate backend development with modern architectural patterns. It brings together:

- **Clean Architecture** for strict separation of business logic and infrastructure layers
- **Domain‑Driven Design (DDD)** to model complex domains with clarity
- **Event‑Driven Microservices** for loosely‑coupled, asynchronous communication
- **Cloud‑Native Microservices** optimized for containerized, orchestrated environments

Additionally, WOTOP integrates core patterns and tools out of the box:

- **CQRS** (Command Query Responsibility Segregation)
- **RabbitMQ** message broker
- **Event Sourcing** for append‑only event storage

---

## 🎯 Features

- **Layered Clean Architecture**  
  Enforces the Dependency Rule, isolating Use Cases, Entities, and Interfaces for maximum testability and maintainability.

- **Domain‑Driven Design**  
  Implements Aggregates, Repositories, and Bounded Contexts so your code mirrors real‑world domain concepts.

- **Event‑Driven Communication**  
  Built‑in event bus with publish/subscribe support for decoupled services.

- **Cloud‑Native Ready**  
  Services are containerized and Kubernetes‑friendly, with support for self‑healing, auto‑scaling, and sidecar integrations.

- **CQRS**  
  Separates read and write workloads into dedicated models for better performance and scalability.

- **RabbitMQ Integration**  
  Reliable, configurable message broker support for asynchronous workflows.

- **Event Sourcing**  
  Append‑only event store that lets you reconstruct any entity’s state at any point in time.

---

## 🚀 Quick Start

**WOTOP CLI** is a lightweight CLI tool for scaffolding Go backend code. It provides two main commands:

- `usecase` – generates a usecase directory with `inport.go`, `outport.go`, and `interactor.go` files.
- `entity`  – generates an entity file under your models with struct definitions and helper methods.

This README covers:

1. Installation
2. Command: `wotop usecase`
3. Command: `wotop entity`
4. Examples

---

## Installation

Install globally

```bash
# Installs to $GOBIN or $GOPATH/bin
go install github.com/a-aslani/wotop/cmd/wotop@latest
```
```bash
go get -u github.com/a-aslani/wotop@latest
```

Make sure $GOBIN (or $HOME/go/bin) is in your PATH so you can run wotop from anywhere.

---

## `usecase` Command
  
```bash
wotop usecase <domain> <name>
```

`<domain>`
- The parent domain under internal/. 
- E.g. if you pass product, files go under internal/product.

`<name>`
- The usecase identifier. 
- The folder name is converted to snake_case. 
- The Go package inside will match that snake_case folder.


## What it creates

Given:
```bash
wotop usecase product getUserInfo
```

It will create:
```
internal/
└── product/
    └── usecase/
        └── get_user_info/
            ├── inport.go
            ├── outport.go
            └── interactor.go
```

### `inport.go`
```go
package get_user_info

import "github.com/a-aslani/wotop"

type Inport = wotop.Inport[InportRequest, InportResponse]

type InportRequest struct {
    // request fields
}

type InportResponse struct {
    // response fields
}
```

### `outport.go`
```go
package get_user_info

type Outport interface {
    // define methods to call downstream adapters
}
```

### `interactor.go`
```go
package get_user_info

import "context"

type interactor struct {
    outport Outport
}

func NewUsecase(outport Outport) Inport {
    return &interactor{outport: outport}
}

func (i interactor) Execute(ctx context.Context, req InportRequest) (*InportResponse, error) {
    res := InportResponse{}
    // TODO: implement usecase logic
    return &res, nil
}
```

## `entity` Command

```bash
wotop entity <domain> <name>
```

`<domain>`
- The parent domain under internal/.
- E.g. product → will place files in internal/product/model/entity.

`<name>`
- The entity name in camelCase, snake_case, or PascalCase.
- Internally the struct name is converted to PascalCase.
- The file name is generated in snake_case.

## What it creates

Given:
```bash
wotop entity product userState
```

It will create:
```
internal/
└── product/
    └── model/
        └── entity/
            └── user_state.go
```

### `user_state.go`
```go
package entity

type UserState struct {}

type UserStateFilter struct {}

type CreateUserStateRequest struct {}

func (c CreateUserStateRequest) Validate() error {
    // TODO: add validation logic
    return nil
}

func NewUserState(req CreateUserStateRequest) (*UserState, error) {
    // TODO: add creation logic
    return &UserState{}, nil
}
```