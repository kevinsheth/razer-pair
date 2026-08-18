# razer-pair

Unofficial CLI for pairing supported Razer keyboards with replacement 2.4 GHz
receivers. It uses [HIDAPI](https://github.com/libusb/hidapi) on macOS, Linux,
and Windows.

Currently supported:

| Model | Keyboard | Receiver |
|---|---:|---:|
| Razer Pro Type Ultra | `1532:0277` | `1532:027b` |

The Pro Type Ultra flow has been tested on macOS hardware. Linux and Windows
use the same HIDAPI transport but have not yet been tested with real devices.

## Build

Requires Go 1.24, cgo, and a C compiler.

- macOS: Xcode Command Line Tools
- Debian/Ubuntu: `build-essential libudev-dev`
- Windows: MinGW-w64

```sh
make build
```

## Usage

Connect the keyboard by USB and plug in the receiver:

```sh
./bin/razer-pair inspect
./bin/razer-pair dry-run
./bin/razer-pair pair
```

`pair` asks before sending the final commit. Use `pair --yes` for
non-interactive use. Use `inspect --verbose` for complete HID diagnostics.

After pairing, disconnect the keyboard cable and switch it to 2.4 GHz.

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

## License

MIT. See [LICENSE](LICENSE) and [third-party notices](THIRD_PARTY_NOTICES.md).

Razer is a trademark of Razer Inc. This project is not affiliated with Razer.
