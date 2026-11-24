import os
import sys

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
if CURRENT_DIR not in sys.path:
    sys.path.append(CURRENT_DIR)

from main import predict_anomaly, get_anomaly_batch, normalize_anomaly_risk_level, AnomalyRequest

ALLOWED_RISK_CODES = {'HIGH', 'MODERATE_HIGH', 'MODERATE', 'LOW_MODERATE', 'LOW'}
ALLOWED_RISK_LABELS = {'High', 'Moderate to High', 'Moderate', 'Low to Moderate', 'Low'}

def run_tests():
    print("=" * 60)
    print("TESTING FASTAPI ANOMALY FUNCTIONS DIRECTLY")
    print("=" * 60)

    # 1. Legacy Operational Request (Backward Compatibility)
    legacy_req = AnomalyRequest(
        entity="Jakarta Branch",
        description="Pengeluaran Klaim Operasional Melebihi Batas Toleransi",
        amount=450.0,
        hour_of_day=23,
        day_of_week=6,
        is_new_beneficiary=1,
        is_round_amount=1
    )
    leg_data = predict_anomaly(legacy_req)
    print("[PASS] predict_anomaly (Legacy Operational):")
    print(f"       type: {leg_data.get('type')}, anomaly_score: {leg_data.get('anomaly_score')}, is_anomaly: {leg_data.get('is_anomaly')}, severity: {leg_data.get('severity')}")
    
    required_keys = ['id', 'entity', 'type', 'anomaly_score', 'description', 'severity', 'amount', 'is_anomaly', 'predicted_impact', 'predicted_likelihood', 'risk_level', 'xMetric', 'hour', 'hour_of_day']
    for k in required_keys:
        assert k in leg_data, f"Missing key in response: {k}"

    # 2. New Banking Category Request (Funding)
    bank_req_funding = AnomalyRequest(
        entity="Cabang Jakarta",
        category="Funding",
        channel="Cabang / Teller",
        description="Pemberian suku bunga deposito di atas batas penjaminan LPS tanpa persetujuan ALCO",
        amount=12500.0,
        interest_margin=8.5,
        hour_of_day=14,
        document_status="Tidak Lengkap",
        failed_logins=0,
        customer_risk_profile="PEP / High Risk Watchlist",
        authorization_status="Override / Tanpa Otorisasi",
        historical_dev_pct=145.0
    )
    bf_data = predict_anomaly(bank_req_funding)
    print("[PASS] predict_anomaly (Bank Funding):")
    print(f"       type: {bf_data.get('type')}, anomaly_score: {bf_data.get('anomaly_score')}, is_anomaly: {bf_data.get('is_anomaly')}, severity: {bf_data.get('severity')}, xMetric: {bf_data.get('xMetric')}")
    assert bf_data.get('type') == 'Funding'
    for k in required_keys:
        assert k in bf_data, f"Missing key in response: {k}"

    # 3. New Banking Category Request (IT Control)
    bank_req_it = AnomalyRequest(
        entity="Cabang Medan",
        category="IT Control",
        channel="Host-to-Host Core Banking",
        description="Percobaan login berulang kali dan akses basis data pada dini hari",
        amount=0.0,
        interest_margin=0.0,
        hour_of_day=2,
        document_status="Lengkap",
        failed_logins=7,
        customer_risk_profile="High Risk",
        authorization_status="Override / Tanpa Otorisasi",
        historical_dev_pct=210.0
    )
    bit_data = predict_anomaly(bank_req_it)
    print("[PASS] predict_anomaly (Bank IT Control):")
    print(f"       type: {bit_data.get('type')}, anomaly_score: {bit_data.get('anomaly_score')}, is_anomaly: {bit_data.get('is_anomaly')}, xMetric: {bit_data.get('xMetric')}")
    assert bit_data.get('type') == 'IT Control'
    assert bit_data.get('xMetric') == 7.0

    # 4. Batch Anomaly Function
    batch_data = get_anomaly_batch()
    anomalies = batch_data.get('anomalies', [])
    scatter = batch_data.get('scatter_data', [])
    summary = batch_data.get('summary', {})

    print(f"[PASS] get_anomaly_batch():")
    print(f"       Total anomalies: {len(anomalies)}")
    print(f"       Scatter points: {len(scatter)}")
    print(f"       Summary: {summary}")

    # Check categories covered
    detected_cats = {a.get('type') for a in anomalies}
    print(f"       Categories detected in batch: {detected_cats}")
    for cat in ['Funding', 'Lending', 'Treasury', 'Payment', 'KYC', 'IT Control', 'Audit Findings']:
        assert cat in detected_cats, f"Expected category {cat} in batch anomalies, got {detected_cats}"

    af_anoms = [a for a in anomalies if a.get('type') == 'Audit Findings']
    af_scatter = [s for s in scatter if s.get('type') == 'Audit Findings']
    assert len(af_anoms) > 0, "Expected at least 1 Audit Findings anomaly from Action Taken Report"
    assert len([s for s in af_scatter if not s.get('is_anomaly')]) > 0, "Expected normal baseline scatter points for Audit Findings"
    print(f"[PASS] Action Taken Report -> Audit Findings: {len(af_anoms)} anomalies, {len(af_scatter)} scatter points")

    # Verify Y coordinates are in bounds (0.0 - 24.0) and no synthetic 25+ clipping values
    for pt in scatter:
        assert 0.0 <= pt['y'] <= 24.0, f"Scatter point y value out of bounds (0-24): {pt}"

    # Verify each anomaly has a valid hour and strictly uses the 5 allowed risk levels (no Critical)
    for anom in anomalies:
        assert 'hour' in anom and anom['hour'] is not None, f"Anomaly missing hour: {anom['id']}"
        assert 0 <= anom['hour'] <= 23, f"Anomaly hour out of bounds (0-23): {anom['hour']}"
        assert anom.get('risk_level') in ALLOWED_RISK_CODES, f"Invalid risk_level: {anom.get('risk_level')}"
        assert anom.get('severity') in ALLOWED_RISK_LABELS, f"Invalid severity: {anom.get('severity')}"

    # 5. Verify External Database & Critical Risk Level Normalization
    for single_pred in [leg_data, bf_data, bit_data]:
        assert single_pred.get('risk_level') in ALLOWED_RISK_CODES, f"Invalid single risk_level: {single_pred.get('risk_level')}"
        assert single_pred.get('severity') in ALLOWED_RISK_LABELS, f"Invalid single severity: {single_pred.get('severity')}"

    ext_cases = [
        ("Critical", "HIGH", "High"),
        ("CRITICAL", "HIGH", "High"),
        ("Extreme", "HIGH", "High"),
        ("Very High", "HIGH", "High"),
        ("Elevated", "MODERATE_HIGH", "Moderate to High"),
        ("Moderate to High", "MODERATE_HIGH", "Moderate to High"),
        ("Moderate", "MODERATE", "Moderate"),
        ("Minor", "LOW_MODERATE", "Low to Moderate"),
        ("Low to Moderate", "LOW_MODERATE", "Low to Moderate"),
        ("Insignificant", "LOW", "Low"),
        ("Very Low", "LOW", "Low"),
        ("Unrecognized Custom DB Tier", "MODERATE", "Moderate"),
    ]
    for raw_ext, exp_code, exp_label in ext_cases:
        code, label, _, _ = normalize_anomaly_risk_level(raw_ext)
        assert code == exp_code, f"{raw_ext} -> expected {exp_code}, got {code}"
        assert label == exp_label, f"{raw_ext} -> expected {exp_label}, got {label}"

    ext_req = AnomalyRequest(
        entity="External Core DB",
        category="Transaction",
        amount=650.0,
        hour_of_day=23,
        risk_level="Critical"
    )
    ext_pred = predict_anomaly(ext_req)
    assert ext_pred['risk_level'] == 'HIGH', f"Expected HIGH for Critical external input, got {ext_pred['risk_level']}"
    assert ext_pred['severity'] == 'High', f"Expected High for Critical external input, got {ext_pred['severity']}"

    print("=" * 60)
    print("ALL 5 VERIFICATION TESTS PASSED SUCCESSFULLY!")
    print("=" * 60)

if __name__ == '__main__':
    run_tests()
