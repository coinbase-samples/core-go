# Core Package README

## Overview 

The core package provides a centralized and reusable implementation for making HTTP requests and handling API responses for Coinbase Institutional Go SDKs. It includes features for setting custom headers, managing credentials, and providing structured error handling.

## Installation 

The core package is already integrated with the [Coinbase Prime](https://github.com/coinbase-samples/prime-sdk-go) and [Coinbase International Exchange (INTX)](https://github.com/coinbase-samples/intx-sdk-go) Go SDKs. To manually install the core package, use the following command:

```
go get github.com/coinbase-samples/core-go
```

## Usage

To use the core package, import it into your project:

```go
import "github.com/coinbase-samples/core-go"
```

Then, create a new instance of the `Client` struct:

```go
client := core.NewClient()
```
