# Pro Type Ultra protocol notes

## Feature report

The supported devices expose a 90-byte HID feature report. Requests use:

| Offset | Length | Meaning |
|---:|---:|---|
| 0 | 1 | Status; zero in requests |
| 1 | 1 | Transaction ID (`0x1f` keyboard, `0xff` receiver) |
| 5 | 1 | Payload length (`6`) |
| 6 | 1 | Command class (`0x00`) |
| 7 | 1 | Command |
| 8 | 6 | Command payload |
| 88 | 1 | XOR checksum of bytes 2 through 87 |

A successful response must be exactly 90 bytes, use status `0x02`, echo command
class `0x00` and the requested command, and contain a valid checksum.

## Pairing sequence

1. Send receiver command `0x95` with six zero bytes. Its six-byte response is
   the receiver identity buffer.
2. Send keyboard command `0x24`, passing that exact six-byte buffer. Preserve
   the six-byte response as the handshake result.
3. Obtain user confirmation.
4. Send keyboard command `0xa4` with a new six-byte zero buffer.
5. Compare the first four bytes returned by `0x24` and `0xa4`. Pairing is
   reported as successful only when they match.

The buffer reuse in step 2 and fresh zero buffer in step 4 are required.
