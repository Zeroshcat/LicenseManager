# Device UUID Design

English | [Simplified Chinese](DEVICE_UUID_CN.md)

## Overview

LicenseManager uses a standard UUID string as the device identifier stored in licenses and device records. The UUID is derived from stable hardware serial data, not from an operating-system installation identifier.

For enterprise licensing, device identity should follow the physical machine. A license should remain valid after reinstalling the operating system on the same hardware, but should not remain valid after replacing the mainboard or machine. The current implementation therefore uses a fail-closed policy: if a trusted hardware serial cannot be read, device UUID generation fails.

## UUID Format

Device UUIDs use the standard UUID text format:

```text
XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
```

Example:

```text
F6235A40-C9E2-5681-B236-ED9C4C15E58D
```

The raw hardware serial is not stored directly in the license. LicenseManager canonicalizes the hardware value and hashes it into a deterministic UUID.

## Supported Hardware Sources

Device UUID generation only accepts mainboard or machine hardware serial numbers.

| Platform | Source |
| --- | --- |
| Linux | DMI mainboard serial from `/sys/class/dmi/id/board_serial`; if needed, `dmidecode -s baseboard-serial-number`. |
| macOS | Machine serial from `system_profiler SPHardwareDataType`. |
| Windows | Baseboard serial from `Win32_BaseBoard.SerialNumber` through PowerShell/CIM; if needed, `wmic baseboard get SerialNumber`. |

Multiple read methods may be tried on a platform, but they are still reading the same class of hardware identity. The implementation does not switch to mutable fallback identity sources.

## Rejected Identity Sources

The following values are intentionally not used as device identity sources:

- `/etc/machine-id`
- `/var/lib/dbus/machine-id`
- hostname
- MAC address
- operating system name or CPU architecture
- OS install time, boot ID, user profile data, or similar installation-state values
- placeholder hardware values such as `To Be Filled By O.E.M.`, `Not Specified`, `Default string`, all-zero serials, or all-`F` serials

These values are unsuitable for enterprise device binding because they can change during normal operations such as OS reinstall, network adapter replacement, virtualization configuration changes, or hostname updates.

## Generation Flow

1. Read the platform-specific hardware serial.
2. Normalize whitespace, casing, and null/BOM characters.
3. Reject known placeholder or invalid serial values.
4. Hash the normalized serial with an internal source label.
5. Format the first 16 bytes as a deterministic UUID string.

This gives the license system a stable public device ID without exposing the original hardware serial.

## CLI Usage

Run this command on the target machine:

```bash
./licensemanager uuid
# or
./licensemanager checkuuid
```

Example output:

```text
F6235A40-C9E2-5681-B236-ED9C4C15E58D
```

If the command fails, the target machine did not expose a trusted hardware serial through the supported APIs, or the serial value was rejected as a placeholder.

## Go Usage

```go
import "github.com/Zeroshcat/LicenseManager/pkg/device"

deviceUUID, err := device.GetDeviceID()
if err != nil {
    log.Fatalf("failed to get device UUID: %v", err)
}

fmt.Printf("Device UUID: %s\n", deviceUUID)
```

## Behavior Guarantees

1. **OS reinstall**: the device UUID should remain stable if the mainboard or machine hardware serial remains stable and readable.
2. **Mainboard or machine replacement**: the device UUID should change, so the old license should no longer validate.
3. **Unreadable hardware serial**: device UUID generation fails instead of silently binding to a weaker identifier.
4. **Placeholder serials**: known vendor placeholders are rejected to avoid issuing licenses to non-unique identities.

## Migration Notes

Older builds that used system UUIDs or machine IDs may produce a different device ID from the current hardware-serial policy. Licenses issued against those old IDs may need to be reissued or migrated through an explicit operational process.

Recommended migration approach:

1. Ask the customer or deployment agent to run the new `licensemanager uuid` command.
2. Confirm that the new UUID comes from a valid hardware serial source.
3. Reissue the license for the new UUID.
4. Revoke or archive the old system-ID-based license record.

## Troubleshooting

### Device UUID Cannot Be Read

Check whether the hardware serial is available on the machine.

Linux:

```bash
cat /sys/class/dmi/id/board_serial
sudo dmidecode -s baseboard-serial-number
```

macOS:

```bash
system_profiler SPHardwareDataType
```

Windows PowerShell:

```powershell
Get-CimInstance Win32_BaseBoard | Select-Object SerialNumber
```

If these commands return placeholder values such as `To Be Filled By O.E.M.` or `Not Specified`, fix the firmware/vendor asset data or use an explicit enterprise asset registration process. LicenseManager will not silently fall back to a mutable local identifier.

### UUID Differs Across Operating Systems On The Same Hardware

Linux and Windows usually expose a mainboard serial, but vendor behavior varies. macOS uses the machine serial exposed by Apple tooling. In dual-boot or virtualization environments, capture the UUID from the exact runtime environment where the licensed software will run.

### Linux Requires Elevated Permissions

Some Linux distributions restrict access to DMI data. Prefer reading `/sys/class/dmi/id/board_serial` when available. If the deployment requires `dmidecode`, document the permission requirement for the runtime or activation workflow.

## Operational Guidance

- Capture and store the generated device UUID during device registration.
- Store audit metadata separately if you need to know which hardware source was used; do not place raw serials in customer-facing license payloads unless the business explicitly requires it.
- Treat failure to read a hardware serial as an activation issue that needs operational handling, not as a reason to bind to a weaker identifier.
- For machines without reliable hardware serials, design an explicit server-side registration or enterprise asset-number workflow rather than adding automatic local fallback logic.
