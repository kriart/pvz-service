#!/bin/sh
oapi-codegen -generate types -o internal/transport/http/openapi/gen/types.go -package gen api/openapi/swagger.yaml
oapi-codegen -generate chi-server -o internal/transport/http/openapi/gen/server.go -package gen api/openapi/swagger.yaml
