# skyros
Simple Ecommerce Service

[![GitHub go.mod Go version of a Go module](https://img.shields.io/badge/Go-v1.16-green)](https://github.com/situmorangbastian/skyros/blob/main/go.mod)
[![SKYROS Actions Status](https://github.com/situmorangbastian/skyros/actions/workflows/test.yml/badge.svg)](https://github.com/situmorangbastian/skyros/actions?query=workflow%3Atest)


## Getting Started

### Project Structure
```
├── cmd/skyros              # Code for application init
├── internal                # Internal code which cannot be exported
│     ├── http              # Code for delivery layer. Mainly use HTTP as delivery layer
│     └── mysql             # Code for repository layer source layer using mysql
├── mocks                   # auto generated file which mock existing interface. Use mockery package to generate this
├── order                   # Code for usecase layer which implemented order service interface
├── product                 # Code for usecase layer which implemented product service interface
├── testdata                # Code which held helper for unit testing
├── user                    # Code for usecase layer which implemented user service interface
├── .env.example            # Config file example. Rename this to .env to use on your machine
├── context.go              # Code which held custom context with user
├── entity.go               # Code which held all structs declaration
├── helper.go               # Code which held helper for get config from env
└── skyros.go               # Code which held all interface
```
