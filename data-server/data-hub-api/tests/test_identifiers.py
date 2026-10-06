import unittest
from core.identifiers import validate_identifier, quote_identifier, quote_table_name

class TestIdentifiers(unittest.TestCase):
    def test_valid_identifiers(self):
        self.assertEqual(validate_identifier("users"), "users")
        self.assertEqual(validate_identifier("account_transactions_2024"), "account_transactions_2024")
        self.assertEqual(validate_identifier("_source_id"), "_source_id")

    def test_sql_injection_rejection(self):
        with self.assertRaises(ValueError):
            validate_identifier("users; DROP TABLE users;--")

        with self.assertRaises(ValueError):
            validate_identifier("table name with spaces")

        with self.assertRaises(ValueError):
            validate_identifier("table' OR '1'='1")

        with self.assertRaises(ValueError):
            validate_identifier("table--comment")

    def test_quote_identifier(self):
        self.assertEqual(quote_identifier("users"), '"users"')
        self.assertEqual(quote_identifier("cb_transactions"), '"cb_transactions"')

    def test_quote_table_name(self):
        self.assertEqual(quote_table_name("users"), '"users"')
        self.assertEqual(quote_table_name("users", schema="bronze"), '"bronze"."users"')

if __name__ == '__main__':
    unittest.main()
