#!/usr/bin/env -S uv run --script
#
# /// script
# dependencies = [
#   "tqdm",
#   "pyshark",
# ]
# ///

import argparse
from dataclasses import dataclass
from pathlib import Path
from string import Template
from typing import Iterable, List, Union

import pyshark
from tqdm import tqdm

Files = Iterable[Union[str, Path]]


def to_go_literal(raw: bytes) -> str:
    """Convert bytes to a Go literal string."""
    return "[]byte{" + ", ".join(f"0x{b:02x}" for b in raw) + "}"


@dataclass
class JA4XRecord:
    file: str
    index: int
    cert: bytes
    ja4x: str

    def go(self) -> str:
        payload = to_go_literal(self.cert)
        return f"""{{
            File:        "{self.file}",
            Index:       {self.index},
            Cert:        {payload},
            Fingerprint: "{self.ja4x}",
}},"""


@dataclass
class JA4SRecord:
    file: str
    index: int
    hello: bytes
    ja4s: str

    def go(self) -> str:
        payload = to_go_literal(self.hello)
        return f"""{{
            File:        "{self.file}",
            Index:       {self.index},
            Hello:       {payload},
            Protocol:    "t",
            Fingerprint: "{self.ja4s}",
}},"""


def ja4x_records(file) -> List[JA4XRecord]:
    records: List[JA4XRecord] = []
    cap = pyshark.FileCapture(file, display_filter="tls.handshake")

    for pkt in cap:
        if "tls" in pkt:
            tls = pkt.tls

            if not hasattr(tls, "handshake_certificate"):
                continue
            if not hasattr(tls, "ja4_ja4x"):
                continue

            records.append(
                JA4XRecord(
                    file=file.name,
                    index=pkt.number,
                    cert=bytes.fromhex(tls.handshake_certificate.replace(":", "")),
                    ja4x=tls.ja4_ja4x,
                )
            )

    cap.close()
    return records


def ja4s_records(file) -> List[JA4SRecord]:
    records: List[JA4SRecord] = []
    cap = pyshark.FileCapture(
        file,
        display_filter="tls.handshake.type == 2",
        include_raw=True,
        use_ek=True,
    )

    for pkt in cap:
        # print(pkt.tls)
        if "tls" in pkt:
            tls = pkt.tls
            if not hasattr(tls, "ja4"):
                continue

            hello = tls.handshake.raw
            if isinstance(tls.handshake.raw, list):
                for r in tls.handshake.raw:
                    if r.startswith("02"):
                        hello = r
                        break

            records.append(
                JA4SRecord(
                    file=file.name,
                    index=pkt.number,
                    hello=bytes.fromhex(hello),
                    ja4s=tls.ja4["ja4_ja4_ja4s"],
                )
            )

    cap.close()
    return records


def ja4_records(file) -> List[JA4SRecord]:
    records: List[JA4SRecord] = []
    cap = pyshark.FileCapture(
        file,
        display_filter="tls.handshake.type == 1",
        include_raw=True,
        use_ek=True,
    )

    for pkt in cap:
        # print(pkt.tls)
        if "tls" in pkt:
            tls = pkt.tls
            handshake = tls.handshake

            if not hasattr(handshake, "ja4"):
                continue

            print("JA4", handshake.ja4.value)

            hello = tls.handshake.raw
            if isinstance(tls.handshake.raw, list):
                for r in tls.handshake.raw:
                    if r.startswith("01"):
                        hello = r
                        break

            records.append(
                JA4SRecord(
                    file=file.name,
                    index=pkt.number,
                    hello=bytes.fromhex(hello),
                    ja4s=tls.handshake.ja4.value,
                )
            )

    cap.close()
    return records


JA4X_TEMPLATE = Template(
    """package ja4

import (
	"crypto/x509"
	"fmt"
	"testing"
)

type JA4XRecord struct {
	File        string
	Index       int
	Cert        []byte
	Fingerprint string
}

func mustParseCertificate(raw []byte) *x509.Certificate {
	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		panic(err)
	}
	return cert
}

func TestJA4X(t *testing.T) {
	for _, record := range ja4xRecords {
		name := fmt.Sprintf("%s - %d", record.File, record.Index)
		t.Run(name, func(t *testing.T) {
            cert := mustParseCertificate(record.Cert)
			fingerprint := JA4X(cert)

			if fingerprint != record.Fingerprint {
				t.Errorf("Test %s failed: expected JA4X=%v, got=%v", name, record.Fingerprint, fingerprint)
			}
		})
	}
}

var ja4xRecords = []JA4XRecord{
    $records
}
"""
)

JA4S_TEMPLATE = Template(
    """package ja4

import (
	"fmt"
	"testing"
)

type JA4SRecord struct {
	File        string
	Index       int
	Protocol    string
	Hello       []byte
	Fingerprint string
}

func TestJA4S(t *testing.T) {
	for _, record := range ja4sRecords {
		name := fmt.Sprintf("%s - %d", record.File, record.Index)
		t.Run(name, func(t *testing.T) {
			fingerprint, err := JA4S(record.Hello)

			if fingerprint != record.Fingerprint {
				t.Errorf("Test %s failed: expected JA4S=%v, got=%v", name, record.Fingerprint, fingerprint)
			}

			if err != nil && record.Fingerprint != "" {
				t.Errorf("Test %s failed: unexpected error: %v", name, err)
			}
		})
	}
}

var ja4sRecords = []JA4SRecord{
    $records
}
"""
)


JA4_TEMPLATE = Template(
    """package ja4

import (
	"fmt"
	"testing"
)

type JA4Record struct {
	File        string
	Index       int
	Protocol    string
	Hello       []byte
	Fingerprint string
}

func TestJA4(t *testing.T) {
	for _, record := range ja4Records {
		name := fmt.Sprintf("%s - %d", record.File, record.Index)
		t.Run(name, func(t *testing.T) {
			fingerprint, err := JA4(record.Hello)

			if fingerprint != record.Fingerprint {
				t.Errorf("Test %s failed: expected JA4=%v, got=%v", name, record.Fingerprint, fingerprint)
			}

			if err != nil && record.Fingerprint != "" {
				t.Errorf("Test %s failed: unexpected error: %v", name, err)
			}
		})
	}
}

var ja4Records = []JA4Record{
    $records
}
"""
)


def ja4x_test_file(output_file: str | Path, input_files: Files):
    records = []
    for file in tqdm(input_files):
        records += ja4x_records(file)
    records_str = "\n".join(record.go() for record in records)
    output = JA4X_TEMPLATE.substitute(records=records_str)
    with open(output_file, "w") as f:
        f.write(output)


def ja4s_test_file(output_file: str | Path, input_files: Files):
    records = []
    for file in tqdm(input_files):
        records += ja4s_records(file)
    records_str = "\n".join(record.go() for record in records)
    output = JA4S_TEMPLATE.substitute(records=records_str)
    with open(output_file, "w") as f:
        f.write(output)


def ja4_test_file(output_file: str | Path, input_files: Files):
    records = []
    for file in tqdm(input_files):
        records += ja4_records(file)
    records_str = "\n".join(record.go() for record in records)
    output = JA4_TEMPLATE.substitute(records=records_str)
    with open(output_file, "w") as f:
        f.write(output)


def error(msg: str):
    print(f"\033[91m{msg}\033[0m")


def success(msg: str):
    print(f"\033[92m{msg}\033[0m")


DESCRIPTION = """Generate JA4X and JA4S test files.
Beforehand you must have wireshark/tshark installed along with the ja4 plugin (from releases: https://github.com/FoxIO-LLC/ja4/releases).
See https://github.com/FoxIO-LLC/ja4/tree/main/wireshark#installing-the-plugin.
"""


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=DESCRIPTION)
    parser.add_argument(
        "--ja4",
        type=Path,
        required=False,
        help="Output file for JA4 tests",
    )
    parser.add_argument(
        "--ja4s",
        type=Path,
        required=False,
        help="Output file for JA4S tests",
    )
    parser.add_argument(
        "--ja4x",
        type=Path,
        required=False,
        help="Output file for JA4X tests",
    )
    parser.add_argument(
        "pcap_files",
        nargs="+",
        help="Input PCAP files",
    )

    args = parser.parse_args()
    if args.ja4x is None and args.ja4s is None and args.ja4 is None:
        error("Please specify at least one output file with --ja4, --ja4s or --ja4x")
        exit(1)

    pcaps = [Path(f) for f in args.pcap_files]
    if len(pcaps) == 0:
        error("No PCAP files provided. Using default PCAP files.")
        exit(2)

    print(f"{len(pcaps)} PCAP files provided")
    if args.ja4x:
        ja4x_test_file(args.ja4x, [Path(f) for f in args.pcap_files])
        success(f"JA4X test file generated: {args.ja4x}")
    if args.ja4s:
        ja4s_test_file(args.ja4s, [Path(f) for f in args.pcap_files])
        success(f"JA4S test file generated: {args.ja4s}")
    if args.ja4:
        ja4_test_file(args.ja4, [Path(f) for f in args.pcap_files])
        success(f"JA4 test file generated: {args.ja4}")
