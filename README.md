# razer-pair

Unofficial CLI for pairing supported Razer peripherals with replacement
2.4 GHz receivers. It uses [HIDAPI](https://github.com/libusb/hidapi) on macOS,
Linux, and Windows.

Currently supported:

| Model | Device | Receiver | Status |
|---|---:|---:|---|
| Razer Pro Type Ultra | `1532:0277` | `1532:027b` | Pairing tested |
| Razer Basilisk Ultimate | `1532:0086` | `1532:0088` | Pairing tested |
| Razer Pro Click V2 | `1532:00d0` | `1532:00d1` | Hardware test pending |

The tested pairing flows were verified on macOS hardware. Linux and Windows use
the same HIDAPI transport but have not yet been tested with real devices.

## Build

Requires Go 1.24, cgo, and a C compiler.

- macOS: Xcode Command Line Tools
- Debian/Ubuntu: `build-essential libudev-dev`
- Windows: MinGW-w64

```sh
make build
```

## Usage

Find connected Razer devices without sending reports:

```sh
./bin/razer-pair scan
```

Connect the supported device by USB, plug in its receiver, and select its
profile:

```sh
./bin/razer-pair --model pro-type-ultra inspect
./bin/razer-pair --model pro-type-ultra dry-run
./bin/razer-pair --model pro-type-ultra pair
```

`pair` asks before sending the final commit. Use `pair --yes` for
non-interactive use. Use `inspect --verbose` for complete HID diagnostics.

After pairing, disconnect the cable and switch the device to 2.4 GHz.

### Permissions

- macOS: grant Input Monitoring to the terminal or app running `razer-pair`.
- Linux: configure a udev rule that permits access to the HID interface.
- Windows: no additional setup is expected.

## Development

```sh
make check
./bin/razer-pair --mock success pair --yes
./bin/razer-pair --mock mismatch pair --yes
```

The CLI accepts only model-specific pairing commands. It rejects unknown USB
IDs, unexpected report sizes, malformed responses, and failed identity checks.

See [protocol notes](docs/protocol.md) for the report layout and command
sequence.

### Adding a device

Run `razer-pair scan` and include its short output in an issue. A new profile
needs exact USB IDs, accessible feature-report sizes, a verified command
sequence, a mock transcript, and a real-hardware test before it is marked
pairing-tested. Use `scan --verbose` only when deeper HID diagnostics are
requested.

## License

MIT. See [LICENSE](LICENSE) and [third-party notices](THIRD_PARTY_NOTICES.md).

Razer is a trademark of Razer Inc. This project is not affiliated with Razer.
