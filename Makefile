.PHONY: build test lint arch-check

build:
	go build -o kervo ./cmd/kervo

test:
	go test ./...

# ARCH-0001 §2: core는 adapters/외부 I/O를 import할 수 없다 (역방향 의존 금지)
arch-check:
	@! grep -rn "internal/adapters" internal/core internal/ports \
		&& echo "OK: core/ports do not import adapters" \
		|| (echo "VIOLATION: core imports adapters" && exit 1)
