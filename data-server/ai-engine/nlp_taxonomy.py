import os
import re
import json
from datetime import datetime
from typing import Dict, List, Any, Optional, Tuple

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
DATA_DIR = os.path.join(BASE_DIR, "data")
os.makedirs(DATA_DIR, exist_ok=True)

TAXONOMY_FILE = os.path.join(DATA_DIR, "nlp_taxonomy.json")
IMPORTED_DOCS_FILE = os.path.join(DATA_DIR, "imported_nlp_documents.json")
LIVE_CORPUS_FILE = os.path.join(DATA_DIR, "live_nlp_corpus.json")

DEFAULT_TAXONOMY: Dict[str, Any] = {
    "risk_categories": [
        {
            "name": "Financial",
            "is_default": True,
            "description": "Risiko terkait laporan keuangan, kas, likuiditas, kredit, pembayaran, anggaran, akuntansi, dan pajak.",
            "aliases": ["keuangan", "finansial", "finance", "accounting", "treasury", "credit", "lending", "funding"],
            "keywords": [
                "kas", "pembayaran", "invoice", "jurnal", "rekonsiliasi", "piutang", "utang", "kredit",
                "pencairan", "anggaran", "biaya", "laporan keuangan", "neraca", "laba rugi", "likuiditas",
                "agunan", "pencadangan", "ckpn", "npl", "bunga", "deposito", "buku besar", "ledger",
                "financial", "payment", "disbursement", "cash", "revenue", "expense", "voucher", "reimbursement"
            ]
        },
        {
            "name": "Operational",
            "is_default": True,
            "description": "Risiko terkait proses bisnis internal, pengadaan, logistik, SDM, pelayanan cabang, dan efisiensi operasional.",
            "aliases": ["operations", "operasional", "operation", "procurement", "human resources", "sdm", "strategic", "proses bisnis"],
            "keywords": [
                "operasional", "pengadaan", "vendor", "logistik", "inventaris", "gudang", "persediaan",
                "cabang", "pelayanan", "sla", "keterlambatan", "proses bisnis", "antrian", "pemeliharaan",
                "aset tetap", "sdm", "karyawan", "pelatihan", "produktivitas", "kinerja", "kontrak kerja",
                "operational", "operations", "supply chain", "inventory", "branch", "service", "workflow", "maintenance"
            ]
        },
        {
            "name": "Compliance",
            "is_default": True,
            "description": "Risiko terkait kepatuhan regulasi (OJK, BI, PPATK, UU), SOP internal, tata kelola (GCG), APU-PPT/KYC, dan hukum.",
            "aliases": ["kepatuhan", "governance", "legal", "regulatory", "regulasi", "gcg", "hukum", "aml", "kyc"],
            "keywords": [
                "kepatuhan", "regulasi", "pojk", "pbi", "ppatk", "apu-ppt", "kyc", "cdd", "edd",
                "pencucian uang", "aml", "sop", "kebijakan", "tata kelola", "gcg", "komite", "direksi",
                "sanksi", "denda", "pelaporan", "perizinan", "lisensi", "hukum", "kontrak", "piagam",
                "compliance", "regulatory", "policy", "violation", "watchlist", "pep", "str", "governance"
            ]
        },
        {
            "name": "Technology",
            "is_default": True,
            "description": "Risiko terkait sistem TI, keamanan siber, kontrol akses, server, database, jaringan, arsitektur sistem, spesifikasi resource, dan integritas data.",
            "aliases": ["teknologi", "it", "ti", "cybersecurity", "cyber", "information technology", "it control", "security", "tech", "infrastructure", "infrastruktur", "spesifikasi", "arsitektur"],
            "keywords": [
                "teknologi", "sistem", "server", "database", "aplikasi", "jaringan", "siber", "cyber",
                "password", "hak akses", "user access", "login", "firewall", "enkripsi", "backup",
                "disaster recovery", "drc", "patch", "malware", "ransomware", "api", "core banking",
                "technology", "privilege", "admin", "log", "vulnerability", "data breach", "cloud", "otomasi",
                "keamanan", "keamanan sistem", "keamanan informasi", "keamanan jaringan", "infrastructure",
                "infrastruktur", "hardware", "software", "security", "security audit", "it control",
                "spesifikasi", "specification", "resource", "resources", "arsitektur", "architecture",
                "cpu", "ram", "memory", "storage", "kapasitas", "capacity", "deployment", "hosting",
                "auditsphere", "spesifikasi resource", "resource specification", "spesifikasi teknis",
                "technical specification", "konfigurasi", "configuration", "bandwidth", "node", "cluster"
            ]
        }
    ],
    "control_indicators": [
        {
            "name": "Fraud Indication",
            "is_default": True,
            "severity_boost": 22,
            "default_sentiment": "Negative",
            "description": "Kalimat menunjukkan indikasi kecurangan (fraud), transaksi fiktif, penggelapan, manipulasi data, atau suap/gratifikasi.",
            "keywords": [
                "fraud", "kecurangan", "fiktif", "penggelapan", "manipulasi", "pemalsuan", "mark-up", "markup",
                "suap", "gratifikasi", "kickback", "pencucian uang", "rekayasa", "penyalahgunaan dana",
                "penyalahgunaan wewenang", "transaksi mencurigakan", "indikasi fraud", "embezzlement",
                "forgery", "bribery", "ghost employee", "vendor fiktif", "split transaction", "memecah transaksi",
                "dormant account", "rapid turnover", "escalated_to_ppatk", "aml_str_alert"
            ]
        },
        {
            "name": "Segregation of Duties (SoD) Violation",
            "is_default": True,
            "severity_boost": 16,
            "default_sentiment": "Negative",
            "description": "Pelanggaran pemisahan fungsi (maker-checker/dual control) atau eksekusi & otorisasi oleh pihak yang sama.",
            "keywords": [
                "pemisahan tugas", "pemisahan fungsi", "segregation of duties", "sod_violation", "sod violation",
                "dual control", "maker-checker", "maker dan checker", "pejabat yang sama", "user yang sama",
                "same user initiated and authorized", "tanpa otorisasi berjenjang", "single approval",
                "override_auth", "is_override_auth", "branch_override", "bypass otorisasi", "konflik kepentingan"
            ]
        },
        {
            "name": "IT & Access Control Weakness",
            "is_default": True,
            "severity_boost": 14,
            "default_sentiment": "Negative",
            "description": "Kelemahan pengendalian akses TI, akun bersama, hak akses berlebih, atau kegagalan patch/backup.",
            "keywords": [
                "hak akses", "user dormant", "akun aktif", "shared account", "akun bersama", "password",
                "mutation_table_permissions", "unauthorized access", "akses tanpa izin", "gagal login",
                "belum di-patch", "kerentanan", "log akses", "super_admin", "adm_super", "backup gagal",
                "enkripsi", "firewall", "kontrol akses", "privilege escalation"
            ]
        },
        {
            "name": "Internal Control Weakness",
            "is_default": True,
            "severity_boost": 12,
            "default_sentiment": "Negative",
            "description": "Kelemahan pengendalian internal seperti kurangnya verifikasi, rekonsiliasi terlambat, atau dokumen pendukung tidak lengkap.",
            "keywords": [
                "kelemahan kontrol", "kelemahan pengendalian", "pengendalian internal", "belum dilakukan rekonsiliasi",
                "tidak dilakukan verifikasi", "dokumen pendukung tidak lengkap", "tanpa bukti", "selisih",
                "keterlambatan", "kurangnya pengawasan", "belum terdokumentasi", "tidak memadai", "kelemahan",
                "ketidaksesuaian", "kontrol belum berjalan", "monitoring lemah", "weakness", "deficiency",
                "unreconciled", "missing documentation", "pending_approval", "kol-2", "npl"
            ]
        },
        {
            "name": "Compliance Deviation",
            "is_default": True,
            "severity_boost": 10,
            "default_sentiment": "Negative",
            "description": "Ketidakpatuhan terhadap SOP internal, regulasi eksternal (OJK/BI/Pajak), atau keterlambatan pelaporan wajib.",
            "keywords": [
                "ketidakpatuhan", "pelanggaran sop", "tidak sesuai prosedur", "tidak sesuai regulasi",
                "melewati batas waktu", "keterlambatan pelaporan", "dokumen kyc belum", "cdd belum",
                "belum diperbarui", "non-compliance", "deviation", "melanggar ketentuan", "batas maksimum"
            ]
        },
        {
            "name": "Adequate Control",
            "is_default": True,
            "severity_boost": -20,
            "default_sentiment": "Positive",
            "description": "Pengendalian internal berjalan efektif, patuh terhadap prosedur, dan tidak ditemukan penyimpangan material.",
            "keywords": [
                " telah memadai", "berjalan efektif", "sesuai prosedur", "patuh", "tidak ditemukan penyimpangan",
                "telah sesuai", "terkendali dengan baik", "efektif", "memuaskan", "adequate", "effective",
                "compliant", "terverifikasi dual control", "kol-1", "lancar"
            ]
        }
    ]
}


def load_taxonomy() -> Dict[str, Any]:
    if os.path.exists(TAXONOMY_FILE):
        try:
            with open(TAXONOMY_FILE, "r", encoding="utf-8") as f:
                data = json.load(f)
                if "risk_categories" in data and "control_indicators" in data:
                    # Sync default category keywords and aliases from DEFAULT_TAXONOMY
                    default_cats = {c["name"]: c for c in DEFAULT_TAXONOMY.get("risk_categories", [])}
                    for cat in data.get("risk_categories", []):
                        if cat.get("name") in default_cats:
                            def_kw = default_cats[cat["name"]].get("keywords", [])
                            cat["keywords"] = list(dict.fromkeys(cat.get("keywords", []) + def_kw))
                            def_aliases = default_cats[cat["name"]].get("aliases", [])
                            cat["aliases"] = list(dict.fromkeys(cat.get("aliases", []) + def_aliases))
                    return data
        except Exception as e:
            print(f"Warning loading nlp_taxonomy.json: {e}")
    save_taxonomy(DEFAULT_TAXONOMY)
    return json.loads(json.dumps(DEFAULT_TAXONOMY))


def save_taxonomy(taxonomy: Dict[str, Any]) -> None:
    try:
        with open(TAXONOMY_FILE, "w", encoding="utf-8") as f:
            json.dump(taxonomy, f, indent=2, ensure_ascii=False)
    except Exception as e:
        print(f"Warning saving nlp_taxonomy.json: {e}")


def get_active_risk_categories() -> List[str]:
    tax = load_taxonomy()
    return [c["name"] for c in tax.get("risk_categories", [])]


def get_active_control_indicators() -> List[str]:
    tax = load_taxonomy()
    return [c["name"] for c in tax.get("control_indicators", [])]


def add_taxonomy_item(
    item_type: str,
    name: str,
    keywords: Optional[List[str]] = None,
    description: str = "",
    aliases: Optional[List[str]] = None,
    severity_boost: int = 12,
    default_sentiment: str = "Negative"
) -> Dict[str, Any]:
    """
    Dynamically registers a new Risk Category or Control Indicator into nlp_taxonomy.json.
    """
    tax = load_taxonomy()
    clean_name = name.strip()
    if not clean_name:
        return tax

    kw_list = [k.strip().lower() for k in (keywords or []) if k.strip()]
    if clean_name.lower() not in kw_list:
        kw_list.append(clean_name.lower())

    if item_type in ("risk_category", "category", "riskCategory"):
        existing = tax.get("risk_categories", [])
        for cat in existing:
            if cat["name"].lower() == clean_name.lower():
                merged_kw = list(dict.fromkeys(cat.get("keywords", []) + kw_list))
                cat["keywords"] = merged_kw
                if description:
                    cat["description"] = description
                save_taxonomy(tax)
                return tax

        existing.append({
            "name": clean_name,
            "is_default": False,
            "description": description or f"Kategori risiko dinamis: {clean_name}",
            "aliases": [a.strip().lower() for a in (aliases or []) if a.strip()],
            "keywords": kw_list
        })
        tax["risk_categories"] = existing

        # Seed the live training corpus so the supervised ML model learns this category immediately
        kw_phrase = ", ".join(kw_list[:8])
        seed_items = [
            {
                "text": f"Pemeriksaan audit dan evaluasi kepatuhan pada domain {clean_name}: ditemukan indikasi terkait {kw_phrase}.",
                "risk_category": clean_name,
                "sentiment": "Negative",
                "impact": 4,
                "likelihood": 3,
                "control_indicator": "Internal Control Weakness"
            },
            {
                "text": f"Hasil audit operasional menunjukkan kelemahan implementasi {clean_name} khususnya pada {kw_list[0] if kw_list else clean_name}.",
                "risk_category": clean_name,
                "sentiment": "Negative",
                "impact": 4,
                "likelihood": 4,
                "control_indicator": "Internal Control Weakness"
            }
        ]
        upsert_live_nlp_corpus(seed_items)
    elif item_type in ("control_indicator", "indicator", "controlIndicator"):
        existing = tax.get("control_indicators", [])
        for ind in existing:
            if ind["name"].lower() == clean_name.lower():
                merged_kw = list(dict.fromkeys(ind.get("keywords", []) + kw_list))
                ind["keywords"] = merged_kw
                if description:
                    ind["description"] = description
                ind["severity_boost"] = severity_boost
                save_taxonomy(tax)
                return tax

        # Insert before "Adequate Control" so specific risk indicators are evaluated first
        new_ind = {
            "name": clean_name,
            "is_default": False,
            "severity_boost": severity_boost,
            "default_sentiment": default_sentiment,
            "description": description or f"Indikator kontrol dinamis: {clean_name}",
            "keywords": kw_list
        }
        if existing and existing[-1]["name"] == "Adequate Control":
            existing.insert(len(existing) - 1, new_ind)
        else:
            existing.append(new_ind)
        tax["control_indicators"] = existing

        # Seed the live training corpus so the supervised ML model learns this indicator immediately
        kw_phrase = ", ".join(kw_list[:8])
        seed_items = [
            {
                "text": f"Hasil pengujian substantif menemukan indikasi {clean_name}: terdeteksi kelemahan terkait {kw_phrase}.",
                "risk_category": "Operational",
                "sentiment": default_sentiment,
                "impact": 4,
                "likelihood": 4,
                "control_indicator": clean_name
            },
            {
                "text": f"Evaluasi kertas kerja audit mengonfirmasi adanya temuan {clean_name} pada proses operasional cabang.",
                "risk_category": "Operational",
                "sentiment": default_sentiment,
                "impact": 4,
                "likelihood": 4,
                "control_indicator": clean_name
            }
        ]
        upsert_live_nlp_corpus(seed_items)

    save_taxonomy(tax)
    return tax


def normalize_or_resolve_category(raw_category: Optional[str]) -> Optional[str]:
    """
    Maps a raw category name to an active category in the taxonomy (either the 4 default
    categories or any dynamically added category).
    """
    if not raw_category:
        return None
    clean = raw_category.strip()
    clean_lower = clean.lower()
    tax = load_taxonomy()
    categories = tax.get("risk_categories", [])

    # 1. Exact match against active categories (including newly added custom categories)
    for cat in categories:
        if cat["name"].lower() == clean_lower:
            return cat["name"]

    # 2. Alias match
    for cat in categories:
        aliases = [a.lower() for a in cat.get("aliases", [])]
        if clean_lower in aliases:
            return cat["name"]

    # 3. Substring heuristic
    if "finan" in clean_lower or "keuangan" in clean_lower or "credit" in clean_lower or "treasury" in clean_lower:
        return "Financial"
    if "operat" in clean_lower or "sdm" in clean_lower or "human" in clean_lower or "procure" in clean_lower or "strateg" in clean_lower:
        return "Operational"
    if "compli" in clean_lower or "kepatuhan" in clean_lower or "govern" in clean_lower or "legal" in clean_lower or "regul" in clean_lower:
        return "Compliance"
    if any(w in clean_lower for w in [
        "tech", "it", "ti", "cyber", "siber", "sistem", "system", "spesifikasi", "resource",
        "server", "hardware", "software", "arsitektur", "architecture", "kapasitas", "capacity",
        "cloud", "auditsphere", "storage", "database", "jaringan", "network"
    ]):
        return "Technology"

    return clean


def check_keyword_in_text(kw: str, text: str) -> bool:
    """Checks whether a keyword matches text as a whole word or delimited phrase to prevent substring false positives."""
    kw_clean = kw.strip().lower()
    text_clean = text.lower()
    if not kw_clean:
        return False
    if " " in kw_clean or "-" in kw_clean or "_" in kw_clean:
        pattern = r'(?:^|[\s,.\-;:!?/()\[\]{}"])+' + re.escape(kw_clean) + r'(?:$|[\s,.\-;:!?/()\[\]{}"])+'
        return bool(re.search(pattern, " " + text_clean + " "))
    else:
        pattern = r"\b" + re.escape(kw_clean) + r"\b"
        return bool(re.search(pattern, text_clean))


def classify_risk_category(
    text: str,
    model_pred: Optional[str] = None,
    explicit_category: Optional[str] = None
) -> Tuple[str, float]:
    """
    Hybrid IndoBERT + Dynamic Taxonomy Risk Category classifier.
    Supports the 4 core categories (Financial, Operational, Compliance, Technology)
    plus any dynamically added Risk Category in nlp_taxonomy.json.
    """
    tax = load_taxonomy()
    categories = tax.get("risk_categories", [])
    text_lower = (text or "").lower()

    if explicit_category:
        resolved = normalize_or_resolve_category(explicit_category)
        if resolved and resolved in [c["name"] for c in categories]:
            return resolved, 0.96

    scores: Dict[str, float] = {cat["name"]: 0.0 for cat in categories}
    for cat in categories:
        cat_name = cat["name"]
        for kw in cat.get("keywords", []):
            if check_keyword_in_text(kw, text_lower):
                # Longer phrases carry higher specificity weight
                weight = 2.5 if (" " in kw or "-" in kw) else 1.8
                scores[cat_name] = scores.get(cat_name, 0.0) + weight

    resolved_model_pred = normalize_or_resolve_category(model_pred)
    if resolved_model_pred and resolved_model_pred in scores:
        max_kw_score = max(scores.values()) if scores else 0.0
        # If there is strong keyword evidence for another category, model vote gets modest weight
        # so generic model bias doesn't overrule clear technical or compliance terms
        boost = 1.0 if max_kw_score >= 2.0 else 1.5
        scores[resolved_model_pred] += boost

    best_cat = "Operational"
    best_score = -1.0
    for cat_name, sc in scores.items():
        if sc > best_score:
            best_score = sc
            best_cat = cat_name

    if best_score <= 0.0:
        best_cat = resolved_model_pred if (resolved_model_pred and resolved_model_pred in scores) else "Operational"
        confidence = 0.86
    else:
        confidence = min(0.98, round(0.87 + min(best_score, 10.0) * 0.011, 2))

    return best_cat, confidence


def detect_control_indicator_and_severity(
    text: str,
    model_sentiment: Optional[str] = None,
    model_impact: int = 3,
    model_likelihood: int = 3,
    risk_category: str = "Operational"
) -> Dict[str, Any]:
    """
    Analyzes extracted document/note text to determine:
    - control_indicator (Fraud Indication, Internal Control Weakness, SoD Violation, etc. or custom indicator)
    - sentiment (Negative, Neutral, Positive)
    - impact (1-5), likelihood (1-5), severityScore (0-100), risk_level
    """
    tax = load_taxonomy()
    indicators = tax.get("control_indicators", [])
    text_lower = (text or "").lower()

    matched_indicator = None
    best_ind_score = 0.0
    matched_boost = 0
    default_sent = "Negative"

    for ind in indicators:
        ind_name = ind["name"]
        ind_score = 0.0
        for kw in ind.get("keywords", []):
            if check_keyword_in_text(kw, text_lower):
                ind_score += 2.5 if (" " in kw or "-" in kw) else 1.8
        if ind_score > best_ind_score:
            best_ind_score = ind_score
            matched_indicator = ind_name
            matched_boost = int(ind.get("severity_boost", 10))
            default_sent = ind.get("default_sentiment", "Negative")

    # Determine Sentiment
    neg_markers = [
        "kelemahan", "indikasi", "fraud", "kecurangan", "fiktif", "selisih", "belum",
        "tidak", "terlambat", "keterlambatan", "pelanggaran", "penyimpangan", "tanpa",
        "kurangnya", "gagal", "kerugian", "sanksi", "denda", "anomali", "bypass",
        "override", "escalated", "violation", "weakness", "unauthorized", "overdue"
    ]
    pos_markers = [
        "memadai", "efektif", "sesuai prosedur", "patuh", "terkendali", "lancar",
        "tidak ditemukan penyimpangan", "selesai tepat waktu", "optimal", "terverifikasi"
    ]

    neg_hits = sum(1 for w in neg_markers if w in text_lower)
    pos_hits = sum(1 for w in pos_markers if w in text_lower)

    if matched_indicator == "Adequate Control" or (pos_hits >= 2 and neg_hits == 0):
        sentiment = "Positive"
        if not matched_indicator:
            matched_indicator = "Adequate Control"
    elif matched_indicator and default_sent == "Negative":
        sentiment = "Negative"
    elif neg_hits > 0:
        sentiment = "Negative"
        if not matched_indicator:
            matched_indicator = "Internal Control Weakness"
            matched_boost = 12
    elif model_sentiment in ("Negative", "Neutral", "Positive"):
        sentiment = model_sentiment
        if sentiment == "Negative" and not matched_indicator:
            matched_indicator = "Internal Control Weakness"
            matched_boost = 10
        elif sentiment == "Positive" and not matched_indicator:
            matched_indicator = "Adequate Control"
            matched_boost = -15
        elif not matched_indicator:
            matched_indicator = "Internal Control Weakness"
            matched_boost = 5
    else:
        sentiment = "Neutral"
        if not matched_indicator:
            if risk_category == "Technology":
                matched_indicator = "IT & Access Control Weakness"
                matched_boost = 10
            elif risk_category == "Compliance":
                matched_indicator = "Compliance Deviation"
                matched_boost = 8
            else:
                matched_indicator = "Internal Control Weakness"

    # Calculate Impact & Likelihood (1-5)
    imp = max(1, min(5, int(model_impact or 3)))
    lik = max(1, min(5, int(model_likelihood or 3)))

    if matched_indicator == "Fraud Indication":
        imp = max(imp, 5 if any(w in text_lower for w in ["miliar", "juta", "fiktif", "penggelapan", "ppatk", "suap"]) else 4)
        lik = max(lik, 4)
    elif matched_indicator in ("Segregation of Duties (SoD) Violation", "IT & Access Control Weakness"):
        imp = max(imp, 4)
        lik = max(lik, 4)
    elif matched_indicator == "Internal Control Weakness":
        imp = max(imp, 3 if neg_hits == 1 else 4)
        lik = max(lik, 3)
    elif matched_indicator == "Compliance Deviation":
        imp = max(imp, 4 if any(w in text_lower for w in ["sanksi", "denda", "pojk", "pbi", "ppatk"]) else 3)
        lik = max(lik, 3)
    elif matched_indicator == "Adequate Control" or sentiment == "Positive":
        imp = min(imp, 2)
        lik = min(lik, 2)

    raw_score = imp * lik
    base_severity = int(raw_score * 3.4) + matched_boost + min(neg_hits * 3, 12)
    if sentiment == "Positive":
        severity_score = max(12, min(35, base_severity))
    elif sentiment == "Neutral":
        severity_score = max(36, min(62, base_severity))
    else:
        severity_score = max(65, min(98, base_severity))

    if raw_score >= 20 or severity_score >= 85:
        risk_level = "HIGH"
    elif raw_score >= 13 or severity_score >= 72:
        risk_level = "MODERATE_HIGH"
    elif raw_score >= 8 or severity_score >= 50:
        risk_level = "MODERATE"
    elif raw_score >= 4 or severity_score >= 30:
        risk_level = "LOW_MODERATE"
    else:
        risk_level = "LOW"

    return {
        "sentiment": sentiment,
        "control_indicator": matched_indicator,
        "impact": imp,
        "likelihood": lik,
        "severityScore": severity_score,
        "risk_level": risk_level
    }


def extract_finding_and_recommendation(
    text: str,
    risk_category: str = "Operational",
    control_indicator: str = "Internal Control Weakness",
    sentiment: str = "Negative"
) -> Dict[str, str]:
    """
    Extracts structured Audit Finding (Temuan) and Audit Recommendation (Rekomendasi)
    from raw document text or audit notes using sentence-level clause analysis.
    """
    clean_text = re.sub(r"\s+", " ", (text or "")).strip()
    if not clean_text:
        return {
            "extracted_finding": "Tidak ditemukan narasi temuan pada dokumen.",
            "extracted_recommendation": "Lakukan verifikasi kelengkapan isi dokumen sumber."
        }

    # Split into sentences while preserving abbreviations
    raw_sentences = re.split(r"(?<=[.!?])\s+|\n+", clean_text)
    sentences = [s.strip(" -•\t\r\n") for s in raw_sentences if len(s.strip(" -•\t\r\n")) >= 15]
    if not sentences:
        sentences = [clean_text]

    finding_markers = [
        "temuan:", "kondisi:", "ditemukan", "terdapat", "hasil pemeriksaan", "hasil audit",
        "mengindikasikan", "indikasi", "kelemahan", "belum", "tidak", "selisih", "penyimpangan",
        "pelanggaran", "terlambat", "keterlambatan", "fiktif", "tanpa otorisasi", "override",
        "finding:", "observed", "identified", "violation", "weakness", "discrepancy", "alert",
        "spesifikasi", "resource", "kapasitas", "arsitektur", "server", "kebutuhan", "parameter",
        "utilisasi", "storage", "ram", "cpu", "memory", "konfigurasi", "persyaratan", "observasi",
        "analisis", "evaluasi", "pengujian", "performansi", "kinerja", "beban", "threshold"
    ]
    rec_markers = [
        "rekomendasi:", "direkomendasikan", "disarankan", "saran:", "tindak lanjut:",
        "agar ", "perlu melakukan", "wajib ", "segera melakukan", "harus ", "meningkatkan",
        "memperketat", "menyusun", "melakukan rekonsiliasi", "recommendation:", "recommended",
        "should ", "must ", "implement", "establish", "enforce"
    ]

    finding_candidates: List[Tuple[int, str]] = []
    rec_candidates: List[Tuple[int, str]] = []

    for idx, sent in enumerate(sentences):
        s_lower = sent.lower()
        # Check if sentence is explicitly a recommendation
        rec_score = sum(2 if m in s_lower else 0 for m in rec_markers)
        find_score = sum(2 if m in s_lower else 0 for m in finding_markers)

        if rec_score > 0 and rec_score >= find_score:
            cleaned_rec = re.sub(r"^(rekomendasi|saran|tindak lanjut|recommendation)\s*[:\-]\s*", "", sent, flags=re.IGNORECASE).strip()
            rec_candidates.append((rec_score, cleaned_rec))
        elif find_score > 0:
            cleaned_find = re.sub(r"^(temuan|kondisi|finding)\s*[:\-]\s*", "", sent, flags=re.IGNORECASE).strip()
            finding_candidates.append((find_score, cleaned_find))

    if finding_candidates:
        finding_candidates.sort(key=lambda x: x[0], reverse=True)
        extracted_finding = finding_candidates[0][1]
    else:
        extracted_finding = sentences[0]

    # Prepend informative observation context for specification / technical documents without explicit violation wording
    if risk_category == "Technology" and any(w in clean_text.lower() for w in ["spesifikasi", "resource", "kapasitas", "arsitektur", "server", "cpu", "ram", "memory", "storage"]):
        if not any(extracted_finding.lower().startswith(m) for m in ["temuan", "kondisi", "kelemahan", "observasi", "catatan", "hasil"]):
            extracted_finding = f"Observasi Teknis & Spesifikasi: {extracted_finding}"

    if len(extracted_finding) > 260:
        extracted_finding = extracted_finding[:257].rstrip() + "..."

    if rec_candidates:
        rec_candidates.sort(key=lambda x: x[0], reverse=True)
        extracted_recommendation = rec_candidates[0][1]
    else:
        # Synthesize contextual recommendation based on control_indicator & risk_category
        if sentiment == "Positive" or control_indicator == "Adequate Control":
            extracted_recommendation = (
                f"Pertahankan efektivitas pengendalian internal pada domain {risk_category} "
                f"serta lakukan pemantauan berkala melalui Continuous Auditing."
            )
        elif control_indicator == "Fraud Indication":
            extracted_recommendation = (
                "Lakukan audit investigatif khusus (forensik), bekukan sementara akses/transaksi terkait, "
                "amankan bukti jejak audit (audit trail), dan eskalasikan ke Komite Audit serta Unit Anti-Fraud."
            )
        elif control_indicator == "Segregation of Duties (SoD) Violation":
            extracted_recommendation = (
                "Terapkan pemisahan fungsi (Maker-Checker-Signer) secara ketat pada sistem, cabut hak override "
                "tunggal, dan wajibkan otorisasi berjenjang untuk setiap transaksi atau perubahan parameter."
            )
        elif control_indicator == "IT & Access Control Weakness":
            if any(w in clean_text.lower() for w in ["spesifikasi", "resource", "kapasitas", "capacity", "ram", "cpu", "memory", "storage", "server", "arsitektur"]):
                extracted_recommendation = (
                    "Lakukan capacity planning dan uji beban (stress test) berkala, pastikan alokasi resource CPU, RAM, & storage "
                    "sesuai kebutuhan sistem, serta optimalkan arsitektur server dan monitoring utilisasi real-time."
                )
            else:
                extracted_recommendation = (
                    "Lakukan User Access Review (UAR) menyeluruh, nonaktifkan akun dormant/shared account, "
                    "terapkan patch keamanan terbaru, serta aktifkan pemantauan log akses real-time."
                )
        elif control_indicator == "Compliance Deviation":
            extracted_recommendation = (
                "Lakukan pemenuhan kepatuhan terhadap regulasi/SOP terkait segera, lengkapi dokumen KYC/CDD "
                "yang tertunda, dan perkuat mekanisme early-warning jadwal pelaporan regulasi."
            )
        elif risk_category == "Financial":
            extracted_recommendation = (
                "Wajibkan rekonsiliasi berkala secara independen, perketat verifikasi dokumen pendukung "
                "sebelum pencairan dana, dan terapkan ambang batas otorisasi berjenjang."
            )
        elif risk_category == "Technology":
            if any(w in clean_text.lower() for w in ["spesifikasi", "resource", "kapasitas", "capacity", "ram", "cpu", "memory", "storage", "server", "arsitektur"]):
                extracted_recommendation = (
                    "Lakukan capacity planning dan uji beban (stress test) berkala, pastikan alokasi resource CPU, RAM, & storage "
                    "sesuai proyeksi beban kerja, serta optimalkan arsitektur server dan monitoring utilisasi real-time."
                )
            else:
                extracted_recommendation = (
                    "Perkuat kontrol keamanan sistem TI, tinjau matriks hak akses pengguna secara periodik, "
                    "dan pastikan pengujian pencadangan (backup/DRC) terdokumentasi."
                )
        elif risk_category == "Compliance":
            extracted_recommendation = (
                "Perbarui SOP internal agar selaras dengan regulasi terbaru dan tingkatkan kepatuhan "
                "verifikasi kepatuhan pada setiap unit operasional."
            )
        else:
            extracted_recommendation = (
                "Perbaiki desain dan pelaksanaan kontrol internal pada unit terkait, lengkapi dokumentasi "
                "kertas kerja, serta tetapkan tenggat waktu rencana tindak lanjut (Action Plan)."
            )

    if len(extracted_recommendation) > 260:
        extracted_recommendation = extracted_recommendation[:257].rstrip() + "..."

    return {
        "extracted_finding": extracted_finding,
        "extracted_recommendation": extracted_recommendation
    }


DEFAULT_MULTI_SOURCE_DOCS: List[Dict[str, Any]] = [
    {
        "docId": "WP-2026-001",
        "title": "Kertas Kerja Pemeriksaan Kas & Otorisasi Pembayaran Cabang Surabaya",
        "source": "Working Paper",
        "sourceType": "Imported Document",
        "fileName": "KKP_Kas_Cabang_Surabaya_2026.pdf",
        "fileType": "PDF",
        "extractionMethod": "PDF Text & OCR Engine",
        "risk_category": "Financial",
        "autoCategory": "Financial",
        "sentiment": "Negative",
        "impact": 5,
        "likelihood": 4,
        "severityScore": 91,
        "control_indicator": "Fraud Indication",
        "controlIndicator": "Fraud Indication",
        "extracted_finding": "Ditemukan 14 transaksi pembayaran vendor senilai Rp 1,85 Miliar yang dipecah (split transaction) untuk menghindari batas otorisasi Pimpinan Cabang dengan bukti penerimaan barang yang terindikasi fiktif.",
        "extractedFinding": "Ditemukan 14 transaksi pembayaran vendor senilai Rp 1,85 Miliar yang dipecah (split transaction) untuk menghindari batas otorisasi Pimpinan Cabang dengan bukti penerimaan barang yang terindikasi fiktif.",
        "extracted_recommendation": "Lakukan audit investigatif khusus atas 14 transaksi vendor tersebut, blokir sementara pembayaran ke rekening vendor terkait, dan terapkan validasi otomatis pendeteksi split invoice pada sistem ERP.",
        "extractedRecommendation": "Lakukan audit investigatif khusus atas 14 transaksi vendor tersebut, blokir sementara pembayaran ke rekening vendor terkait, dan terapkan validasi otomatis pendeteksi split invoice pada sistem ERP.",
        "confidence": 0.96,
        "date": "2026-06-15",
        "excerpt": "Hasil ekstraksi KKP_Kas_Cabang_Surabaya_2026.pdf: Ditemukan 14 transaksi pembayaran vendor senilai Rp 1,85 Miliar yang dipecah (split transaction) untuk menghindari batas otorisasi Pimpinan Cabang dengan bukti penerimaan barang yang terindikasi fiktif. Direkomendasikan untuk melakukan audit investigatif khusus atas 14 transaksi vendor tersebut dan menerapkan validasi otomatis pendeteksi split invoice.",
        "risk_level": "HIGH"
    },
    {
        "docId": "LHA-2026-002",
        "title": "Laporan Hasil Audit Keamanan Akses Core Banking & Server Database",
        "source": "Audit Result Report",
        "sourceType": "Imported Document",
        "fileName": "LHA_IT_Security_CoreBanking_Q2.docx",
        "fileType": "DOCX",
        "extractionMethod": "OOXML Word Extractor",
        "risk_category": "Technology",
        "autoCategory": "Technology",
        "sentiment": "Negative",
        "impact": 4,
        "likelihood": 4,
        "severityScore": 84,
        "control_indicator": "IT & Access Control Weakness",
        "controlIndicator": "IT & Access Control Weakness",
        "extracted_finding": "Terdapat 23 akun administrator database dan pegawai yang telah mutasi/resign masih berstatus aktif serta ditemukan penggunaan shared account pada server produksi tanpa pencatatan log individual.",
        "extractedFinding": "Terdapat 23 akun administrator database dan pegawai yang telah mutasi/resign masih berstatus aktif serta ditemukan penggunaan shared account pada server produksi tanpa pencatatan log individual.",
        "extracted_recommendation": "Cabut segera hak akses 23 akun pegawai mutasi/resign, hapus penggunaan shared account pada server produksi, dan integrasikan proses offboarding HR dengan pencabutan akses otomatis (IAM).",
        "extractedRecommendation": "Cabut segera hak akses 23 akun pegawai mutasi/resign, hapus penggunaan shared account pada server produksi, dan integrasikan proses offboarding HR dengan pencabutan akses otomatis (IAM).",
        "confidence": 0.95,
        "date": "2026-06-14",
        "excerpt": "Hasil ekstraksi LHA_IT_Security_CoreBanking_Q2.docx: Terdapat 23 akun administrator database dan pegawai yang telah mutasi/resign masih berstatus aktif serta ditemukan penggunaan shared account pada server produksi. Direkomendasikan untuk mencabut segera hak akses 23 akun tersebut dan mengintegrasikan proses offboarding HR dengan IAM.",
        "risk_level": "HIGH"
    },
    {
        "docId": "EXT-DB-003",
        "title": "Compliance Incident DB: Alert Transaksi Rekening Dormant & Eskalasi PPATK",
        "source": "External DB: Compliance Incident Records",
        "sourceType": "External Database",
        "fileName": "compliance_incident_records.investigator_notes",
        "fileType": "DB_RECORD",
        "extractionMethod": "External Data Source Sync",
        "risk_category": "Compliance",
        "autoCategory": "Compliance",
        "sentiment": "Negative",
        "impact": 5,
        "likelihood": 4,
        "severityScore": 92,
        "control_indicator": "Fraud Indication",
        "controlIndicator": "Fraud Indication",
        "extracted_finding": "Terjadi perputaran dana mencurigakan (rapid turnover of funds) sebesar Rp 4,2 Miliar pada rekening korporasi dormant tanpa pembaruan dokumen Enhanced Due Diligence (EDD) dan berstatus ESCALATED_TO_PPATK.",
        "extractedFinding": "Terjadi perputaran dana mencurigakan (rapid turnover of funds) sebesar Rp 4,2 Miliar pada rekening korporasi dormant tanpa pembaruan dokumen Enhanced Due Diligence (EDD) dan berstatus ESCALATED_TO_PPATK.",
        "extracted_recommendation": "Selesaikan pelaporan Laporan Transaksi Keuangan Mencurigakan (LTKM/STR) ke PPATK tepat waktu, lakukan pemblokiran sementara mutasi debit rekening dormant, dan wajibkan verifikasi ulang Beneficial Owner.",
        "extractedRecommendation": "Selesaikan pelaporan Laporan Transaksi Keuangan Mencurigakan (LTKM/STR) ke PPATK tepat waktu, lakukan pemblokiran sementara mutasi debit rekening dormant, dan wajibkan verifikasi ulang Beneficial Owner.",
        "confidence": 0.95,
        "date": "2026-06-13",
        "excerpt": "Catatan Database Eksternal (compliance_incident_records.investigator_notes): Terjadi perputaran dana mencurigakan (rapid turnover of funds) sebesar Rp 4,2 Miliar pada rekening korporasi dormant tanpa pembaruan dokumen EDD dan berstatus ESCALATED_TO_PPATK.",
        "risk_level": "HIGH"
    },
    {
        "docId": "EXT-DB-004",
        "title": "External Audit Log DB: Override Otorisasi Jurnal Buku Besar Tanpa Dual Control",
        "source": "External DB: User Access Audit Logs",
        "sourceType": "External Database",
        "fileName": "user_access_audit_logs.action_notes",
        "fileType": "DB_RECORD",
        "extractionMethod": "External Data Source Sync",
        "risk_category": "Financial",
        "autoCategory": "Financial",
        "sentiment": "Negative",
        "impact": 4,
        "likelihood": 4,
        "severityScore": 85,
        "control_indicator": "Segregation of Duties (SoD) Violation",
        "controlIndicator": "Segregation of Duties (SoD) Violation",
        "extracted_finding": "Pejabat yang sama menginisiasi sekaligus mengotorisasi penyesuaian jurnal buku besar (Same user initiated and authorized ledger adjustment) senilai Rp 890 Juta melalui jalur BRANCH_OVERRIDE.",
        "extractedFinding": "Pejabat yang sama menginisiasi sekaligus mengotorisasi penyesuaian jurnal buku besar (Same user initiated and authorized ledger adjustment) senilai Rp 890 Juta melalui jalur BRANCH_OVERRIDE.",
        "extracted_recommendation": "Terapkan pemisahan fungsi (Maker-Checker) secara sistemik pada modul jurnal buku besar, nonaktifkan fitur self-approval pada parameter BRANCH_OVERRIDE, dan lakukan verifikasi bukti pendukung atas jurnal Rp 890 Juta.",
        "extractedRecommendation": "Terapkan pemisahan fungsi (Maker-Checker) secara sistemik pada modul jurnal buku besar, nonaktifkan fitur self-approval pada parameter BRANCH_OVERRIDE, dan lakukan verifikasi bukti pendukung atas jurnal Rp 890 Juta.",
        "confidence": 0.94,
        "date": "2026-06-12",
        "excerpt": "Catatan Database Eksternal (user_access_audit_logs & gl_transactions): Pejabat yang sama menginisiasi sekaligus mengotorisasi penyesuaian jurnal buku besar (SOD_VIOLATION) senilai Rp 890 Juta melalui jalur BRANCH_OVERRIDE.",
        "risk_level": "HIGH"
    },
    {
        "docId": "NOTE-FW-005",
        "title": "Catatan Observasi Fieldwork: Pemeriksaan Fisik Persediaan & Kontrak Pengadaan Logistik",
        "source": "Internal Note: Fieldwork Observation",
        "sourceType": "Internal Note",
        "fileName": "fieldwork_observations.notes",
        "fileType": "NOTE",
        "extractionMethod": "Internal Feature Note Sync",
        "risk_category": "Operational",
        "autoCategory": "Operational",
        "sentiment": "Negative",
        "impact": 3,
        "likelihood": 4,
        "severityScore": 74,
        "control_indicator": "Internal Control Weakness",
        "controlIndicator": "Internal Control Weakness",
        "extracted_finding": "Terdapat keterlambatan pengiriman barang pengadaan logistik rata-rata 18 hari kerja di 4 kantor wilayah serta kartu stok gudang belum direkonsiliasi dengan pencatatan modul persediaan selama 2 kuartal.",
        "extractedFinding": "Terdapat keterlambatan pengiriman barang pengadaan logistik rata-rata 18 hari kerja di 4 kantor wilayah serta kartu stok gudang belum direkonsiliasi dengan pencatatan modul persediaan selama 2 kuartal.",
        "extracted_recommendation": "Lakukan stock opname dan rekonsiliasi fisik persediaan setiap akhir bulan, serta kenakan klaim penalti keterlambatan (SLA) sesuai klausul kontrak kepada vendor logistik.",
        "extractedRecommendation": "Lakukan stock opname dan rekonsiliasi fisik persediaan setiap akhir bulan, serta kenakan klaim penalti keterlambatan (SLA) sesuai klausul kontrak kepada vendor logistik.",
        "confidence": 0.92,
        "date": "2026-06-10",
        "excerpt": "Catatan Observasi Fieldwork: Terdapat keterlambatan pengiriman barang pengadaan logistik rata-rata 18 hari kerja di 4 kantor wilayah serta kartu stok gudang belum direkonsiliasi dengan modul persediaan. Direkomendasikan melakukan stock opname bulanan dan mengenakan penalti SLA.",
        "risk_level": "MODERATE_HIGH"
    },
    {
        "docId": "PKPT-2026-006",
        "title": "Evaluasi Matriks Pengendalian Risiko (RCM) & Kepatuhan Pelaporan OJK",
        "source": "Annual Audit Plan",
        "sourceType": "Imported Document",
        "fileName": "Evaluasi_RCM_Kepatuhan_2026.xlsx",
        "fileType": "XLSX",
        "extractionMethod": "Spreadsheet Table Extractor",
        "risk_category": "Compliance",
        "autoCategory": "Compliance",
        "sentiment": "Neutral",
        "impact": 3,
        "likelihood": 2,
        "severityScore": 52,
        "control_indicator": "Compliance Deviation",
        "controlIndicator": "Compliance Deviation",
        "extracted_finding": "Sebanyak 3 dari 28 kewajiban pelaporan berkala regulator memerlukan pembaruan format template sesuai SEOJK terbaru yang mulai berlaku pada Semester II.",
        "extractedFinding": "Sebanyak 3 dari 28 kewajiban pelaporan berkala regulator memerlukan pembaruan format template sesuai SEOJK terbaru yang mulai berlaku pada Semester II.",
        "extracted_recommendation": "Perbarui template otomasi pelaporan pada modul Compliance sebelum tenggat waktu Semester II dan lakukan uji coba validasi pelaporan bersama Divisi Kepatuhan.",
        "extractedRecommendation": "Perbarui template otomasi pelaporan pada modul Compliance sebelum tenggat waktu Semester II dan lakukan uji coba validasi pelaporan bersama Divisi Kepatuhan.",
        "confidence": 0.91,
        "date": "2026-06-08",
        "excerpt": "Hasil ekstraksi Evaluasi_RCM_Kepatuhan_2026.xlsx: Sebanyak 3 dari 28 kewajiban pelaporan berkala regulator memerlukan pembaruan format template sesuai SEOJK terbaru. Disarankan memperbarui template otomasi pelaporan sebelum Semester II.",
        "risk_level": "MODERATE"
    },
    {
        "docId": "QA-2026-007",
        "title": "Laporan Quality Assurance (QAR) Pelaksanaan Standar Audit Internal",
        "source": "Quality Assurance",
        "sourceType": "Imported Document",
        "fileName": "Laporan_QAR_Internal_Audit_2026.pdf",
        "fileType": "PDF",
        "extractionMethod": "PDF Text & OCR Engine",
        "risk_category": "Operational",
        "autoCategory": "Operational",
        "sentiment": "Positive",
        "impact": 1,
        "likelihood": 2,
        "severityScore": 22,
        "control_indicator": "Adequate Control",
        "controlIndicator": "Adequate Control",
        "extracted_finding": "Pelaksanaan supervisi berjenjang dan dokumentasi kertas kerja audit pada 94% penugasan telah memadai dan berjalan efektif sesuai Standar Internasional Praktik Profesional Audit Internal.",
        "extractedFinding": "Pelaksanaan supervisi berjenjang dan dokumentasi kertas kerja audit pada 94% penugasan telah memadai dan berjalan efektif sesuai Standar Internasional Praktik Profesional Audit Internal.",
        "extracted_recommendation": "Pertahankan standar reviu berjenjang kertas kerja audit dan lanjutkan program sertifikasi berkelanjutan bagi seluruh auditor.",
        "extractedRecommendation": "Pertahankan standar reviu berjenjang kertas kerja audit dan lanjutkan program sertifikasi berkelanjutan bagi seluruh auditor.",
        "confidence": 0.94,
        "date": "2026-06-05",
        "excerpt": "Hasil ekstraksi Laporan_QAR_Internal_Audit_2026.pdf: Pelaksanaan supervisi berjenjang dan dokumentasi kertas kerja audit pada 94% penugasan telah memadai dan berjalan efektif sesuai standar.",
        "risk_level": "LOW"
    }
]


def load_analyzed_documents() -> List[Dict[str, Any]]:
    if os.path.exists(IMPORTED_DOCS_FILE):
        try:
            with open(IMPORTED_DOCS_FILE, "r", encoding="utf-8") as f:
                docs = json.load(f)
                if isinstance(docs, list) and len(docs) > 0:
                    return docs
        except Exception as e:
            print(f"Warning loading imported_nlp_documents.json: {e}")
    save_analyzed_documents(DEFAULT_MULTI_SOURCE_DOCS)
    return json.loads(json.dumps(DEFAULT_MULTI_SOURCE_DOCS))


def save_analyzed_documents(docs: List[Dict[str, Any]]) -> None:
    try:
        with open(IMPORTED_DOCS_FILE, "w", encoding="utf-8") as f:
            json.dump(docs[:150], f, indent=2, ensure_ascii=False)
    except Exception as e:
        print(f"Warning saving imported_nlp_documents.json: {e}")


def upsert_analyzed_document(doc: Dict[str, Any]) -> List[Dict[str, Any]]:
    """
    Prepends a newly analyzed document/note/external-record to imported_nlp_documents.json,
    deduplicating by docId or (title + source).
    """
    docs = load_analyzed_documents()
    doc_id = str(doc.get("docId", "")).strip()
    file_name = str(doc.get("fileName", "")).strip().lower()
    source = str(doc.get("source", "")).strip().lower()

    filtered = []
    for existing in docs:
        ex_id = str(existing.get("docId", "")).strip()
        ex_file = str(existing.get("fileName", "")).strip().lower()
        ex_source = str(existing.get("source", "")).strip().lower()
        if doc_id and ex_id == doc_id:
            continue
        if file_name and ex_file and file_name == ex_file and ex_source == source:
            continue
        filtered.append(existing)

    filtered.insert(0, doc)
    save_analyzed_documents(filtered)
    return filtered


IN_MEMORY_SEED_CORPUS: List[Dict[str, Any]] = [
    {
        "text": "Ditemukan selisih fisik kas pada brankas kasir sebesar Rp 45 Juta tanpa bukti transaksi dan rekonsiliasi harian belum pernah ditandatangani.",
        "risk_category": "Financial",
        "sentiment": "Negative",
        "impact": 4,
        "likelihood": 4,
        "control_indicator": "Internal Control Weakness"
    },
    {
        "text": "Terdapat 12 transaksi transfer valas mencurigakan yang dipecah (split transactions) untuk menghindari ambang batas pelaporan transaksi keuangan mencurigakan (STR).",
        "risk_category": "Compliance",
        "sentiment": "Negative",
        "impact": 5,
        "likelihood": 4,
        "control_indicator": "Fraud Indication"
    },
    {
        "text": "Pegawai bagian accounting merangkap sebagai otorisator pembayaran vendor pada modul e-banking (SoD violation) tanpa kontrol maker-checker independen.",
        "risk_category": "Financial",
        "sentiment": "Negative",
        "impact": 4,
        "likelihood": 4,
        "control_indicator": "Segregation of Duties (SoD) Violation"
    },
    {
        "text": "Server database produksi core banking masih menggunakan shared password administrator dan patch keamanan triwulanan tertunda 6 bulan.",
        "risk_category": "Technology",
        "sentiment": "Negative",
        "impact": 4,
        "likelihood": 4,
        "control_indicator": "IT & Access Control Weakness"
    },
    {
        "text": "Sebanyak 15 berkas kredit komersial belum dilengkapi dokumen jaminan fidusia dan sertifikat hak tanggungan yang telah dilegalisasi notaris.",
        "risk_category": "Compliance",
        "sentiment": "Negative",
        "impact": 3,
        "likelihood": 3,
        "control_indicator": "Compliance Deviation"
    },
    {
        "text": "Pengujian kepatuhan prosedur operasional penyaluran kredit mikro menunjukkan seluruh tahapan verifikasi lapangan telah memadai dan berjalan efektif.",
        "risk_category": "Operational",
        "sentiment": "Positive",
        "impact": 1,
        "likelihood": 2,
        "control_indicator": "Adequate Control"
    },
    {
        "text": "Keterlambatan pengiriman armada logistik pengadaan seragam cabang melewati batas toleransi kontrak kerja sebesar 24 hari kerja.",
        "risk_category": "Operational",
        "sentiment": "Negative",
        "impact": 3,
        "likelihood": 3,
        "control_indicator": "Internal Control Weakness"
    },
    {
        "text": "Pemeriksaan kas kecil dan penutupan rekening operasional cabang seluruhnya terdokumentasi rapi, diverifikasi dual control, serta saldo rekonsiliasi cocok.",
        "risk_category": "Financial",
        "sentiment": "Positive",
        "impact": 1,
        "likelihood": 1,
        "control_indicator": "Adequate Control"
    },
    {
        "text": "Ditemukan akun pengguna sistem ERP milik karyawan yang telah mutasi sejak November masih memiliki hak akses approving purchase order.",
        "risk_category": "Technology",
        "sentiment": "Negative",
        "impact": 4,
        "likelihood": 3,
        "control_indicator": "IT & Access Control Weakness"
    },
    {
        "text": "Hasil audit investigatif mengonfirmasi adanya mark-up biaya renovasi gedung kantor cabang sebesar Rp 380 Juta melibatkan oknum internal dan kontraktor fiktif.",
        "risk_category": "Financial",
        "sentiment": "Negative",
        "impact": 5,
        "likelihood": 4,
        "control_indicator": "Fraud Indication"
    },
    {
        "text": "Pemenuhan laporan profil risiko dan tingkat kesehatan bank ke OJK telah disampaikan tepat waktu sebelum batas akhir pelaporan.",
        "risk_category": "Compliance",
        "sentiment": "Positive",
        "impact": 1,
        "likelihood": 1,
        "control_indicator": "Adequate Control"
    },
    {
        "text": "Proses pencadangan data (backup) harian database transaksi berjalan otomatis dan pengujian disaster recovery center (DRC) sukses dilakukan.",
        "risk_category": "Technology",
        "sentiment": "Positive",
        "impact": 1,
        "likelihood": 1,
        "control_indicator": "Adequate Control"
    }
]


def load_live_nlp_corpus() -> List[Dict[str, Any]]:
    """
    Loads harvested notes & external DB training entries from live_nlp_corpus.json.
    Falls back to in-memory seed corpus if the file does not exist yet.
    """
    if os.path.exists(LIVE_CORPUS_FILE):
        try:
            with open(LIVE_CORPUS_FILE, "r", encoding="utf-8") as f:
                items = json.load(f)
                if isinstance(items, list) and len(items) > 0:
                    return items
        except Exception as e:
            print(f"Warning loading live_nlp_corpus.json: {e}")
    save_live_nlp_corpus(IN_MEMORY_SEED_CORPUS)
    return list(IN_MEMORY_SEED_CORPUS)


def save_live_nlp_corpus(items: List[Dict[str, Any]]) -> None:
    """
    Persists harvested training items to live_nlp_corpus.json.
    """
    try:
        with open(LIVE_CORPUS_FILE, "w", encoding="utf-8") as f:
            json.dump(items[:350], f, indent=2, ensure_ascii=False)
    except Exception as e:
        print(f"Warning saving live_nlp_corpus.json: {e}")


def upsert_live_nlp_corpus(new_items: List[Dict[str, Any]]) -> int:
    """
    Merges newly harvested items into live_nlp_corpus.json, deduplicating by normalized text.
    Returns count of newly added items.
    """
    existing = load_live_nlp_corpus()
    seen = {re.sub(r'\s+', ' ', str(it.get("text", "")).strip().lower()) for it in existing}
    added = 0
    merged = list(existing)

    for item in new_items:
        raw_text = str(item.get("text", "")).strip()
        norm = re.sub(r'\s+', ' ', raw_text.lower())
        if len(norm) >= 12 and norm not in seen:
            seen.add(norm)
            merged.insert(0, item)
            added += 1

    if added > 0:
        save_live_nlp_corpus(merged)
    return added


def harvest_all_live_nlp_training_data() -> List[Dict[str, Any]]:
    """
    Consolidates training data directly from 3 LIVE sources without CSV dependency:
    1. Imported Documents (PDF/Word/Excel extracted text & OCR)
    2. Live Feature Notes & External Database Records (from live_nlp_corpus.json)
    3. In-memory baseline domain seeds (ensures cold-start stability)
    """
    consolidated: List[Dict[str, Any]] = []
    seen_texts = set()

    # 1. Imported Documents (persisted from upload or extraction)
    imported = load_analyzed_documents()
    for doc in imported:
        excerpt = str(doc.get("excerpt") or doc.get("extracted_finding") or "").strip()
        norm = re.sub(r'\s+', ' ', excerpt.lower())
        if len(norm) >= 12 and norm not in seen_texts:
            seen_texts.add(norm)
            cat = normalize_or_resolve_category(doc.get("risk_category") or doc.get("autoCategory")) or "Operational"
            ind = doc.get("control_indicator") or doc.get("controlIndicator") or "Internal Control Weakness"
            consolidated.append({
                "text": excerpt,
                "risk_category": cat,
                "sentiment": doc.get("sentiment", "Negative"),
                "impact": int(doc.get("impact", 3)),
                "likelihood": int(doc.get("likelihood", 3)),
                "control_indicator": ind
            })

    # 2. Live Feature Notes & External DB records
    live_corpus = load_live_nlp_corpus()
    for item in live_corpus:
        raw = str(item.get("text", "")).strip()
        norm = re.sub(r'\s+', ' ', raw.lower())
        if len(norm) >= 12 and norm not in seen_texts:
            seen_texts.add(norm)
            cat = normalize_or_resolve_category(item.get("risk_category")) or "Operational"
            ind = item.get("control_indicator", "Internal Control Weakness")
            consolidated.append({
                "text": raw,
                "risk_category": cat,
                "sentiment": item.get("sentiment", "Negative"),
                "impact": int(item.get("impact", 3)),
                "likelihood": int(item.get("likelihood", 3)),
                "control_indicator": ind
            })

    # 3. Baseline seeds (if not already included)
    for seed in IN_MEMORY_SEED_CORPUS:
        raw = str(seed.get("text", "")).strip()
        norm = re.sub(r'\s+', ' ', raw.lower())
        if norm not in seen_texts:
            seen_texts.add(norm)
            consolidated.append(seed)

    return consolidated

