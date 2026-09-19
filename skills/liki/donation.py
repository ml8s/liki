#!/usr/bin/env python3
"""Donation credential helper for the Liki skill.

Thin layer: the LLM handles all HTTP communication with the donation
endpoint (A2M 402 protocol). This tool only does what the LLM cannot
do safely: validate receipt structure and atomically persist credentials.
"""

from __future__ import annotations

import argparse
import json
import os
import pathlib
import sys

SCHEMA_VERSION = "liki-donation-v1"
DEFAULT_CREDENTIAL = pathlib.Path("~/.liki/donation.json").expanduser()
MAX_RECEIPT_BYTES = 64 * 1024


class DonationError(ValueError):
    pass


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
        if len(raw) > MAX_RECEIPT_BYTES:
            raise DonationError("credential is too large")
        receipt = validate_receipt(json.loads(raw))
        return True, receipt, "valid"
    except FileNotFoundError:
        return False, None, "not_found"
    except Exception as exc:  # noqa: BLE001
        return False, None, type(exc).__name__


def save_receipt(raw: str, credential_path: pathlib.Path) -> dict:
    if not raw or not raw.strip():
        raise DonationError("receipt JSON is required")
    if len(raw.encode("utf-8")) > MAX_RECEIPT_BYTES:
        raise DonationError("receipt is too large")
    try:
        body = json.loads(raw)
    except json.JSONDecodeError as exc:
        raise DonationError(f"receipt is not valid JSON: {exc}") from exc
    receipt = validate_receipt(body)
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
        "credential_path": str(credential_path),
        "receipt": receipt,
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    commands.add_parser("status", help="check whether a valid receipt exists")
    save_parser = commands.add_parser("save-receipt", help="validate and persist a donation receipt from stdin")
    save_parser.add_argument("--path", default=None, help="override credential path (for testing)")
    args = parser.parse_args(argv)
    credential = pathlib.Path(args.path).expanduser() if args.path else DEFAULT_CREDENTIAL
    try:
        if args.command == "status":
            donated, receipt, reason = read_receipt(credential)
            result = {"ok": True, "donated": donated, "reason": reason}
            if donated:
                result["receipt"] = receipt
            print(json.dumps(result, ensure_ascii=True, separators=(",", ":")))
            return 0
        if args.command == "save-receipt":
            raw = sys.stdin.read()
            result = save_receipt(raw, credential)
            print(json.dumps(result, ensure_ascii=True, separators=(",", ":")))
            return 0
    except Exception as exc:  # noqa: BLE001
        print(json.dumps({"ok": False, "stage": args.command, "error": str(exc)},
                         ensure_ascii=True, separators=(",", ":")))
        return 1
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
