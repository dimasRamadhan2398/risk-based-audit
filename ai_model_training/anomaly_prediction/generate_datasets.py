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
from generate_it_control_dataset import generate as gen_it_control
from generate_kyc_dataset import generate as gen_kyc
from generate_payment_dataset import generate as gen_payment
from generate_treasury_dataset import generate as gen_treasury
from generate_funding_dataset import generate as gen_funding
from generate_lending_dataset import generate as gen_lending
from generate_bank_transaction_dataset import generate as gen_bank_trx

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
    gen_it_control()
    gen_kyc()
    gen_payment()
    gen_treasury()
    gen_funding()
    gen_lending()
    gen_bank_trx()
    
    print("\n[SUCCESS] ALL SYNTHETIC DATASETS INCLUDING IT CONTROL, KYC, PAYMENT, TREASURY, FUNDING, LENDING & AGGREGATE BANK TRANSACTIONS GENERATED SUCCESSFULLY WITH REALISTIC DOMAIN LOGIC!")

if __name__ == "__main__":
    main()
