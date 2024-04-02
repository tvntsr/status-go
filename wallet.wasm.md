## Wallet Build for WASM/WASI

Currently, the build is performed within Docker. There are two Dockerfiles provided for this purpose:

- `Dockerfile`
- `Dockerfile.tinygo`

### Standard Go Build

The `Dockerfile` is utilized for building the wallet using standard Go. The following build options are available:

- `TARGET=js`
- `TARGET=wasip1` (default)

**Usage example:**

```bash
docker build -f Dockerfile --build-arg TARGET=js -t status-go-wasm .
```

### TinyGo Build

It is also possible to build the wallet using TinyGo. In this case, `Dockerfile.tinygo` should be used.

The following build options are available:

- `TARGET=wasm`
- `TARGET=wasi` (default)

**Usage example:**

```bash
docker build -f Dockerfile.tinygo --build-arg TARGET=js -t status-go-wasm .
```

