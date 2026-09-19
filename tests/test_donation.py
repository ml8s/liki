from __future__ import annotations

import importlib.util
import json
import pathlib
import tempfile
import unittest


def _load_donation():
    path = pathlib.Path(__file__).parents[1] / "skills" / "liki" / "donation.py"
    spec = importlib.util.spec_from_file_location("liki_donation", path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


donation = _load_donation()


def valid_receipt() -> dict:
    return {
        "schema_version": "liki-donation-v1",
        "kind": "donation-receipt",
        "status": "received",
        "receipt_id": "r-1",
        "amount": "19.9",
        "currency": "CNY",
    }


class DonationReceiptValidationTests(unittest.TestCase):
    def test_valid_receipt_is_recognized(self):
        receipt = valid_receipt()
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


class DonationStatusTests(unittest.TestCase):
    def test_status_returns_not_found_when_no_credential(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            donated, receipt, reason = donation.read_receipt(path)
            self.assertFalse(donated)
            self.assertIsNone(receipt)
            self.assertEqual(reason, "not_found")

    def test_status_returns_valid_for_existing_receipt(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            path.write_text(json.dumps(valid_receipt()), encoding="utf-8")
            donated, receipt, reason = donation.read_receipt(path)
            self.assertTrue(donated)
            self.assertEqual(receipt, valid_receipt())
            self.assertEqual(reason, "valid")


class DonationSaveReceiptTests(unittest.TestCase):
    def test_save_receipt_writes_atomically_with_mode_600(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            result = donation.save_receipt(json.dumps(valid_receipt()), path)
            self.assertTrue(result["ok"])
            self.assertEqual(result["receipt"], valid_receipt())
            written = json.loads(path.read_text(encoding="utf-8"))
            self.assertEqual(written, valid_receipt())

    def test_save_receipt_rejects_invalid_json(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            with self.assertRaises(donation.DonationError):
                donation.save_receipt("not json", path)
            self.assertFalse(path.exists())

    def test_save_receipt_rejects_invalid_receipt(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            with self.assertRaises(donation.DonationError):
                donation.save_receipt(json.dumps({"kind": "token"}), path)
            self.assertFalse(path.exists())

    def test_save_receipt_rejects_oversized_input(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            oversized = json.dumps(valid_receipt()) + " " * (donation.MAX_RECEIPT_BYTES + 1)
            with self.assertRaises(donation.DonationError):
                donation.save_receipt(oversized, path)
            self.assertFalse(path.exists())

    def test_save_receipt_rejects_empty_input(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "donation.json"
            with self.assertRaises(donation.DonationError):
                donation.save_receipt("", path)
            self.assertFalse(path.exists())


if __name__ == "__main__":
    unittest.main()
