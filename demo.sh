#!/bin/bash
docker run --rm -it -p 8080:8080 -v$(pwd):/app --workdir=/app golang:1.19-alpine go run . -input ./demo/input -output ./demo/output -left not_sure -right cat -right dog