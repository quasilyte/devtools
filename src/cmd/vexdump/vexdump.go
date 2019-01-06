// vexdump prints VEX-like encodings information.
//
// Usage:
//
//	vexdump 6272fd098ae8
//	vexdump 6272FD098AE8 '62 72 fd 09 8a c5'
//	vexdump 6272fd098ae8 6272fd098ac5 c4e1315813 c5b15813 c5f877
//
// Vexdump recognizes VEX and EVEX prefix formats.
// Also prints ModR/M byte for convenience (optional).
//
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
)

func progname() string { return path.Base(os.Args[0]) }

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s hexstr...\n", progname())
	os.Exit(1)
}

type prefixKind int

const (
	// 2-byte VEX encoding.
	pVEX2 prefixKind = iota
	// 3-byte VEX encoding.
	pVEX3
	// EVEX encoding of Intel AVX512.
	pEVEX
	// TODO: XOP?
)

// encoding is a parsed instruction octets sequence.
type encoding struct {
	prefix prefixKind
	octets []byte
}

// OctetAt returns octet at i index or 0, if it is out-of-bounds.
// Useful for accessing optional portions of encoding.
func (enc *encoding) OctetAt(i int) byte {
	if len(enc.octets) > i {
		return enc.octets[i]
	}
	return 0
}

// parseArg converts single hex string argument to encoding object.
// Notes:
// - hexstr case does not matter (converted to upper case)
// - spaces are allowed (removed before parsing)
// - hexstr length should not exceed 30
func parseArg(hexstr string) (encoding, error) {
	hexstr = strings.Replace(hexstr, " ", "", -1)
	hexstr = strings.ToUpper(hexstr)

	var enc encoding
	for len(hexstr) >= 2 {
		v, err := strconv.ParseUint(hexstr[:2], 16, 8)
		if err != nil {
			return enc, err
		}
		enc.octets = append(enc.octets, byte(v))
		hexstr = hexstr[2:]
	}

	const ISAinstLenLimit = 15
	if len(enc.octets) > ISAinstLenLimit {
		return enc, errors.New("15 octets are x86 ISA limit")
	}

	switch enc.octets[0] {
	case 0xC5:
		enc.prefix = pVEX2
		if len(enc.octets) < 3 {
			return enc, errors.New("VEX2 requires at least 3 octets")
		}
	case 0xC4:
		enc.prefix = pVEX3
		if len(enc.octets) < 4 {
			return enc, errors.New("VEX3 requires at least 4 octets")
		}
	case 0x62:
		enc.prefix = pEVEX
		if len(enc.octets) < 5 {
			return enc, errors.New("EVEX requires at least 5 octets")
		}
	default:
		return enc, fmt.Errorf("unknown escape byte: %02X", enc.octets[0])
	}

	return enc, nil
}

// parseArgs parses command line arguments into encodings slice.
// Fatal errors lead to program exit,
// while non-fatal errors are only logged without termination.
func parseArgs() []encoding {
	if len(os.Args) < 2 {
		usage()
	}
	var encodings []encoding
	for _, arg := range os.Args[1:] {
		enc, err := parseArg(arg)
		if err != nil {
			log.Printf("%q: %v", arg, err)
			continue
		}
		encodings = append(encodings, enc)
	}

	return encodings
}

// vexFields converts L/pp/mm/W fields to their textual
// representation that follows Intel SDM notation.
func vexFields(ll, pp, mm, w byte) []string {
	var fields []string

	switch ll {
	case 0:
		fields = append(fields, "128")
	case 1:
		fields = append(fields, "256")
	case 2:
		fields = append(fields, "512")
	}
	switch pp {
	case 1:
		fields = append(fields, "66")
	case 2:
		fields = append(fields, "F3")
	case 3:
		fields = append(fields, "F2")
	}
	switch mm {
	case 1:
		fields = append(fields, "0F")
	case 2:
		fields = append(fields, "0F38")
	case 3:
		fields = append(fields, "0F3A")
	}
	switch w {
	case 0:
		fields = append(fields, "W0")
	case 1:
		fields = append(fields, "W1")
	}

	return fields
}

func dumpVEX2(encodings []encoding) {
	printer := newTablePrinter(
		"VEX2", "C5",
		"rvvvvlpp", "%08b",
		"opcode", "%02X",
		"modrm", "%08b",
		"fields", "%s",
	)
	defer printer.Flush()

	printer.PrintHeading()
	for _, enc := range encodings {
		b0 := enc.octets[1]

		fields := []string{"VEX"}
		ll := b0 & (1 << 2) >> 2
		pp := b0 & (3 << 0) >> 0
		mm := byte(1)
		w := byte(0)
		fields = append(fields, vexFields(ll, pp, mm, w)...)

		printer.PrintRow(
			b0,
			enc.octets[2],
			enc.OctetAt(3),
			strings.Join(fields, "."))
	}
}

func dumpVEX3(encodings []encoding) {
	printer := newTablePrinter(
		"VEX3", "C4",
		"rxbmmmmm", "%08b",
		"Wvvvvlpp", "%08b",
		"opcode", "%02X",
		"modrm", "%08b",
		"fields", "%s",
	)
	defer printer.Flush()

	printer.PrintHeading()
	for _, enc := range encodings {
		b0 := enc.octets[1]
		b1 := enc.octets[2]

		fields := []string{"VEX"}
		ll := b1 & (1 << 2) >> 2
		pp := b1 & (3 << 0) >> 0
		mm := b0 & (31 << 0) >> 0
		w := b1 & (1 << 7) >> 7
		fields = append(fields, vexFields(ll, pp, mm, w)...)

		printer.PrintRow(
			b0, b1,
			enc.octets[3],
			enc.OctetAt(4),
			strings.Join(fields, "."))
	}
}

func dumpEVEX(encodings []encoding) {
	printer := newTablePrinter(
		"EVEX", "62",
		"rxbR00mm", "%08b",
		"Wvvvv1pp", "%08b",
		"zLlbVaaa", "%08b",
		"opcode", "%02X",
		"modrm", "%08b",
		"fields", "%s",
	)
	defer printer.Flush()

	printer.PrintHeading()
	for _, enc := range encodings {
		b0 := enc.octets[1]
		b1 := enc.octets[2]
		b2 := enc.octets[3]

		fields := []string{"EVEX"}
		ll := (b2 & (1 << 6) >> 5) | (b2 & (1 << 5) >> 5)
		pp := b1 & (3 << 0) >> 0
		mm := b0 & (3 << 0) >> 0
		w := b1 & (1 << 7) >> 7
		fields = append(fields, vexFields(ll, pp, mm, w)...)

		printer.PrintRow(
			b0, b1, b2,
			enc.octets[4],
			enc.OctetAt(5),
			strings.Join(fields, "."))
	}

}

type tablePrinter struct {
	headingFmt string
	rowFmt     string
	w          *tabwriter.Writer
}

func newTablePrinter(headrow ...string) *tablePrinter {
	var headingParts []string
	var rowParts []string
	for i := 0; i < len(headrow); i += 2 {
		headingParts = append(headingParts, headrow[i+0])
		rowParts = append(rowParts, headrow[i+1])
	}
	return &tablePrinter{
		headingFmt: strings.Join(headingParts, "\t"),
		rowFmt:     strings.Join(rowParts, "\t") + "\n",
		w:          tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0),
	}
}

// PrintHeading writes table heading to stdout.
func (p *tablePrinter) PrintHeading() {
	fmt.Fprintln(p.w, p.headingFmt)
}

// PrintRow uses formatted output to stdout using row format template
// that is bound during tablePrinter object construction.
func (p *tablePrinter) PrintRow(xs ...interface{}) {
	fmt.Fprintf(p.w, p.rowFmt, xs...)
}

// Flush calls internal writer flushing routine.
func (p *tablePrinter) Flush() { p.w.Flush() }

// filterEncodings returns encodings copy which contains only those
// elements that match p prefix kind.
func filterEncodings(p prefixKind, encodings []encoding) []encoding {
	var out []encoding
	for _, enc := range encodings {
		if enc.prefix == p {
			out = append(out, enc)
		}
	}
	return out
}

func main() {
	encodings := parseArgs()

	schemes := [...]struct {
		prefix   prefixKind
		dumpFunc func([]encoding)
	}{
		{pVEX2, dumpVEX2},
		{pVEX3, dumpVEX3},
		{pEVEX, dumpEVEX},
	}

	for _, s := range schemes {
		encodings := filterEncodings(s.prefix, encodings)
		if len(encodings) > 0 {
			s.dumpFunc(encodings)
		}
	}
}
