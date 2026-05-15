# easydi

Reflection-free, compile-time dependency-injection code generator for Go.

## Usage

Annotate constructors and root types with `// di:` comments:

```go
// di:root
type Config = config.Config

// di:provide
func NewManager(
	// di:param Config.Auth
	cfg config.AuthConfig,
) (*Manager, error) { ... }
```

Generate the container:

```sh
easydi gen -o internal/di/easydi_gen.go -pkg di ./...
```

`-o` defaults to `easydi_gen.go` in the current directory; any missing
parent directories in the `-o` path are created automatically. `-pkg` sets
the `package` name in the generated file and is required.

This emits a `Container` struct, `Build(<roots>) (*Container, error)`, and
`Exposed() []any` for `// di:expose` nodes. No `reflect` at build or run time.

## Annotations

| Annotation | Placement | Meaning |
|---|---|---|
| `// di:provide` | constructor func | graph node |
| `// di:provide name=X` | constructor func | explicit node name |
| `// di:root` | type decl | external input to `Build` |
| `// di:param <path>` | line above a parameter | root projection or `pkg.Ident` literal |
| `// di:expose` | constructor func | surfaced via `Exposed()` |
