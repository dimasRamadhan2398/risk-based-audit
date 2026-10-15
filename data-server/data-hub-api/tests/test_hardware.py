import os
import unittest
from core.hardware import get_hardware_profile

class TestHardware(unittest.TestCase):
    def test_hardware_profile_demo(self):
        os.environ["HARDWARE_PROFILE"] = "demo"
        profile = get_hardware_profile()
        self.assertEqual(profile["profile"], "demo")
        self.assertEqual(profile["chunk_size"], 50000)
        self.assertEqual(profile["max_parallel_tables"], 1)
        self.assertEqual(profile["max_connections_per_source"], 2)

    def test_hardware_profile_prod(self):
        os.environ["HARDWARE_PROFILE"] = "prod"
        profile = get_hardware_profile()
        self.assertEqual(profile["profile"], "prod")
        self.assertEqual(profile["chunk_size"], 100000)
        self.assertEqual(profile["max_parallel_tables"], 3)
        self.assertEqual(profile["max_connections_per_source"], 4)

if __name__ == '__main__':
    unittest.main()
