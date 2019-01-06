# vexdump

**vexdump** prints VEX-like encodings information.
Recognizes VEX and EVEX prefix formats.

Notes:
* encoding hex string case does not matter
* spaces are allowed (removed before parsing)
* encoding hex string length should not exceed 30.

Please note that "fields" column will not match Intel Software Developer
manual in some cases.  

For one such example, `128` is used instead of `LIG`.

## Install

```bash
go get -u -v github.com/Quasilyte/devtools/cmd/vexdump
```

## Usage

```
// Dump single instruction info:
vexdump 6272fd098ae8
EVEX rxbR00mm Wvvvv1pp zLlbVaaa opcode modrm    fields
62   01110010 11111101 00001001 8A     11101000 EVEX.128.66.0F38.W1

// Dump multiple instructions for comparison:
vexdump 6272FD098AE8 '62 72 fd 09 8a c5'
EVEX rxbR00mm Wvvvv1pp zLlbVaaa opcode modrm    fields
62   01110010 11111101 00001001 8A     11101000 EVEX.128.66.0F38.W1
62   01110010 11111101 00001001 8A     11000101 EVEX.128.66.0F38.W1

// Dump instructions with different encoding schemes (VEX and EVEX):
vexdump 6272fd098ae8 6272fd098ac5 c4e1315813 c5b15813 c5f877
VEX2 rvvvvlpp opcode modrm    fields
C5   10110001 58     00010011 VEX.128.66.0F.W0
C5   11111000 77     00000000 VEX.128.0F.W0
VEX3 rxbmmmmm Wvvvvlpp opcode modrm    fields
C4   11100001 00110001 58     00010011 VEX.128.66.0F.W0
EVEX rxbR00mm Wvvvv1pp zLlbVaaa opcode modrm    fields
62   01110010 11111101 00001001 8A     11101000 EVEX.128.66.0F38.W1
62   01110010 11111101 00001001 8A     11000101 EVEX.128.66.0F38.W1
```
