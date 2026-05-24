# LicenseManager

English | [Simplified Chinese](README_CN.md)

An enterprise-oriented license management toolkit written in Go. It supports offline licenses, online verification, and dual verification, with hardware-bound device IDs and a built-in web administration console.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Highlights

- **Three license modes**: offline, online, and dual verification.
- **Hardware-bound device identity**: device IDs are derived from stable mainboard or machine hardware serial numbers.
- **Fail-closed device binding**: the device ID code does not fall back to `/etc/machine-id`, hostnames, MAC addresses, or other mutable installation/network identifiers.
- **Cryptographic protection**: AES-256-GCM for license payload encryption and RSA-4096 for signature verification.
- **Web administration console**: generate, download, list, and revoke licenses from a browser UI.
- **Single CLI tool**: initialization, license generation, verification, device inspection, and admin server commands are exposed through `licensemanager`.
- **SQLite storage**: lightweight local database backed by pure Go SQLite dependencies.

## License Modes

| Mode | Behavior | Typical use case |
| --- | --- | --- |
| Offline | Validates a signed and encrypted local `license.key` file. | Air-gapped deployments, private networks, on-premise software. |
| Online | Calls a license server to verify the current device. | SaaS-connected clients, centrally controlled activation. |
| Dual | Requires both local offline validation and server-side verification. | Higher assurance deployments where local tampering and server-side revocation must both be considered. |

## Device Binding Policy

LicenseManager binds licenses to a generated device UUID. The UUID is derived from stable hardware serial data instead of system installation identifiers.

Supported sources:

- **Linux**: DMI mainboard serial from `/sys/class/dmi/id/board_serial`, with `dmidecode -s baseboard-serial-number` as a hardware-source read method.
- **macOS**: machine serial from `system_profiler SPHardwareDataType`.
- **Windows**: baseboard serial from PowerShell/CIM or `wmic baseboard get SerialNumber`.

Not used as device identity sources:

- `/etc/machine-id` or `/var/lib/dbus/machine-id`
- hostname
- MAC address
- OS name, CPU architecture, install time, or other installation-state values
- placeholder hardware values such as `To Be Filled By O.E.M.`, `Not Specified`, all-zero serials, or all-`F` serials

If no trusted hardware serial can be read, device ID generation fails. This is intentional for enterprise licensing: a license should survive OS reinstall on the same hardware, but should not silently bind to a mutable fallback value.

See [docs/DEVICE_UUID.md](docs/DEVICE_UUID.md) for the full device UUID design and troubleshooting guide.

## Project Layout

```text
LicenseManager/
├── README.md                # English documentation (default)
├── README_CN.md             # Simplified Chinese documentation
├── cmd/
│   ├── licensemanager/      # Main CLI tool
│   └── server/              # Optional license server entrypoint
├── internal/
│   ├── admin/               # Web admin console
│   ├── auth/                # Admin authentication helpers
│   ├── crypto/              # RSA/AES helpers
│   ├── database/            # SQLite models and database access
│   ├── license/             # License generation and verification internals
│   └── server/              # Online verification HTTP server
├── pkg/
│   ├── device/              # Hardware-bound device UUID generation
│   ├── license/             # Public license verification package
│   └── output/              # CLI output formatting
├── docs/                    # Design and security notes
├── examples/                # Integration examples
└── web/                     # Static web assets
```

## Requirements

- Go 1.21 or later
- Network access to download Go modules on first build, unless dependencies are already present in the module cache
- Hardware serial access on target machines where `licensemanager uuid` or automatic device ID detection is used

SQLite is provided through Go modules. No CGO SQLite installation is required for the default build.

## Build

```bash
git clone https://github.com/Zeroshcat/LicenseManager.git
cd LicenseManager

# Build the main CLI tool
go build -o licensemanager ./cmd/licensemanager

# Build the optional license server
go build -o license-server ./cmd/server
```

On Windows, the output binaries will usually be named `licensemanager.exe` and `license-server.exe`.

## Quick Start

### 1. Initialize Keys And Database

```bash
./licensemanager init
```

This creates:

- `license.db`: SQLite database
- `private_key.pem`: RSA private key used to sign licenses
- `public_key.pem`: RSA public key used to verify licenses
- `aes_key.bin`: AES key used to encrypt/decrypt license payloads

Keep these files secure. If the key material is lost, existing licenses cannot be regenerated or verified with a new key set.

### 2. Get The Device UUID

Run this on the target device:

```bash
./licensemanager uuid
# or
./licensemanager checkuuid
```

Example output:

```text
F6235A40-C9E2-5681-B236-ED9C4C15E58D
```

If this command fails, the machine did not expose a trusted hardware serial through the supported platform APIs. See [Device UUID Troubleshooting](docs/DEVICE_UUID.md#troubleshooting).

### 3. Generate A License

```bash
# Offline license
./licensemanager generate --type offline --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --expiry 2026-12-31 --output license.key

# Online license
./licensemanager generate --type online --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --expiry 2026-12-31

# Dual verification license
./licensemanager generate --type dual --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --expiry 2026-12-31 --output license.key
```

If `--device-id` is omitted, the CLI attempts to read the current machine's hardware-bound device UUID.

### 4. Verify A License

```bash
# Offline verification with automatic current-device ID detection
./licensemanager verify --license-file license.key

# Offline verification with an explicit device ID
./licensemanager verify --license-file license.key --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D

# Online verification
./licensemanager verify --online --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --api-url http://localhost:8080

# Dual verification
./licensemanager verify --dual --license-file license.key --device-id F6235A40-C9E2-5681-B236-ED9C4C15E58D --api-url http://localhost:8080
```

### 5. Run The Web Admin Console

```bash
./licensemanager admin serve --passwd your_strong_password --port 8080
```

Then open:

```text
http://localhost:8080
```

The web console supports device listing, license listing, license generation, and license download.

## CLI Overview

```text
licensemanager init [--db license.db]
licensemanager uuid
licensemanager checkuuid
licensemanager generate --type offline|online|dual --device-id <uuid> --expiry YYYY-MM-DD [--output license.key]
licensemanager verify [--license-file license.key] [--device-id <uuid>] [--online|--dual] [--api-url <url>]
licensemanager device list [--format text|json]
licensemanager device show [device-uuid]
licensemanager device bind <device-uuid>
licensemanager admin serve --passwd <password> [--host 0.0.0.0] [--port 8080] [--db license.db]
licensemanager admin token create --type client|admin [--app-id <app-id>] [--expires YYYY-MM-DD]
licensemanager version
```

Run `licensemanager <command> --help` for command-specific flags.

## Online Server Endpoints

When the server/admin components are running, the license APIs are exposed under `/api/v1`:

| Endpoint | Method | Purpose |
| --- | --- | --- |
| `/api/v1/license/verify/online` | `POST` | Online license verification. |
| `/api/v1/license/verify/dual` | `POST` | Dual-mode server verification. |
| `/api/v1/device/register` | `POST` | Register a device record. |
| `/api/v1/device/{device_id}` | `GET` | Fetch device and license status. |

## Go Integration

### Offline Verification

```go
package main

import (
    "log"

    "github.com/Zeroshcat/LicenseManager/pkg/device"
    "github.com/Zeroshcat/LicenseManager/pkg/license"
)

func main() {
    deviceID, err := device.GetDeviceID()
    if err != nil {
        log.Fatalf("failed to get device ID: %v", err)
    }

    licenseKey, err := license.LoadLicenseFromFile("license.key")
    if err != nil {
        log.Fatalf("failed to load license: %v", err)
    }

    publicKeyPEM := []byte("...")
    aesKey := []byte("...")

    verifier, err := license.NewOfflineVerifier(publicKeyPEM, aesKey)
    if err != nil {
        log.Fatalf("failed to create verifier: %v", err)
    }

    result, err := verifier.Verify(licenseKey, deviceID)
    if err != nil || !result.Valid || result.Expired {
        log.Fatalf("license verification failed: %v", err)
    }
}
```

### Online Verification

```go
verifier := license.NewOnlineVerifier(&license.OnlineConfig{
    APIURL:  "http://localhost:8080/api/v1",
    AppID:   "default",
    Timeout: 10,
})

result, err := verifier.Verify(deviceID)
```

### Dual Verification

```go
verifier, err := license.NewDualVerifier(&license.DualConfig{
    APIURL:  "http://localhost:8080/api/v1",
    AppID:   "default",
    Timeout: 10,
}, publicKeyPEM, aesKey)

result, err := verifier.Verify(licenseKey, deviceID)
```

For a complete key-embedding example, see [examples/embedded_keys](examples/embedded_keys).

## Security Notes

- Treat `private_key.pem` and `aes_key.bin` as production secrets.
- Do not ship the private key with client applications.
- Prefer embedding or securely provisioning the public key and AES key for offline verification clients.
- Protect the admin console with a strong password and place it behind HTTPS, a VPN, or a trusted internal network.
- Use dual verification when server-side revocation is required.
- Existing licenses are tied to the key set that generated them. Replacing keys invalidates previously issued licenses unless a migration plan is implemented.

## Device ID Migration Note

Older builds that used system UUIDs or machine IDs may produce different device IDs than the current hardware-serial policy. After upgrading, licenses issued against the old identifier may need to be reissued or migrated through an explicit operational process.

## Documentation

- [Device UUID design](docs/DEVICE_UUID.md)
- [Device UUID design - Chinese](docs/DEVICE_UUID_CN.md)
- [Key replacement scenario](docs/KEY_REPLACEMENT_SCENARIO.md)
- [Attack scenario](docs/ATTACK_SCENARIO.md)
- [Security analysis](docs/SECURITY_ANALYSIS.md)
- [Embedded keys example](examples/embedded_keys/README.md)
- [Simplified Chinese README](README_CN.md)

## Testing

```bash
# Device package tests
go test ./pkg/device

# Main non-database packages
go test ./internal/auth ./internal/crypto ./internal/license ./pkg/device ./pkg/license ./pkg/output

# Full repository tests, after dependencies are available
go test ./...
```

The full test run may download Go modules such as `gorm` and `modernc.org/sqlite` on a clean machine.

## FAQ

### Will a license survive OS reinstall?

Yes, when the mainboard or machine hardware serial remains the same and is readable after reinstall. The device ID is not based on `/etc/machine-id` or similar installation IDs.

### Will a license survive mainboard replacement?

No. Mainboard or machine replacement changes the hardware identity, so a new license should be issued.

### Why not fall back to MAC address or hostname?

Those values are mutable and easy to change. Silent fallback would create licenses that appear valid but break after normal operations such as OS reinstall, NIC changes, virtualization changes, or hostname updates.

### Can licenses be transferred?

Not directly. A license is bound to a device UUID. To move software to another device, revoke or remove the old license record and issue a new license for the new device UUID.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).

## Contributing

Issues and pull requests are welcome. For security-sensitive changes, include the threat model and test coverage that supports the change.

## Author

LicenseManager Team
