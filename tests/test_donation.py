from __future__ import annotations

import base64
import importlib.util
import json
import pathlib
import tempfile
import unittest
from unittest import mock


def _load_donation():
    path = pathlib.Path(__file__).parents[1] / "skills" / "liki" / "donation.py"
    spec = importlib.util.spec_from_file_location("liki_donation", path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


donation = _load_donation()


def encoded_bill() -> str:
    bill = {
        "protocol": {
            "out_trade_no": "order-1",
            "amount": "1.00",
            "currency": "CNY",
            "resource_id": "/api/donations/redeem",
        }
    }
    raw = json.dumps(bill, ensure_ascii=False).encode("utf-8")
    return base64.urlsafe_b64encode(raw).rstrip(b"=").decode("ascii")


class FakeResponse:
    def __init__(self, status: int, body: bytes, headers: dict[str, str] | None = None):
        self.status = status
        self.body = body
        self.headers = {key.lower(): value for key, value in (headers or {}).items()}

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, traceback):
        return False

    def read(self):
        return self.body


class DonationReceiptValidationTests(unittest.TestCase):
    def test_valid_receipt_is_recognized(self):
        receipt = {
            "schema_version": "liki-donation-v1",
            "kind": "donation-receipt",
            "status": "received",
            "receipt_id": "r-1",
            "amount": "1.00",
            "currency": "CNY",
        }
        self.assertEqual(donation.validate_receipt(receipt), receipt)

    def test_invalid_receipts_are_rejected(self):
        receipts = [
            None,
            "receipt",
            {},
            {"schema_version": "old", "kind": "donation-receipt", "status": "received"},
            {"schema_version": "liki-donation-v1", "kind": "token", "status": "received"},
            {"schema_version": "liki-donation-v1", "kind": "donation-receipt", "status": "pending"},
            {"schema_version": "liki-donation-v1", "kind": "donation-receipt", "status": "received", "receipt_id": ""},
        ]
        for receipt in receipts:
            with self.subTest(receipt=receipt):
                with self.assertRaises(donation.DonationError):
                    donation.validate_receipt(receipt)


class DonationRequestTests(unittest.TestCase):
    def test_request_returns_payment_needed_without_writing_credential(self):
        response = FakeResponse(402, b'{"code":"Payment-Needed"}', {"Payment-Needed": encoded_bill()})
        with mock.patch.object(donation, "_post", return_value=response) as poster:
            result = donation.request_bill("https://liki.example/api/donations/redeem")
        poster.assert_called_once_with("https://liki.example/api/donations/redeem", "")
        self.assertTrue(result["ok"])
        self.assertEqual(result["stage"], "payment_needed")
        self.assertEqual(result["payment_needed"]["protocol"]["out_trade_no"], "order-1")
        self.assertEqual(result["amount"], "1.00")
        self.assertIn("payment_needed_encoded", result)

    def test_request_rejects_missing_bill_header(self):
        response = FakeResponse(402, b"{}")
        with mock.patch.object(donation, "_post", return_value=response):
            with self.assertRaises(donation.DonationError):
                donation.request_bill("https://liki.example/api/donations/redeem")

    def test_request_rejects_oversized_bill_header(self):
        oversized = "x" * (donation.MAX_RESPONSE_BYTES + 1)
        response = FakeResponse(402, b"{}", {"Payment-Needed": oversized})
        with mock.patch.object(donation, "_post", return_value=response):
            with self.assertRaises(donation.DonationError):
                donation.request_bill("https://liki.example/api/donations/redeem")


class DonationConfirmTests(unittest.TestCase):
    def test_confirm_writes_whole_content_object(self):
        receipt = {
            "schema_version": "liki-donation-v1",
            "kind": "donation-receipt",
            "status": "received",
            "receipt_id": "r-1",
            "amount": "1.00",
            "currency": "CNY",
        }
        body = {
            "resource_id": "/api/donations/redeem",
            "content": receipt,
            "credential_path": "~/.liki/donation.json",
            "fulfillment_confirmed": True,
        }
        response = FakeResponse(200, json.dumps(body).encode("utf-8"))
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            with mock.patch.object(donation, "_post", return_value=response) as poster:
                result = donation.confirm_receipt(
                    "https://liki.example/api/donations/redeem",
                    "proof-from-alipay",
                    path,
                )
            poster.assert_called_once_with(
                "https://liki.example/api/donations/redeem",
                "proof-from-alipay",
            )
            self.assertTrue(result["ok"])
            self.assertTrue(result["fulfillment_confirmed"])
            self.assertEqual(json.loads(path.read_text(encoding="utf-8")), receipt)

    def test_confirm_does_not_write_invalid_receipt(self):
        body = {"resource_id": "/api/donations/redeem", "content": {"kind": "token"}}
        response = FakeResponse(200, json.dumps(body).encode("utf-8"))
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            with mock.patch.object(donation, "_post", return_value=response):
                with self.assertRaises(donation.DonationError):
                    donation.confirm_receipt(
                        "https://liki.example/api/donations/redeem",
                        "proof",
                        path,
                    )
            self.assertFalse(path.exists())


if __name__ == "__main__":
    unittest.main()
