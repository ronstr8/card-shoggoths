# Card Shoggoths - Version Management

## Version Source of Truth

**Version is defined in:** `cmd/card-shoggoths-server/version.go`

```go
const Version = "1.0.0"
```

## Updating the Version

1. **Edit `version.go`**: Change the `Version` constant
2. **Sync `Chart.yaml`**: Update `version` and `appVersion` to match
3. **Build & Deploy**: The Makefile automatically uses the Go version

## Version Display

```bash
# Show current version
go run ./cmd/card-shoggoths-server --version

# Makefile uses this version for image tags
make deploy  # Will tag as localhost:5000/card-shoggoths:1.0.0
```

## Version Scheme

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., 1.0.0)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes
