"""
Data Hub API — Canonical Mapping Synonyms
Dictionary of Indonesian and English banking, accounting, and ERP column name synonyms.
Used by the AI/rule-based auto-mapper to map client schemas to AuditSphere canonical entities.
"""

CANONICAL_SCHEMAS = {
    "transactions": {
        "description": "Financial and operational transactions (fact_transactions)",
        "required_fields": ["trx_ref_number", "trx_date", "amount"],
        "fields": {
            "trx_ref_number": ["no_transaksi", "no_trx", "ref_no", "reference_number", "trx_id", "id_transaksi", "nomor_referensi", "no_ref", "kode_transaksi", "trans_id", "uuid"],
            "trx_date": ["tgl_transaksi", "tgl_trx", "tanggal", "trans_date", "posting_date", "transaction_date", "tgl_posting", "created_at", "trans_dt", "waktu_transaksi"],
            "trx_time": ["jam_transaksi", "jam_trx", "trx_time", "trans_time", "waktu"],
            "amount": ["nominal", "jumlah", "nilai", "nilai_transaksi", "total_amount", "trans_amount", "debit_credit_amount", "saldo_trx", "amount", "trx_amount"],
            "channel": ["kanal", "channel_type", "tipe_kanal", "media", "trans_channel", "channel", "metode"],
            "category": ["kategori", "jenis_transaksi", "trx_category", "trans_type", "tipe", "category", "txn_type"],
            "account_id": ["no_rekening", "id_rekening", "nomor_rekening", "account_no", "acc_num", "rekening", "account_id", "norek"],
            "branch_id": ["kode_cabang", "id_cabang", "cabang", "branch_code", "office_id", "branch_id", "unit_kerja"],
            "teller_id": ["user_id", "teller", "id_petugas", "operator", "posted_by", "teller_id"],
            "description": ["keterangan", "deskripsi", "uraian", "narasi", "memo", "remarks", "narration", "description"],
            "is_suspicious": ["flag_mencurigakan", "is_suspicious", "suspicious", "flag_str", "fraud_flag"]
        }
    },
    "accounts": {
        "description": "Customer banking accounts (acct_cleaned)",
        "required_fields": ["account_number", "balance"],
        "fields": {
            "account_number": ["no_rekening", "nomor_rekening", "acc_no", "account_no", "rekening", "norek", "account_number"],
            "balance": ["saldo", "saldo_akhir", "current_balance", "available_balance", "nominal_saldo", "balance"],
            "account_type": ["jenis_rekening", "tipe_rekening", "product_type", "acct_type", "account_type"],
            "interest_rate": ["bunga", "suku_bunga", "interest_rate", "bagi_hasil", "nisbah"],
            "status": ["status_rekening", "acct_status", "kondisi", "state", "status"],
            "customer_id": ["cif", "id_nasabah", "nomor_nasabah", "cust_id", "customer_no", "cif_no"],
            "branch_id": ["kode_cabang", "id_cabang", "cabang", "branch_code", "branch_id"],
            "opened_date": ["tgl_buka", "opened_date", "open_date", "tgl_registrasi", "created_at"]
        }
    },
    "customers": {
        "description": "Customer master records (cust_cleaned)",
        "required_fields": ["cif_number", "customer_name"],
        "fields": {
            "cif_number": ["cif", "no_cif", "nomor_cif", "customer_code", "kode_nasabah", "cif_number", "cif_no"],
            "customer_name": ["nama_nasabah", "nama_lengkap", "nama", "cust_name", "client_name", "customer_name"],
            "customer_type": ["jenis_nasabah", "tipe_nasabah", "tipe_customer", "individual_corporate", "customer_type"],
            "risk_profile": ["profil_risiko", "risk_rating", "tingkat_risiko", "risk_category", "risk_profile"],
            "branch_id": ["kode_cabang", "id_cabang", "cabang", "branch_id", "branch_code"],
            "is_active": ["aktif", "is_active", "status_aktif", "active"]
        }
    },
    "loans": {
        "description": "Financing and loan portfolios (fact_loans)",
        "required_fields": ["loan_number", "outstanding", "collectability"],
        "fields": {
            "loan_number": ["no_pinjaman", "no_kredit", "no_kontrak", "loan_no", "fasilitas_no", "nomor_akad", "loan_number"],
            "loan_type": ["jenis_kredit", "jenis_pembiayaan", "tipe_kredit", "loan_type", "produk_kredit"],
            "plafond": ["plafon", "plafond_kredit", "limit", "credit_limit", "komitmen", "plafond"],
            "outstanding": ["baki_debet", "sisa_pinjaman", "outstanding_principal", "saldo_pokok", "outstanding", "baki_debet_pokok"],
            "interest_rate": ["suku_bunga", "bunga", "margin", "rate", "interest", "interest_rate"],
            "tenor_months": ["jangka_waktu", "tenor", "tenor_bulan", "period_months", "tenor_months"],
            "collateral_value": ["nilai_agunan", "agunan", "collateral", "nilai_jaminan", "collateral_value"],
            "collectability": ["kolektibilitas", "kol", "kualitas_kredit", "kolek", "collect_status", "collectability"],
            "days_past_due": ["hari_tunggakan", "dpd", "past_due_days", "tunggakan_hari", "days_past_due"],
            "status": ["status_kredit", "loan_status", "status"]
        }
    },
    "gl_entries": {
        "description": "General Ledger journal entries (fact_gl_entries)",
        "required_fields": ["gl_date", "gl_account_code", "debit_amount", "credit_amount"],
        "fields": {
            "gl_date": ["tgl_jurnal", "tgl_posting", "entry_date", "posting_date", "gl_date", "tanggal"],
            "gl_account_code": ["kode_akun", "nomor_akun", "coa", "account_code", "gl_code", "no_coa", "gl_account_code"],
            "gl_account_name": ["nama_akun", "nama_coa", "account_name", "coa_name", "gl_account_name"],
            "voucher_number": ["no_bukti", "nomor_voucher", "voucher_no", "journal_ref", "no_jurnal", "voucher_number"],
            "debit_amount": ["debet", "debit", "jumlah_debet", "debit_val", "debit_amount"],
            "credit_amount": ["kredit", "credit", "jumlah_kredit", "credit_val", "credit_amount"],
            "description": ["keterangan", "deskripsi", "uraian", "narasi", "remarks", "description"],
            "posted_by": ["petugas", "user_id", "posted_by", "created_by"]
        }
    }
}
