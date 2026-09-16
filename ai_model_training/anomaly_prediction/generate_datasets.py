import os
import sys

from generate_anomaly_dataset import generate as gen_anomaly
from generate_access_pattern_dataset import generate as gen_access
from generate_audit_budget_dataset import generate as gen_budget
from generate_expense_report_dataset import generate as gen_expense
from generate_fieldwork_dataset import generate as gen_fieldwork
from generate_inventory_dataset import generate as gen_inventory
from generate_mitigation_overdue_dataset import generate as gen_mitigation
from generate_repeat_finding_dataset import generate as gen_repeat
from generate_risk_score_spike_dataset import generate as gen_spike

def main():
    print("Mulai menghasilkan semua dataset...")
    
    gen_anomaly()
    gen_access()
    gen_budget()
    gen_expense()
    gen_fieldwork()
    gen_inventory()
    gen_mitigation()
    gen_repeat()
    gen_spike()
    
    print("\n[SUCCESS] ALL 9 SYNTHETIC DATASETS GENERATED SUCCESSFULLY WITH REALISTIC DOMAIN LOGIC!")

if __name__ == "__main__":
    main()
