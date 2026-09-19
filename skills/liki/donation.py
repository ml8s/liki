#!/usr/bin/env python3
"""Whole-file donation receipt helper for the Liki skill.

The helper only talks to `/api/donations/redeem`. It never grants features and
never stores a Payment-Proof locally.
"""

from __future__ import annotations

import argparse
import base64
import json
import os
import pathlib
import sys
import urllib.error
import urllib.request


SCHEMA_VERSION = "liki-donation-v1"
DEFAULT_ENDPOINT = "https://liki.hk/api/donations/redeem"
DEFAULT_CREDENTIAL = pathlib.Path("~/.liki/donation.json").expanduser()
TIMEOUT_SECONDS = 10
MAX_RESPONSE_BYTES = 64 * 1024
MAX_PROOF_BYTES = 64 * 1024


class DonationError(ValueError):
    pass


class _Response:
    """Small transport-independent response used by tests and HTTPError."""

    def __init__(self, status: int, body: bytes, headers=None):
        self.status = status
        self.body = body
        self.headers = headers

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, traceback):
        return False

    def read(self):
        return self.body


def _post(endpoint: str, proof: str = "") -> _Response:
    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json; charset=utf-8",
    }
    if proof:
        headers["Payment-Proof"] = proof
    request = urllib.request.Request(endpoint, data=b"", headers=headers, method="POST")
    try:
        with urllib.request.urlopen(request, timeout=TIMEOUT_SECONDS) as response:
            return _Response(response.status, response.read(MAX_RESPONSE_BYTES + 1), response.headers)
    except urllib.error.HTTPError as exc:
        return _Response(exc.code, exc.read(MAX_RESPONSE_BYTES + 1), exc.headers)


def validate_receipt(value) -> dict:
    if not isinstance(value, dict):
        raise DonationError("credential must be an object")
    if value.get("schema_version") != SCHEMA_VERSION:
        raise DonationError("schema_version is invalid")
    if value.get("kind") != "donation-receipt":
        raise DonationError("kind is invalid")
    if value.get("status") != "received":
        raise DonationError("status is invalid")
    for key in ("receipt_id", "amount", "currency"):
        if not isinstance(value.get(key), str) or not value[key].strip():
            raise DonationError(f"{key} is invalid")
    return value


def read_receipt(path: pathlib.Path) -> tuple[bool, dict | None, str]:
    try:
        raw = path.read_bytes()
        if len(raw) > MAX_RESPONSE_BYTES:
            raise DonationError("credential is too large")
        receipt = validate_receipt(json.loads(raw))
        return True, receipt, "valid"
    except FileNotFoundError:
        return False, None, "not_found"
    except Exception as exc:  # noqa: BLE001 - status must remain machine-readable
        return False, None, type(exc).__name__


def request_bill(endpoint: str) -> dict:
    response = _post(endpoint, "")
    if response.status != 402:
        raise DonationError(f"expected HTTP 402, got HTTP {response.status}")
    raw_header = _header(response.headers, "Payment-Needed")
    if not raw_header:
        raise DonationError("Payment-Needed header is missing")
    if len(raw_header) > MAX_RESPONSE_BYTES:
        raise DonationError("Payment-Needed header is too large")
    try:
        payment_needed = json.loads(_decode_base64url(raw_header))
    except (ValueError, TypeError) as exc:
        raise DonationError("Payment-Needed header is not valid JSON") from exc
    if not isinstance(payment_needed, dict) or not isinstance(payment_needed.get("protocol"), dict):
        raise DonationError("Payment-Needed protocol is missing")
    protocol = payment_needed["protocol"]
    return {
        "ok": True,
        "stage": "payment_needed",
        "endpoint": endpoint,
        "out_trade_no": protocol.get("out_trade_no"),
        "amount": protocol.get("amount"),
        "currency": protocol.get("currency"),
        "payment_needed": payment_needed,
        "payment_needed_encoded": raw_header,
        "message": "Use Alipay AI Pay with this bill, then confirm with the resulting Payment-Proof.",
    }


def confirm_receipt(endpoint: str, proof: str, credential_path: pathlib.Path) -> dict:
    if not proof or not proof.strip():
        raise DonationError("Payment-Proof is required")
    if len(proof) > MAX_PROOF_BYTES:
        raise DonationError("Payment-Proof is too large")
    response = _post(endpoint, proof.strip())
    if len(response.body) > MAX_RESPONSE_BYTES:
        raise DonationError("response is too large")
    try:
        body = json.loads(response.body)
    except json.JSONDecodeError as exc:
        raise DonationError(f"HTTP {response.status} returned invalid JSON") from exc
    if response.status != 200:
        raise DonationError(f"expected HTTP 200, got HTTP {response.status}")
    if not isinstance(body, dict) or body.get("resource_id") != "/api/donations/redeem":
        raise DonationError("resource_id is invalid")
    if body.get("fulfillment_confirmed") is not True:
        raise DonationError("fulfillment is not confirmed")
    receipt = validate_receipt(body.get("content"))
    credential_path.parent.mkdir(parents=True, exist_ok=True)
    temp_path = credential_path.with_name(credential_path.name + ".tmp")
    temp_path.write_text(
        json.dumps(receipt, ensure_ascii=False, separators=(",", ":")) + "\n",
        encoding="utf-8",
    )
    os.replace(temp_path, credential_path)
    try:
        credential_path.chmod(0o600)
    except OSError:
        pass
    return {
        "ok": True,
        "stage": "received",
        "endpoint": endpoint,
        "credential_path": str(credential_path),
        "receipt": receipt,
        "fulfillment_confirmed": True,
    }


def _header(headers, name: str) -> str:
    if headers is None:
        return ""
    try:
        return headers.get(name) or headers.get(name.lower()) or ""
    except AttributeError:
        lower = name.lower()
        for key, value in headers.items():
            if key.lower() == lower:
                return value
    return ""


def _decode_base64url(value: str) -> bytes:
    value = value.strip()
    if len(value) % 4:
        value += "=" * (4 - len(value) % 4)
    return base64.urlsafe_b64decode(value.encode("ascii"))


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)

    commands.add_parser("status", help="check whether a valid receipt exists")

    request_parser = commands.add_parser("request", help="request a Payment-Needed bill")
    request_parser.add_argument("--endpoint", default=DEFAULT_ENDPOINT)

    confirm_parser = commands.add_parser("confirm", help="redeem a Payment-Proof")
    confirm_parser.add_argument("--endpoint", default=DEFAULT_ENDPOINT)
    confirm_parser.add_argument("--proof-file", help="file containing only the Payment-Proof")

    args = parser.parse_args(argv)
    endpoint = os.environ.get("LIKI_DONATION_URL", args.endpoint)
    try:
        if args.command == "status":
            donated, receipt, reason = read_receipt(DEFAULT_CREDENTIAL)
            result = {"ok": True, "donated": donated, "reason": reason}
            if donated:
                result["receipt"] = receipt
            print(json.dumps(result, ensure_ascii=True, separators=(",", ":")))
            return 0
        if args.command == "request":
            print(json.dumps(request_bill(endpoint), ensure_ascii=True, separators=(",", ":")))
            return 0
        if args.command == "confirm":
            if args.proof_file:
                proof = pathlib.Path(args.proof_file).read_text(encoding="utf-8").strip()
            else:
                proof = sys.stdin.read().strip()
            print(json.dumps(confirm_receipt(endpoint, proof, DEFAULT_CREDENTIAL),
                             ensure_ascii=True, separators=(",", ":")))
            return 0
    except Exception as exc:  # noqa: BLE001 - CLI returns a JSON envelope
        print(json.dumps({"ok": False, "stage": args.command, "error": str(exc)},
                         ensure_ascii=True, separators=(",", ":")))
        return 1
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
