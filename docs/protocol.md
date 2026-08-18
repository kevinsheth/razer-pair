# Pairing protocol notes

## Feature report

The supported devices expose a 90-byte HID feature report. Requests use:

| Offset | Length | Meaning |
|---:|---:|---|
| 0 | 1 | Status; zero in requests |
| 1 | 1 | Transaction ID (`0x1f` wired device, `0xff` receiver) |
| 5 | 1 | Payload length (`6`) |
| 6 | 1 | Command class (`0x00`) |
| 7 | 1 | Command |
| 8 | 6 | Command payload |
| 88 | 1 | XOR checksum of bytes 2 through 87 |

A successful response must be exactly 90 bytes, use status `0x02`, echo command
class `0x00` and the requested command, and contain a valid checksum.

## Commands

| Model | Identity | Prepare | Commit |
|---|---:|---:|---:|
| Pro Type Ultra | receiver `0x95` | keyboard `0x24` | keyboard `0xa4` |
| Basilisk Ultimate | mouse `0x97` | receiver `0x15` | receiver `0x95` |

## Pairing sequence

1. Send the profile's identity command with six zero bytes. Preserve its
   six-byte response.
2. Send the prepare command to its profile-defined target with that exact
   response. Preserve the prepare response as the handshake result.
3. Obtain user confirmation.
4. Send the device commit command with a new six-byte zero buffer.
5. Compare the first four bytes returned by prepare and commit. Pairing is
   successful only when they match.

The buffer reuse in step 2 and fresh zero buffer in step 4 are required.

The Basilisk profile was derived by static analysis of Razer Mouse Pairing
Utility v1.00.07_r2 (SHA-256
`18743c5b3253a5308df98abd7d0011bed48a0dbef858cb2a2ec2f952ab03aba2`).
