"""
Unit tests for Canonical Entity Mapping and Synonyms auto-detection.
"""
import unittest
import sys
import os

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from mapping.canonical import suggest_canonical_mapping, validate_canonical_mapping, generate_canonical_sql


class TestCanonicalMapping(unittest.TestCase):

    def test_suggest_transactions_indonesian(self):
        cols = ["no_transaksi", "tgl_transaksi", "nominal", "kode_cabang", "kategori", "keterangan"]
        res = suggest_canonical_mapping(cols, "tbl_transaksi_nasabah")
        self.assertEqual(res["entity"], "transactions")
        self.assertGreater(res["confidence"], 0.5)
        self.assertIn("trx_ref_number", res["mappings"])
        self.assertEqual(res["mappings"]["trx_ref_number"], "no_transaksi")
        self.assertEqual(res["mappings"]["trx_date"], "tgl_transaksi")
        self.assertEqual(res["mappings"]["amount"], "nominal")

    def test_suggest_loans_indonesian(self):
        cols = ["no_pinjaman", "baki_debet", "plafon", "suku_bunga", "kolektibilitas", "hari_tunggakan"]
        res = suggest_canonical_mapping(cols, "data_kredit")
        self.assertEqual(res["entity"], "loans")
        self.assertGreater(res["confidence"], 0.5)
        self.assertEqual(res["mappings"]["loan_number"], "no_pinjaman")
        self.assertEqual(res["mappings"]["outstanding"], "baki_debet")

    def test_suggest_generic_fallback(self):
        cols = ["random_foo", "bar_baz", "custom_xyz"]
        res = suggest_canonical_mapping(cols, "custom_table")
        self.assertEqual(res["entity"], "generic")
        self.assertEqual(res["confidence"], 0.0)

    def test_validate_canonical_mapping(self):
        # Valid mapping
        valid_map = {"trx_ref_number": "no_trx", "trx_date": "tgl", "amount": "amt"}
        res = validate_canonical_mapping("transactions", valid_map)
        self.assertTrue(res["valid"])

        # Invalid mapping (missing required amount)
        invalid_map = {"trx_ref_number": "no_trx", "trx_date": "tgl"}
        res2 = validate_canonical_mapping("transactions", invalid_map)
        self.assertFalse(res2["valid"])
        self.assertIn("amount", res2["missing_fields"])

    def test_generate_canonical_sql(self):
        mapping = {
            "trx_ref_number": "no_trx",
            "trx_date": "tgl_trx",
            "amount": "nilai_trx",
            "category": "tipe",
            "channel": "kanal"
        }
        sql = generate_canonical_sql("src-01", "ClientBank", "transactions", "transactions", mapping)
        self.assertIsNotNone(sql)
        self.assertIn("INSERT INTO silver.trx_cleaned", sql)
        self.assertIn("src_clientbank_transactions", sql)
        self.assertIn("ON CONFLICT (source_id, trx_ref_number)", sql)


if __name__ == "__main__":
    unittest.main()
