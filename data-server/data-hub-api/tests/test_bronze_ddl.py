import unittest
from ingestion.bronze_ddl import get_bronze_table_name, TYPE_MAP

class TestBronzeDDL(unittest.TestCase):
    def test_bronze_table_naming(self):
        name = get_bronze_table_name("Core_Banking_Sim", "Account_Transactions")
        self.assertEqual(name, "core_banking_sim_account_transactions")

        name2 = get_bronze_table_name("Client-DB #1", "GL-Entries-2024")
        self.assertEqual(name2, "client_db__1_gl_entries_2024")

    def test_type_mapping(self):
        self.assertEqual(TYPE_MAP["integer"], "BIGINT")
        self.assertEqual(TYPE_MAP["decimal"], "NUMERIC(18,4)")
        self.assertEqual(TYPE_MAP["timestamp"], "TIMESTAMPTZ")
        self.assertEqual(TYPE_MAP["varchar"], "TEXT")
        self.assertEqual(TYPE_MAP["boolean"], "BOOLEAN")

if __name__ == '__main__':
    unittest.main()
