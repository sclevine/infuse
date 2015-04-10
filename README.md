Infuse
======

[![Build Status](https://api.travis-ci.org/sclevine/infuse.png?branch=master)](http://travis-ci.org/sclevine/infuse)
[![GoDoc](https://godoc.org/github.com/sclevine/infuse?status.svg)](https://godoc.org/github.com/sclevine/infuse)

Infuse provides an immutable, concurrency-safe middleware handler
that conforms to http.Handler. An infuse.Handler is fully compatible with
the standard library, supports flexible chaining, and provides a shared
context between middleware handlers within a single request without
relying on global state, locks, or closures.