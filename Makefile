.PHONY: generate-oapi

generate-oapi:
	@set -e; \
	echo "📦 Bundling OpenAPI spec..."; \
	npx @redocly/cli bundle api/openapi/openapi.yaml -o api/openapi/openapi.bundled.yaml; \
	echo "✅ Bundled spec"; \
	echo "🧾 Building HTML docs..."; \
	npx @redocly/cli build-docs api/openapi/openapi.bundled.yaml -o api/docs/index.html; \
	echo "✅ Docs"; \
	echo "⚙️ Generating Go code..."; \
	go tool oapi-codegen -config api/openapi/oapi-codegen.cfg.yaml api/openapi/openapi.bundled.yaml; \
	echo "✅ Go API generated"
