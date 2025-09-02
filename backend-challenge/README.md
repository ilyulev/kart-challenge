# Advanced Challenge

Build an API server implementing our OpenAPI spec for food ordering API in [Go](https://go.dev).\
You can find our [API Documentation](https://orderfoodonline.deno.dev/public/openapi.html) here.

API documentation is based on [OpenAPI3.1](https://swagger.io/specification/v3/) specification.
You can also find spec file [here](https://orderfoodonline.deno.dev/public/openapi.yaml).

> The API immplementation example available to you at orderfoodonline.deno.dev/api is simplified and doesn't handle some edge cases intentionally.
> Use your best judgement to build a Robust API server.

## Basic Requirements

- Implement all APIs described in the OpenAPI specification
- Conform to the OpenAPI specification as close to as possible
- Implement all features our [demo API server](https://orderfoodonline.deno.dev) has implemented
- Validate promo code according to promo code validation logic described below

### Promo Code Validation

You will find three big files containing random text in this repositotory.\
A promo code is valid if the following rules apply:

1. Must be a string of length between 8 and 10 characters
2. It can be found in **at least two** files

> Files containing valid coupons are couponbase1.gz, couponbase2.gz and couponbase3.gz

You can download the files from here

[file 1](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz)
[file 2](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz)
[file 3](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz)

**Example Promo Codes**

Valid promo codes

- HAPPYHRS
- FIFTYOFF

Invalid promo codes

- SUPER100

> [!TIP]
> it should be noted that there are more valid and invalid promo codes that those shown above.

## Getting Started

You might need to configure Git LFS to clone this repository\
https://github.com/oolio-group/kart-challenge/tree/advanced-challenge/backend-challenge

1. Use this repository as a template and create a new repository in your account
2. Start coding
3. Share your repository

# Implementation
A robust, production-ready Go implementation of the Oolio food ordering API following Go project layout standards and best practices.

## 🏗️ Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):
./
├── cmd/api/                    # Application entrypoints
├── internal/                   # Private application code
│   ├── handlers/              # HTTP handlers
│   ├── models/                # Data models
│   ├── services/              # Business logic
│   ├── middleware/            # HTTP middleware
│   └── config/                # Configuration
├── pkg/utils/                 # Public utility libraries
├── api/                       # OpenAPI specifications
├── deployments/               # Docker & deployment configs
├── scripts/                   # Build, install, analysis scripts
└── docs/                      # Documentation

## 🚀 Features

✅ **Clean Architecture**
- Separation of concerns (handlers, services, models)
- Dependency injection
- Testable code structure

✅ **Go Best Practices**
- Standard Go Project Layout
- Proper package organization
- Internal/external API separation

✅ **Production-Ready**
- Echo framework for high performance
- Comprehensive error handling
- Health checks and monitoring
- Docker containerization

✅ **Advanced Promo Code System**
- Concurrent file processing
- O(1) lookup performance
- Robust validation logic

## 🛠️ Quick Start

### Prerequisites
- Go 1.21 or higher
- Make (optional, for convenience commands)

### Development Setup

```bash
# Clone repository
git clone <your-repo>
cd kart-challenge/backend-api

# Install dependencies
make deps

# Run the application
make run

# Or run directly
go run ./cmd/api/main.go