import type { AnomalyRecord, ScatterPoint } from '~/composables/useAnalyticsData'
import { RiskLevel } from '~/types/risk'

export const BANK_ANOMALIES_SEED: AnomalyRecord[] = [
  {
    "id": "BANK-TRX-0001",
    "entity": "Cabang Medan",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.82,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 02:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-01",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0021",
    "entity": "Kantor Pusat",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.85,
    "description": "Pencairan kredit bernilai Rp 66662.33 Juta dengan bunga murah ekstrem (4.18%) pihak terafiliasi melanggar BMPK",
    "severity": "High",
    "date": "2026-06-21",
    "amount": 66662330000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 66662.33
  },
  {
    "id": "BANK-TRX-0030",
    "entity": "Cabang Bandung",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.94,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-02",
    "amount": 50437180000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 50437.18
  },
  {
    "id": "BANK-TRX-0031",
    "entity": "Cabang Surabaya",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.95,
    "description": "Pencairan kredit bernilai Rp 56330.06 Juta dengan bunga murah ekstrem (4.46%) pihak terafiliasi melanggar BMPK",
    "severity": "Critical",
    "date": "2026-06-03",
    "amount": 56330060000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 56330.06
  },
  {
    "id": "BANK-TRX-0036",
    "entity": "Cabang Bandung",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.83,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "High",
    "date": "2026-06-08",
    "amount": 44640190000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 44640.19
  },
  {
    "id": "BANK-TRX-0057",
    "entity": "Kantor Pusat",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.87,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-01",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 127.73
  },
  {
    "id": "BANK-TRX-0060",
    "entity": "Cabang Jakarta",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.9,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-04",
    "amount": 5713940000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 5713.94
  },
  {
    "id": "BANK-TRX-0066",
    "entity": "Kantor Pusat",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.96,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "Critical",
    "date": "2026-06-10",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 280.24
  },
  {
    "id": "BANK-TRX-0073",
    "entity": "Cabang Bali",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.86,
    "description": "Percobaan login gagal berulang kali (9 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "High",
    "date": "2026-06-17",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 9.0
  },
  {
    "id": "BANK-TRX-0077",
    "entity": "Cabang Bali",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.9,
    "description": "Penarikan dana masif (Rp 20114.34 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-21",
    "amount": 20114340000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 20114.34
  },
  {
    "id": "BANK-TRX-0094",
    "entity": "Cabang Surabaya",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.9,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-10",
    "amount": 20579600000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 20579.6
  },
  {
    "id": "BANK-TRX-0101",
    "entity": "Cabang Medan",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.97,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-17",
    "amount": 62412080000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 62412.08
  },
  {
    "id": "BANK-TRX-0114",
    "entity": "Cabang Bandung",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.93,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-02",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 123.53
  },
  {
    "id": "BANK-TRX-0120",
    "entity": "Cabang Surabaya",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.82,
    "description": "Penarikan dana masif (Rp 29380.63 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "High",
    "date": "2026-06-08",
    "amount": 29380630000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 29380.63
  },
  {
    "id": "BANK-TRX-0126",
    "entity": "Cabang Jakarta",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.88,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 02:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-14",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0143",
    "entity": "Cabang Bandung",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.88,
    "description": "Pencairan kredit bernilai Rp 48011.62 Juta dengan bunga murah ekstrem (4.85%) pihak terafiliasi melanggar BMPK",
    "severity": "High",
    "date": "2026-06-03",
    "amount": 48011620000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 48011.62
  },
  {
    "id": "BANK-TRX-0149",
    "entity": "Cabang Surabaya",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.94,
    "description": "Transaksi pasar uang/valas bernilai Rp 18244.65 Juta dieksekusi di luar jam pasar resmi (Jam 23:00)",
    "severity": "Critical",
    "date": "2026-06-09",
    "amount": 18244650000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 18244.65
  },
  {
    "id": "BANK-TRX-0175",
    "entity": "Cabang Bandung",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.86,
    "description": "Bunga simpanan deposito (10.43%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "High",
    "date": "2026-06-07",
    "amount": 12757990000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 12757.99
  },
  {
    "id": "BANK-TRX-0177",
    "entity": "Cabang Surabaya",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.88,
    "description": "Penarikan dana masif (Rp 23236.21 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "High",
    "date": "2026-06-09",
    "amount": 23236210000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 23236.21
  },
  {
    "id": "BANK-TRX-0180",
    "entity": "Kantor Pusat",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.91,
    "description": "Transfer pembayaran jumlah masif (Rp 2842.62 Juta) pada dini hari (Jam 04:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-12",
    "amount": 2842620000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 2842.62
  },
  {
    "id": "BANK-TRX-0186",
    "entity": "Cabang Jakarta",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.97,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-18",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 161.02
  },
  {
    "id": "BANK-TRX-0196",
    "entity": "Kantor Pusat",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.9,
    "description": "Penarikan dana masif (Rp 27782.5 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-28",
    "amount": 27782500000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 27782.5
  },
  {
    "id": "BANK-TRX-0199",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.93,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "Critical",
    "date": "2026-06-03",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 308.04
  },
  {
    "id": "BANK-TRX-0207",
    "entity": "Cabang Bandung",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.84,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-11",
    "amount": 4820700000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 4820.7
  },
  {
    "id": "BANK-TRX-0211",
    "entity": "Cabang Bali",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.88,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 02:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-15",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0212",
    "entity": "Cabang Medan",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.89,
    "description": "Transfer pembayaran jumlah masif (Rp 3774.99 Juta) pada dini hari (Jam 03:00) tanpa maker-checker",
    "severity": "High",
    "date": "2026-06-16",
    "amount": 3774990000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 3774.99
  },
  {
    "id": "BANK-TRX-0224",
    "entity": "Cabang Bandung",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.84,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-28",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 99.76
  },
  {
    "id": "BANK-TRX-0236",
    "entity": "Cabang Jakarta",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.96,
    "description": "Pencairan kredit bernilai Rp 68917.12 Juta dengan bunga murah ekstrem (4.32%) pihak terafiliasi melanggar BMPK",
    "severity": "Critical",
    "date": "2026-06-12",
    "amount": 68917120000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 68917.12
  },
  {
    "id": "BANK-TRX-0239",
    "entity": "Cabang Medan",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.82,
    "description": "Percobaan login gagal berulang kali (7 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "High",
    "date": "2026-06-15",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 7.0
  },
  {
    "id": "BANK-TRX-0252",
    "entity": "Cabang Surabaya",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.95,
    "description": "Pencairan kredit bernilai Rp 51478.32 Juta dengan bunga murah ekstrem (4.81%) pihak terafiliasi melanggar BMPK",
    "severity": "Critical",
    "date": "2026-06-28",
    "amount": 51478320000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 51478.32
  },
  {
    "id": "BANK-TRX-0268",
    "entity": "Cabang Surabaya",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.94,
    "description": "Transaksi pasar uang/valas bernilai Rp 39778.23 Juta dieksekusi di luar jam pasar resmi (Jam 02:00)",
    "severity": "Critical",
    "date": "2026-06-16",
    "amount": 39778230000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 39778.23
  },
  {
    "id": "BANK-TRX-0271",
    "entity": "Cabang Medan",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.97,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-19",
    "amount": 45762260000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 45762.26
  },
  {
    "id": "BANK-TRX-0289",
    "entity": "Cabang Bandung",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.98,
    "description": "Transfer pembayaran jumlah masif (Rp 599.27 Juta) pada dini hari (Jam 04:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-09",
    "amount": 599270000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 599.27
  },
  {
    "id": "BANK-TRX-0292",
    "entity": "Cabang Jakarta",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.84,
    "description": "Percobaan login gagal berulang kali (13 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "High",
    "date": "2026-06-12",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 13.0
  },
  {
    "id": "BANK-TRX-0302",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.94,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "Critical",
    "date": "2026-06-22",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 344.13
  },
  {
    "id": "BANK-TRX-0304",
    "entity": "Kantor Pusat",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.96,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 23:00) tanpa tiket Change Request",
    "severity": "Critical",
    "date": "2026-06-24",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0324",
    "entity": "Cabang Bandung",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.82,
    "description": "Transaksi pasar uang/valas bernilai Rp 23223.61 Juta dieksekusi di luar jam pasar resmi (Jam 02:00)",
    "severity": "High",
    "date": "2026-06-16",
    "amount": 23223610000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 23223.61
  },
  {
    "id": "BANK-TRX-0344",
    "entity": "Cabang Bali",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.85,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-08",
    "amount": 4085890000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 4085.89
  },
  {
    "id": "BANK-TRX-0350",
    "entity": "Cabang Bali",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.91,
    "description": "Pencairan kredit bernilai Rp 72536.55 Juta dengan bunga murah ekstrem (4.12%) pihak terafiliasi melanggar BMPK",
    "severity": "Critical",
    "date": "2026-06-14",
    "amount": 72536550000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 72536.55
  },
  {
    "id": "BANK-TRX-0355",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.96,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-19",
    "amount": 37044230000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 37044.23
  },
  {
    "id": "BANK-TRX-0359",
    "entity": "Cabang Bali",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.83,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-23",
    "amount": 4231330000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 4231.33
  },
  {
    "id": "BANK-TRX-0403",
    "entity": "Cabang Bandung",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.93,
    "description": "Penarikan dana masif (Rp 21924.88 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-11",
    "amount": 21924880000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 21924.88
  },
  {
    "id": "BANK-TRX-0411",
    "entity": "Cabang Medan",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.84,
    "description": "Transaksi pasar uang/valas bernilai Rp 30751.48 Juta dieksekusi di luar jam pasar resmi (Jam 01:00)",
    "severity": "High",
    "date": "2026-06-19",
    "amount": 30751480000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 30751.48
  },
  {
    "id": "BANK-TRX-0421",
    "entity": "Cabang Bali",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.94,
    "description": "Bunga simpanan deposito (8.83%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "Critical",
    "date": "2026-06-01",
    "amount": 7802580000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 7802.58
  },
  {
    "id": "BANK-TRX-0425",
    "entity": "Cabang Surabaya",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.98,
    "description": "Transaksi pasar uang/valas bernilai Rp 17692.79 Juta dieksekusi di luar jam pasar resmi (Jam 00:00)",
    "severity": "Critical",
    "date": "2026-06-05",
    "amount": 17692790000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 17692.79
  },
  {
    "id": "BANK-TRX-0451",
    "entity": "Cabang Surabaya",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.9,
    "description": "Transfer pembayaran jumlah masif (Rp 2036.08 Juta) pada dini hari (Jam 02:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-03",
    "amount": 2036080000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 2036.08
  },
  {
    "id": "BANK-TRX-0463",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.85,
    "description": "Pencairan kredit bernilai Rp 50448.48 Juta dengan bunga murah ekstrem (4.38%) pihak terafiliasi melanggar BMPK",
    "severity": "High",
    "date": "2026-06-15",
    "amount": 50448480000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 50448.48
  },
  {
    "id": "BANK-TRX-0476",
    "entity": "Cabang Medan",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.98,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-28",
    "amount": 30033020000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 30033.02
  },
  {
    "id": "BANK-TRX-0482",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.87,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "High",
    "date": "2026-06-06",
    "amount": 58839920000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 58839.92
  },
  {
    "id": "BANK-TRX-0489",
    "entity": "Cabang Bali",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.94,
    "description": "Transaksi pasar uang/valas bernilai Rp 16635.36 Juta dieksekusi di luar jam pasar resmi (Jam 02:00)",
    "severity": "Critical",
    "date": "2026-06-13",
    "amount": 16635360000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 16635.36
  },
  {
    "id": "BANK-TRX-0499",
    "entity": "Cabang Bali",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.87,
    "description": "Percobaan login gagal berulang kali (8 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "High",
    "date": "2026-06-23",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 8.0
  },
  {
    "id": "BANK-TRX-0507",
    "entity": "Kantor Pusat",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.95,
    "description": "Penarikan dana masif (Rp 27910.12 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-03",
    "amount": 27910120000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 27910.12
  },
  {
    "id": "BANK-TRX-0509",
    "entity": "Cabang Surabaya",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.97,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-05",
    "amount": 57870520000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 57870.52
  },
  {
    "id": "BANK-TRX-0514",
    "entity": "Kantor Pusat",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.85,
    "description": "Percobaan login gagal berulang kali (11 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "High",
    "date": "2026-06-10",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 11.0
  },
  {
    "id": "BANK-TRX-0535",
    "entity": "Kantor Pusat",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.89,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "High",
    "date": "2026-06-03",
    "amount": 15655030000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 15655.03
  },
  {
    "id": "BANK-TRX-0543",
    "entity": "Kantor Pusat",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.97,
    "description": "Bunga simpanan deposito (9.93%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "Critical",
    "date": "2026-06-11",
    "amount": 10345630000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 10345.63
  },
  {
    "id": "BANK-TRX-0546",
    "entity": "Cabang Medan",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.83,
    "description": "Transaksi pasar uang/valas bernilai Rp 41534.36 Juta dieksekusi di luar jam pasar resmi (Jam 01:00)",
    "severity": "High",
    "date": "2026-06-14",
    "amount": 41534360000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 41534.36
  },
  {
    "id": "BANK-TRX-0547",
    "entity": "Cabang Jakarta",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.84,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 00:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-15",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0558",
    "entity": "Cabang Jakarta",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.95,
    "description": "Percobaan login gagal berulang kali (10 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "Critical",
    "date": "2026-06-26",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 10.0
  },
  {
    "id": "BANK-TRX-0573",
    "entity": "Cabang Medan",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.93,
    "description": "Transaksi pasar uang/valas bernilai Rp 31370.88 Juta dieksekusi di luar jam pasar resmi (Jam 23:00)",
    "severity": "Critical",
    "date": "2026-06-13",
    "amount": 31370880000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 31370.88
  },
  {
    "id": "BANK-TRX-0575",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.95,
    "description": "Transaksi pasar uang/valas bernilai Rp 31337.86 Juta dieksekusi di luar jam pasar resmi (Jam 23:00)",
    "severity": "Critical",
    "date": "2026-06-15",
    "amount": 31337860000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 31337.86
  },
  {
    "id": "BANK-TRX-0577",
    "entity": "Kantor Pusat",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.97,
    "description": "Transfer pembayaran jumlah masif (Rp 621.83 Juta) pada dini hari (Jam 03:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-17",
    "amount": 621830000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 621.83
  },
  {
    "id": "BANK-TRX-0585",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.88,
    "description": "Pencairan kredit bernilai Rp 61846.3 Juta dengan bunga murah ekstrem (4.43%) pihak terafiliasi melanggar BMPK",
    "severity": "High",
    "date": "2026-06-25",
    "amount": 61846300000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 61846.3
  },
  {
    "id": "BANK-TRX-0590",
    "entity": "Cabang Bali",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.93,
    "description": "Transfer pembayaran jumlah masif (Rp 4381.75 Juta) pada dini hari (Jam 02:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-02",
    "amount": 4381750000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 4381.75
  },
  {
    "id": "BANK-TRX-0599",
    "entity": "Kantor Pusat",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.85,
    "description": "Bunga simpanan deposito (10.18%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "High",
    "date": "2026-06-11",
    "amount": 12814220000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 12814.22
  },
  {
    "id": "BANK-TRX-0601",
    "entity": "Kantor Pusat",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.87,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "High",
    "date": "2026-06-13",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 106.24
  },
  {
    "id": "BANK-TRX-0618",
    "entity": "Cabang Bali",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.87,
    "description": "Bunga simpanan deposito (8.09%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "High",
    "date": "2026-06-02",
    "amount": 7182620000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 7182.62
  },
  {
    "id": "BANK-TRX-0628",
    "entity": "Cabang Bandung",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.97,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "Critical",
    "date": "2026-06-12",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 174.38
  },
  {
    "id": "BANK-TRX-0629",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.98,
    "description": "Transaksi pasar uang/valas bernilai Rp 42516.62 Juta dieksekusi di luar jam pasar resmi (Jam 01:00)",
    "severity": "Critical",
    "date": "2026-06-13",
    "amount": 42516620000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 42516.62
  },
  {
    "id": "BANK-TRX-0637",
    "entity": "Cabang Bali",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.89,
    "description": "Bunga simpanan deposito (9.03%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "High",
    "date": "2026-06-21",
    "amount": 8173530000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 8173.53
  },
  {
    "id": "BANK-TRX-0651",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.86,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "High",
    "date": "2026-06-07",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 325.52
  },
  {
    "id": "BANK-TRX-0655",
    "entity": "Kantor Pusat",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.9,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-11",
    "amount": 7452500000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 7452.5
  },
  {
    "id": "BANK-TRX-0659",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.94,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-15",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 310.68
  },
  {
    "id": "BANK-TRX-0661",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.96,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-17",
    "amount": 4635650000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 4635.65
  },
  {
    "id": "BANK-TRX-0663",
    "entity": "Cabang Bandung",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.98,
    "description": "Penarikan dana masif (Rp 27718.34 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-19",
    "amount": 27718340000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 27718.34
  },
  {
    "id": "BANK-TRX-0667",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.85,
    "description": "Transaksi pasar uang/valas bernilai Rp 37145.05 Juta dieksekusi di luar jam pasar resmi (Jam 03:00)",
    "severity": "High",
    "date": "2026-06-23",
    "amount": 37145050000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 37145.05
  },
  {
    "id": "BANK-TRX-0683",
    "entity": "Kantor Pusat",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.84,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "High",
    "date": "2026-06-11",
    "amount": 22490440000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 22490.44
  },
  {
    "id": "BANK-TRX-0687",
    "entity": "Cabang Bali",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.88,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "High",
    "date": "2026-06-15",
    "amount": 11854940000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 11854.94
  },
  {
    "id": "BANK-TRX-0704",
    "entity": "Cabang Jakarta",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.88,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-04",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 277.27
  },
  {
    "id": "BANK-TRX-0706",
    "entity": "Cabang Surabaya",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.9,
    "description": "Pencairan kredit bernilai Rp 57846.49 Juta dengan bunga murah ekstrem (3.96%) pihak terafiliasi melanggar BMPK",
    "severity": "Critical",
    "date": "2026-06-06",
    "amount": 57846490000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 57846.49
  },
  {
    "id": "BANK-TRX-0712",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.96,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-12",
    "amount": 42458520000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 42458.52
  },
  {
    "id": "BANK-TRX-0722",
    "entity": "Cabang Surabaya",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.89,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 01:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-22",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0732",
    "entity": "Cabang Medan",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.82,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-04",
    "amount": 1314040000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 1314.04
  },
  {
    "id": "BANK-TRX-0734",
    "entity": "Cabang Jakarta",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.84,
    "description": "Bunga simpanan deposito (10.77%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "High",
    "date": "2026-06-06",
    "amount": 6434170000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 6434.17
  },
  {
    "id": "BANK-TRX-0759",
    "entity": "Cabang Bandung",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.92,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "Critical",
    "date": "2026-06-03",
    "amount": 1991250000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 1991.25
  },
  {
    "id": "BANK-TRX-0763",
    "entity": "Kantor Pusat",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.96,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-07",
    "amount": 55385050000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 55385.05
  },
  {
    "id": "BANK-TRX-0766",
    "entity": "Cabang Bandung",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.82,
    "description": "Penarikan dana masif (Rp 31614.07 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "High",
    "date": "2026-06-10",
    "amount": 31614070000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 31614.07
  },
  {
    "id": "BANK-TRX-0778",
    "entity": "Cabang Jakarta",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.94,
    "description": "Transfer pembayaran jumlah masif (Rp 2927.89 Juta) pada dini hari (Jam 01:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-22",
    "amount": 2927890000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 2927.89
  },
  {
    "id": "BANK-TRX-0785",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.84,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-01",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 247.07
  },
  {
    "id": "BANK-TRX-0788",
    "entity": "Cabang Jakarta",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.87,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-04",
    "amount": 3785450000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 3785.45
  },
  {
    "id": "BANK-TRX-0801",
    "entity": "Cabang Bali",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.83,
    "description": "Bunga simpanan deposito (8.27%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "High",
    "date": "2026-06-17",
    "amount": 9337730000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 9337.73
  },
  {
    "id": "BANK-TRX-0805",
    "entity": "Cabang Bali",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.87,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-21",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 227.75
  },
  {
    "id": "BANK-TRX-0811",
    "entity": "Cabang Surabaya",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.93,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "Critical",
    "date": "2026-06-27",
    "amount": 1622770000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 1622.77
  },
  {
    "id": "BANK-TRX-0826",
    "entity": "Cabang Surabaya",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.91,
    "description": "Transfer pembayaran jumlah masif (Rp 4219.17 Juta) pada dini hari (Jam 00:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-14",
    "amount": 4219170000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 4219.17
  },
  {
    "id": "BANK-TRX-0832",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.97,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-20",
    "amount": 18416430000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 18416.43
  },
  {
    "id": "BANK-TRX-0850",
    "entity": "Cabang Jakarta",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.98,
    "description": "Bunga simpanan deposito (10.43%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "Critical",
    "date": "2026-06-10",
    "amount": 5926790000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 5926.79
  },
  {
    "id": "BANK-TRX-0855",
    "entity": "Cabang Surabaya",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.86,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-15",
    "amount": 2304420000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 2304.42
  },
  {
    "id": "BANK-TRX-0856",
    "entity": "Cabang Bali",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.87,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 00:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-16",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0864",
    "entity": "Cabang Bali",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.95,
    "description": "Percobaan login gagal berulang kali (10 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "Critical",
    "date": "2026-06-24",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 10.0
  },
  {
    "id": "BANK-TRX-0870",
    "entity": "Cabang Surabaya",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.84,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-02",
    "amount": 2641210000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 2641.21
  },
  {
    "id": "BANK-TRX-0876",
    "entity": "Cabang Jakarta",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.9,
    "description": "Percobaan login gagal berulang kali (6 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "Critical",
    "date": "2026-06-08",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 6.0
  },
  {
    "id": "BANK-TRX-0877",
    "entity": "Cabang Medan",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.91,
    "description": "Penarikan dana masif (Rp 35229.11 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-09",
    "amount": 35229110000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 35229.11
  },
  {
    "id": "BANK-TRX-0882",
    "entity": "Cabang Bandung",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.96,
    "description": "Penarikan dana masif (Rp 34358.7 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-14",
    "amount": 34358699999.999996,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 34358.7
  },
  {
    "id": "BANK-TRX-0887",
    "entity": "Cabang Bali",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.84,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-19",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 202.38
  },
  {
    "id": "BANK-TRX-0900",
    "entity": "Cabang Bali",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.97,
    "description": "Bunga simpanan deposito (8.12%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "Critical",
    "date": "2026-06-04",
    "amount": 2297620000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 2297.62
  },
  {
    "id": "BANK-TRX-0903",
    "entity": "Cabang Bandung",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.83,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "High",
    "date": "2026-06-07",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 218.48
  },
  {
    "id": "BANK-TRX-0906",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.86,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "High",
    "date": "2026-06-10",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 212.93
  },
  {
    "id": "BANK-TRX-0911",
    "entity": "Cabang Bali",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.91,
    "description": "Transfer pembayaran jumlah masif (Rp 1303.51 Juta) pada dini hari (Jam 01:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-15",
    "amount": 1303510000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 1303.51
  },
  {
    "id": "BANK-TRX-0924",
    "entity": "Cabang Jakarta",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.87,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 01:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-28",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0928",
    "entity": "Cabang Medan",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.91,
    "description": "Percobaan login gagal berulang kali (12 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "Critical",
    "date": "2026-06-04",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 12.0
  },
  {
    "id": "BANK-TRX-0932",
    "entity": "Cabang Medan",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.95,
    "description": "Percobaan login gagal berulang kali (12 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "Critical",
    "date": "2026-06-08",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 12.0
  },
  {
    "id": "BANK-TRX-0933",
    "entity": "Cabang Medan",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.96,
    "description": "Penarikan dana masif (Rp 28185.35 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-09",
    "amount": 28185350000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 28185.35
  },
  {
    "id": "BANK-TRX-0939",
    "entity": "Cabang Bandung",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.85,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-15",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 191.04
  },
  {
    "id": "BANK-TRX-0940",
    "entity": "Kantor Pusat",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.86,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 02:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-16",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0943",
    "entity": "Kantor Pusat",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.89,
    "description": "Transaksi pasar uang/valas bernilai Rp 41510.52 Juta dieksekusi di luar jam pasar resmi (Jam 01:00)",
    "severity": "High",
    "date": "2026-06-19",
    "amount": 41510520000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 41510.52
  },
  {
    "id": "BANK-TRX-0946",
    "entity": "Cabang Medan",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.92,
    "description": "Percobaan login gagal berulang kali (8 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "Critical",
    "date": "2026-06-22",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 8.0
  },
  {
    "id": "BANK-TRX-0954",
    "entity": "Cabang Bandung",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.83,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 00:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-02",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0956",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.85,
    "description": "Onboarding nasabah profil Politically Exposed Person (PEP) disetujui tanpa proses Enhanced Due Diligence (EDD)",
    "severity": "High",
    "date": "2026-06-04",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 129.13
  },
  {
    "id": "BANK-TRX-0961",
    "entity": "Kantor Pusat",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.9,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-09",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 253.42
  },
  {
    "id": "BANK-TRX-0967",
    "entity": "Cabang Medan",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.96,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-15",
    "amount": 47058460000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 47058.46
  },
  {
    "id": "BANK-TRX-0975",
    "entity": "Cabang Bandung",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.87,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 02:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-23",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-0999",
    "entity": "Cabang Surabaya",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.94,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-19",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 337.05
  },
  {
    "id": "BANK-TRX-1005",
    "entity": "Cabang Medan",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.83,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "High",
    "date": "2026-06-25",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 183.43
  },
  {
    "id": "BANK-TRX-1006",
    "entity": "Cabang Jakarta",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.84,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "High",
    "date": "2026-06-26",
    "amount": 3823940000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 3823.94
  },
  {
    "id": "BANK-TRX-1008",
    "entity": "Kantor Pusat",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.86,
    "description": "Percobaan login gagal berulang kali (11 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "High",
    "date": "2026-06-28",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 11.0
  },
  {
    "id": "BANK-TRX-1029",
    "entity": "Cabang Bandung",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.9,
    "description": "Penarikan dana masif (Rp 40261.0 Juta) pada rekening simpanan pasif/dormant tanpa otorisasi Branch Manager",
    "severity": "Critical",
    "date": "2026-06-21",
    "amount": 40261000000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 40261.0
  },
  {
    "id": "BANK-TRX-1033",
    "entity": "Cabang Surabaya",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.94,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-25",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 166.0
  },
  {
    "id": "BANK-TRX-1034",
    "entity": "Cabang Medan",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.95,
    "description": "Bunga simpanan deposito (11.49%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "Critical",
    "date": "2026-06-26",
    "amount": 8622490000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 8622.49
  },
  {
    "id": "BANK-TRX-1045",
    "entity": "Cabang Surabaya",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.89,
    "description": "Pencairan kredit bernilai Rp 61467.37 Juta dengan bunga murah ekstrem (4.23%) pihak terafiliasi melanggar BMPK",
    "severity": "High",
    "date": "2026-06-09",
    "amount": 61467370000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 61467.37
  },
  {
    "id": "BANK-TRX-1056",
    "entity": "Cabang Medan",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.83,
    "description": "Aktivitas akses administratif basis data core banking pada dini hari (Jam 23:00) tanpa tiket Change Request",
    "severity": "High",
    "date": "2026-06-20",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 5.0
  },
  {
    "id": "BANK-TRX-1065",
    "entity": "Kantor Pusat",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.92,
    "description": "Transfer pembayaran jumlah masif (Rp 1533.57 Juta) pada dini hari (Jam 01:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-01",
    "amount": 1533570000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 1533.57
  },
  {
    "id": "BANK-TRX-1070",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.97,
    "description": "Transaksi pasar uang/valas bernilai Rp 44032.28 Juta dieksekusi di luar jam pasar resmi (Jam 01:00)",
    "severity": "Critical",
    "date": "2026-06-06",
    "amount": 44032280000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 44032.28
  },
  {
    "id": "BANK-TRX-1085",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.95,
    "description": "Pencairan kredit bernilai Rp 65946.5 Juta dengan bunga murah ekstrem (4.62%) pihak terafiliasi melanggar BMPK",
    "severity": "Critical",
    "date": "2026-06-21",
    "amount": 65946500000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 65946.5
  },
  {
    "id": "BANK-TRX-1086",
    "entity": "Kantor Pusat",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.96,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "Critical",
    "date": "2026-06-22",
    "amount": 3326360000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 3326.36
  },
  {
    "id": "BANK-TRX-1088",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.98,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-24",
    "amount": 7720410000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 7720.41
  },
  {
    "id": "BANK-TRX-1104",
    "entity": "Kantor Pusat",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.97,
    "description": "Transfer pembayaran jumlah masif (Rp 3621.88 Juta) pada dini hari (Jam 01:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-12",
    "amount": 3621880000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 3621.88
  },
  {
    "id": "BANK-TRX-1111",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.87,
    "description": "Pencairan kredit bernilai Rp 66059.57 Juta dengan bunga murah ekstrem (4.42%) pihak terafiliasi melanggar BMPK",
    "severity": "High",
    "date": "2026-06-19",
    "amount": 66059570000.00001,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 66059.57
  },
  {
    "id": "BANK-TRX-1114",
    "entity": "Cabang Surabaya",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.9,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-22",
    "amount": 21034770000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 21034.77
  },
  {
    "id": "BANK-TRX-1115",
    "entity": "Cabang Jakarta",
    "type": "Treasury",
    "category": "Treasury",
    "anomalyScore": 0.91,
    "description": "Transaksi valas dengan deviasi kurs melenceng (Off-Market Rate) melampaui limit dealer tanpa verifikasi Middle Office",
    "severity": "Critical",
    "date": "2026-06-23",
    "amount": 63854490000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 63854.49
  },
  {
    "id": "BANK-TRX-1123",
    "entity": "Cabang Surabaya",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.82,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "High",
    "date": "2026-06-03",
    "amount": 18954620000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 18954.62
  },
  {
    "id": "BANK-TRX-1124",
    "entity": "Cabang Bali",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.83,
    "description": "Bunga simpanan deposito (7.75%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "High",
    "date": "2026-06-04",
    "amount": 4070950000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 4070.95
  },
  {
    "id": "BANK-TRX-1127",
    "entity": "Cabang Bandung",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.86,
    "description": "Transfer pembayaran jumlah masif (Rp 4202.33 Juta) pada dini hari (Jam 03:00) tanpa maker-checker",
    "severity": "High",
    "date": "2026-06-07",
    "amount": 4202330000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 4202.33
  },
  {
    "id": "BANK-TRX-1137",
    "entity": "Cabang Surabaya",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.96,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-17",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 170.53
  },
  {
    "id": "BANK-TRX-1150",
    "entity": "Cabang Surabaya",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.92,
    "description": "Transfer pembayaran jumlah masif (Rp 3565.41 Juta) pada dini hari (Jam 00:00) tanpa maker-checker",
    "severity": "Critical",
    "date": "2026-06-02",
    "amount": 3565410000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 3565.41
  },
  {
    "id": "BANK-TRX-1152",
    "entity": "Cabang Jakarta",
    "type": "IT Control",
    "category": "IT Control",
    "anomalyScore": 0.94,
    "description": "Percobaan login gagal berulang kali (11 kali) pada jam tidak wajar mengindikasikan serangan Brute Force",
    "severity": "Critical",
    "date": "2026-06-04",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 11.0
  },
  {
    "id": "BANK-TRX-1156",
    "entity": "Cabang Jakarta",
    "type": "Payment",
    "category": "Payment",
    "anomalyScore": 0.98,
    "description": "Lonjakan frekuensi transaksi transfer berkecepatan tinggi (Velocity Burst / Indikasi Smurfing Dana)",
    "severity": "Critical",
    "date": "2026-06-08",
    "amount": 3488040000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 3488.04
  },
  {
    "id": "BANK-TRX-1161",
    "entity": "Cabang Medan",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.86,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "High",
    "date": "2026-06-13",
    "amount": 8310200000.000001,
    "isAnomaly": true,
    riskLevel: RiskLevel.MODERATE_HIGH,
    "xMetric": 8310.2
  },
  {
    "id": "BANK-TRX-1172",
    "entity": "Cabang Bandung",
    "type": "Funding",
    "category": "Funding",
    "anomalyScore": 0.97,
    "description": "Bunga simpanan deposito (11.47%) melampaui batas penjaminan LPS tanpa persetujuan komite ALCO",
    "severity": "Critical",
    "date": "2026-06-24",
    "amount": 8460530000.000001,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 8460.53
  },
  {
    "id": "BANK-TRX-1173",
    "entity": "Cabang Surabaya",
    "type": "KYC",
    "category": "KYC",
    "anomalyScore": 0.98,
    "description": "Persetujuan pembukaan rekening nasabah padahal dokumen identitas dan persyaratan APU-PPT Tidak Lengkap",
    "severity": "Critical",
    "date": "2026-06-25",
    "amount": 0.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 175.48
  },
  {
    "id": "BANK-TRX-1185",
    "entity": "Cabang Bandung",
    "type": "Lending",
    "category": "Lending",
    "anomalyScore": 0.93,
    "description": "Pencairan fasilitas kredit mendahului pemenuhan dokumen syarat efektif dan bukti pengikatan agunan",
    "severity": "Critical",
    "date": "2026-06-09",
    "amount": 19746590000.0,
    "isAnomaly": true,
    riskLevel: RiskLevel.HIGH,
    "xMetric": 19746.59
  }
]

export const BANK_SCATTER_SEED: ScatterPoint[] = [
  {
    "x": 0.0,
    "y": 20.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0006"
  },
  {
    "x": 0.0,
    "y": 34.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0041"
  },
  {
    "x": 1.0,
    "y": 31.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0043"
  },
  {
    "x": 0.0,
    "y": 36.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0052"
  },
  {
    "x": 1.0,
    "y": 17.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0055"
  },
  {
    "x": 0.0,
    "y": 30.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0058"
  },
  {
    "x": 0.0,
    "y": 39.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0059"
  },
  {
    "x": 0.0,
    "y": 60.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0081"
  },
  {
    "x": 1.0,
    "y": 44.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0086"
  },
  {
    "x": 0.0,
    "y": 30.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0087"
  },
  {
    "x": 0.0,
    "y": 44.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0089"
  },
  {
    "x": 0.0,
    "y": 41.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0095"
  },
  {
    "x": 1.0,
    "y": 32.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0107"
  },
  {
    "x": 0.0,
    "y": 47.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0109"
  },
  {
    "x": 0.0,
    "y": 42.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0116"
  },
  {
    "x": 2.0,
    "y": 45.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0123"
  },
  {
    "x": 0.0,
    "y": 28.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0124"
  },
  {
    "x": 0.0,
    "y": 55.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0132"
  },
  {
    "x": 0.0,
    "y": 25.5,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0135"
  },
  {
    "x": 0.0,
    "y": 55.0,
    "type": "IT Control",
    "isAnomaly": false,
    "label": "BANK-TRX-0141"
  },
  {
    "x": 5.0,
    "y": 28.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0001"
  },
  {
    "x": 9.0,
    "y": 26.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0073"
  },
  {
    "x": 5.0,
    "y": 33.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0126"
  },
  {
    "x": 5.0,
    "y": 34.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0211"
  },
  {
    "x": 7.0,
    "y": 39.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0239"
  },
  {
    "x": 13.0,
    "y": 34.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0292"
  },
  {
    "x": 5.0,
    "y": 62.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0304"
  },
  {
    "x": 8.0,
    "y": 32.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0499"
  },
  {
    "x": 11.0,
    "y": 40.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0514"
  },
  {
    "x": 5.0,
    "y": 49.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0547"
  },
  {
    "x": 10.0,
    "y": 34.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0558"
  },
  {
    "x": 5.0,
    "y": 27.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0722"
  },
  {
    "x": 5.0,
    "y": 46.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0856"
  },
  {
    "x": 10.0,
    "y": 42.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0864"
  },
  {
    "x": 6.0,
    "y": 40.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0876"
  },
  {
    "x": 5.0,
    "y": 37.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0924"
  },
  {
    "x": 12.0,
    "y": 31.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0928"
  },
  {
    "x": 12.0,
    "y": 33.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0932"
  },
  {
    "x": 5.0,
    "y": 31.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0940"
  },
  {
    "x": 8.0,
    "y": 38.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0946"
  },
  {
    "x": 5.0,
    "y": 48.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0954"
  },
  {
    "x": 5.0,
    "y": 30.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-0975"
  },
  {
    "x": 11.0,
    "y": 42.0,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-1008"
  },
  {
    "x": 5.0,
    "y": 70.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-1056"
  },
  {
    "x": 11.0,
    "y": 37.5,
    "type": "IT Control",
    "isAnomaly": true,
    "label": "BANK-TRX-1152"
  },
  {
    "x": 6.89,
    "y": 36.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0002"
  },
  {
    "x": 10.75,
    "y": 44.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0008"
  },
  {
    "x": 1.91,
    "y": 27.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0010"
  },
  {
    "x": 24.73,
    "y": 31.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0011"
  },
  {
    "x": 20.87,
    "y": 44.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0014"
  },
  {
    "x": 16.08,
    "y": 27.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0019"
  },
  {
    "x": 4.02,
    "y": 26.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0032"
  },
  {
    "x": 22.44,
    "y": 38.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0045"
  },
  {
    "x": 13.93,
    "y": 28.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0079"
  },
  {
    "x": 6.02,
    "y": 34.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0080"
  },
  {
    "x": 17.77,
    "y": 38.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0083"
  },
  {
    "x": 5.42,
    "y": 47.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0098"
  },
  {
    "x": 11.41,
    "y": 37.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0105"
  },
  {
    "x": 0.46,
    "y": 43.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0108"
  },
  {
    "x": 5.24,
    "y": 36.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0110"
  },
  {
    "x": 1.98,
    "y": 33.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0112"
  },
  {
    "x": 21.24,
    "y": 44.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0122"
  },
  {
    "x": 5.98,
    "y": 44.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0140"
  },
  {
    "x": 7.45,
    "y": 28.5,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0155"
  },
  {
    "x": 9.6,
    "y": 38.0,
    "type": "KYC",
    "isAnomaly": false,
    "label": "BANK-TRX-0162"
  },
  {
    "x": 127.73,
    "y": 54.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0057"
  },
  {
    "x": 280.24,
    "y": 45.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0066"
  },
  {
    "x": 123.53,
    "y": 48.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0114"
  },
  {
    "x": 161.02,
    "y": 54.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0186"
  },
  {
    "x": 308.04,
    "y": 44.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0199"
  },
  {
    "x": 99.76,
    "y": 47.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0224"
  },
  {
    "x": 344.13,
    "y": 41.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0302"
  },
  {
    "x": 106.24,
    "y": 38.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0601"
  },
  {
    "x": 174.38,
    "y": 40.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0628"
  },
  {
    "x": 325.52,
    "y": 51.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0651"
  },
  {
    "x": 310.68,
    "y": 53.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0659"
  },
  {
    "x": 277.27,
    "y": 54.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0704"
  },
  {
    "x": 247.07,
    "y": 44.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0785"
  },
  {
    "x": 227.75,
    "y": 46.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0805"
  },
  {
    "x": 202.38,
    "y": 59.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0887"
  },
  {
    "x": 218.48,
    "y": 40.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0903"
  },
  {
    "x": 212.93,
    "y": 51.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0906"
  },
  {
    "x": 191.04,
    "y": 39.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0939"
  },
  {
    "x": 129.13,
    "y": 48.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0956"
  },
  {
    "x": 253.42,
    "y": 41.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0961"
  },
  {
    "x": 337.05,
    "y": 49.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-0999"
  },
  {
    "x": 183.43,
    "y": 52.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-1005"
  },
  {
    "x": 166.0,
    "y": 47.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-1033"
  },
  {
    "x": 170.53,
    "y": 48.0,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-1137"
  },
  {
    "x": 175.48,
    "y": 46.5,
    "type": "KYC",
    "isAnomaly": true,
    "label": "BANK-TRX-1173"
  },
  {
    "x": 206.52,
    "y": 29.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0003"
  },
  {
    "x": 72.05,
    "y": 41.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0005"
  },
  {
    "x": 303.83,
    "y": 34.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0012"
  },
  {
    "x": 267.06,
    "y": 38.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0013"
  },
  {
    "x": 387.16,
    "y": 31.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0016"
  },
  {
    "x": 26.09,
    "y": 28.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0025"
  },
  {
    "x": 280.2,
    "y": 41.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0038"
  },
  {
    "x": 271.1,
    "y": 50.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0042"
  },
  {
    "x": 1287.88,
    "y": 52.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0048"
  },
  {
    "x": 130.97,
    "y": 40.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0049"
  },
  {
    "x": 950.52,
    "y": 25.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0054"
  },
  {
    "x": 152.66,
    "y": 42.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0062"
  },
  {
    "x": 747.78,
    "y": 30.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0064"
  },
  {
    "x": 243.32,
    "y": 41.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0065"
  },
  {
    "x": 236.32,
    "y": 48.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0067"
  },
  {
    "x": 116.86,
    "y": 28.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0072"
  },
  {
    "x": 432.82,
    "y": 33.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0074"
  },
  {
    "x": 134.2,
    "y": 44.5,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0075"
  },
  {
    "x": 283.83,
    "y": 30.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0078"
  },
  {
    "x": 428.52,
    "y": 42.0,
    "type": "Funding",
    "isAnomaly": false,
    "label": "BANK-TRX-0084"
  },
  {
    "x": 20114.34,
    "y": 53.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0077"
  },
  {
    "x": 29380.63,
    "y": 60.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0120"
  },
  {
    "x": 12757.99,
    "y": 50.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0175"
  },
  {
    "x": 23236.21,
    "y": 46.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0177"
  },
  {
    "x": 27782.5,
    "y": 47.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0196"
  },
  {
    "x": 21924.88,
    "y": 50.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0403"
  },
  {
    "x": 7802.58,
    "y": 38.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0421"
  },
  {
    "x": 27910.12,
    "y": 51.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0507"
  },
  {
    "x": 10345.63,
    "y": 45.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0543"
  },
  {
    "x": 12814.22,
    "y": 47.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0599"
  },
  {
    "x": 7182.62,
    "y": 45.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0618"
  },
  {
    "x": 8173.53,
    "y": 41.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0637"
  },
  {
    "x": 27718.34,
    "y": 39.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0663"
  },
  {
    "x": 6434.17,
    "y": 47.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0734"
  },
  {
    "x": 31614.07,
    "y": 58.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0766"
  },
  {
    "x": 9337.73,
    "y": 48.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0801"
  },
  {
    "x": 5926.79,
    "y": 52.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0850"
  },
  {
    "x": 35229.11,
    "y": 40.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0877"
  },
  {
    "x": 34358.7,
    "y": 52.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0882"
  },
  {
    "x": 2297.62,
    "y": 54.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0900"
  },
  {
    "x": 28185.35,
    "y": 49.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-0933"
  },
  {
    "x": 40261.0,
    "y": 45.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-1029"
  },
  {
    "x": 8622.49,
    "y": 38.0,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-1034"
  },
  {
    "x": 4070.95,
    "y": 51.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-1124"
  },
  {
    "x": 8460.53,
    "y": 51.5,
    "type": "Funding",
    "isAnomaly": true,
    "label": "BANK-TRX-1172"
  },
  {
    "x": 1649.0,
    "y": 30.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0004"
  },
  {
    "x": 4802.35,
    "y": 45.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0009"
  },
  {
    "x": 1436.13,
    "y": 47.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0017"
  },
  {
    "x": 1670.17,
    "y": 38.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0020"
  },
  {
    "x": 2551.48,
    "y": 22.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0028"
  },
  {
    "x": 2241.43,
    "y": 32.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0037"
  },
  {
    "x": 4070.67,
    "y": 29.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0068"
  },
  {
    "x": 1413.83,
    "y": 40.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0069"
  },
  {
    "x": 1029.33,
    "y": 31.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0070"
  },
  {
    "x": 7640.05,
    "y": 29.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0071"
  },
  {
    "x": 4260.0,
    "y": 28.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0076"
  },
  {
    "x": 3852.3,
    "y": 40.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0085"
  },
  {
    "x": 3118.12,
    "y": 30.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0091"
  },
  {
    "x": 2048.49,
    "y": 36.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0092"
  },
  {
    "x": 1353.63,
    "y": 41.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0097"
  },
  {
    "x": 3071.71,
    "y": 28.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0121"
  },
  {
    "x": 5219.83,
    "y": 37.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0125"
  },
  {
    "x": 2359.36,
    "y": 39.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0138"
  },
  {
    "x": 2771.85,
    "y": 25.5,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0139"
  },
  {
    "x": 1134.9,
    "y": 26.0,
    "type": "Treasury",
    "isAnomaly": false,
    "label": "BANK-TRX-0146"
  },
  {
    "x": 50437.18,
    "y": 52.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0030"
  },
  {
    "x": 44640.19,
    "y": 52.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0036"
  },
  {
    "x": 62412.08,
    "y": 45.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0101"
  },
  {
    "x": 18244.65,
    "y": 63.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0149"
  },
  {
    "x": 39778.23,
    "y": 31.0,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0268"
  },
  {
    "x": 45762.26,
    "y": 53.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0271"
  },
  {
    "x": 23223.61,
    "y": 39.0,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0324"
  },
  {
    "x": 37044.23,
    "y": 44.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0355"
  },
  {
    "x": 30751.48,
    "y": 28.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0411"
  },
  {
    "x": 17692.79,
    "y": 47.0,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0425"
  },
  {
    "x": 30033.02,
    "y": 54.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0476"
  },
  {
    "x": 58839.92,
    "y": 45.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0482"
  },
  {
    "x": 16635.36,
    "y": 36.0,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0489"
  },
  {
    "x": 57870.52,
    "y": 44.0,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0509"
  },
  {
    "x": 41534.36,
    "y": 31.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0546"
  },
  {
    "x": 31370.88,
    "y": 67.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0573"
  },
  {
    "x": 31337.86,
    "y": 69.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0575"
  },
  {
    "x": 42516.62,
    "y": 30.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0629"
  },
  {
    "x": 37145.05,
    "y": 35.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0667"
  },
  {
    "x": 42458.52,
    "y": 47.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0712"
  },
  {
    "x": 55385.05,
    "y": 44.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0763"
  },
  {
    "x": 41510.52,
    "y": 32.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0943"
  },
  {
    "x": 47058.46,
    "y": 47.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-0967"
  },
  {
    "x": 44032.28,
    "y": 27.5,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-1070"
  },
  {
    "x": 63854.49,
    "y": 50.0,
    "type": "Treasury",
    "isAnomaly": true,
    "label": "BANK-TRX-1115"
  },
  {
    "x": 2037.7,
    "y": 28.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0007"
  },
  {
    "x": 548.26,
    "y": 30.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0015"
  },
  {
    "x": 180.62,
    "y": 42.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0024"
  },
  {
    "x": 876.17,
    "y": 29.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0026"
  },
  {
    "x": 1702.35,
    "y": 33.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0029"
  },
  {
    "x": 3781.2,
    "y": 39.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0035"
  },
  {
    "x": 1048.98,
    "y": 34.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0039"
  },
  {
    "x": 1610.03,
    "y": 47.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0044"
  },
  {
    "x": 187.61,
    "y": 40.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0046"
  },
  {
    "x": 3266.71,
    "y": 41.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0047"
  },
  {
    "x": 5076.84,
    "y": 24.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0050"
  },
  {
    "x": 163.87,
    "y": 39.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0053"
  },
  {
    "x": 1878.18,
    "y": 28.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0056"
  },
  {
    "x": 167.76,
    "y": 43.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0061"
  },
  {
    "x": 948.67,
    "y": 30.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0063"
  },
  {
    "x": 156.29,
    "y": 32.5,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0082"
  },
  {
    "x": 942.6,
    "y": 37.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0093"
  },
  {
    "x": 1830.12,
    "y": 48.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0099"
  },
  {
    "x": 2499.24,
    "y": 22.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0111"
  },
  {
    "x": 2482.84,
    "y": 43.0,
    "type": "Lending",
    "isAnomaly": false,
    "label": "BANK-TRX-0117"
  },
  {
    "x": 66662.33,
    "y": 52.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0021"
  },
  {
    "x": 56330.06,
    "y": 50.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0031"
  },
  {
    "x": 5713.94,
    "y": 57.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0060"
  },
  {
    "x": 20579.6,
    "y": 58.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0094"
  },
  {
    "x": 48011.62,
    "y": 57.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0143"
  },
  {
    "x": 68917.12,
    "y": 48.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0236"
  },
  {
    "x": 51478.32,
    "y": 49.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0252"
  },
  {
    "x": 72536.55,
    "y": 41.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0350"
  },
  {
    "x": 50448.48,
    "y": 50.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0463"
  },
  {
    "x": 15655.03,
    "y": 43.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0535"
  },
  {
    "x": 61846.3,
    "y": 48.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0585"
  },
  {
    "x": 7452.5,
    "y": 46.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0655"
  },
  {
    "x": 4635.65,
    "y": 44.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0661"
  },
  {
    "x": 22490.44,
    "y": 53.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0683"
  },
  {
    "x": 11854.94,
    "y": 51.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0687"
  },
  {
    "x": 57846.49,
    "y": 49.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0706"
  },
  {
    "x": 18416.43,
    "y": 41.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-0832"
  },
  {
    "x": 61467.37,
    "y": 38.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1045"
  },
  {
    "x": 65946.5,
    "y": 45.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1085"
  },
  {
    "x": 7720.41,
    "y": 50.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1088"
  },
  {
    "x": 66059.57,
    "y": 49.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1111"
  },
  {
    "x": 21034.77,
    "y": 50.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1114"
  },
  {
    "x": 18954.62,
    "y": 55.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1123"
  },
  {
    "x": 8310.2,
    "y": 45.0,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1161"
  },
  {
    "x": 19746.59,
    "y": 49.5,
    "type": "Lending",
    "isAnomaly": true,
    "label": "BANK-TRX-1185"
  },
  {
    "x": 78.92,
    "y": 38.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0018"
  },
  {
    "x": 55.95,
    "y": 35.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0022"
  },
  {
    "x": 2.76,
    "y": 31.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0023"
  },
  {
    "x": 47.51,
    "y": 48.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0027"
  },
  {
    "x": 14.28,
    "y": 42.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0033"
  },
  {
    "x": 84.64,
    "y": 31.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0034"
  },
  {
    "x": 13.97,
    "y": 23.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0040"
  },
  {
    "x": 2.16,
    "y": 42.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0051"
  },
  {
    "x": 18.54,
    "y": 50.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0090"
  },
  {
    "x": 23.16,
    "y": 42.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0096"
  },
  {
    "x": 95.71,
    "y": 43.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0103"
  },
  {
    "x": 22.57,
    "y": 61.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0106"
  },
  {
    "x": 98.42,
    "y": 28.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0115"
  },
  {
    "x": 1.54,
    "y": 26.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0119"
  },
  {
    "x": 3.8,
    "y": 49.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0129"
  },
  {
    "x": 61.55,
    "y": 55.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0130"
  },
  {
    "x": 10.4,
    "y": 50.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0136"
  },
  {
    "x": 1.48,
    "y": 16.0,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0137"
  },
  {
    "x": 32.88,
    "y": 33.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0142"
  },
  {
    "x": 102.75,
    "y": 60.5,
    "type": "Payment",
    "isAnomaly": false,
    "label": "BANK-TRX-0144"
  },
  {
    "x": 2842.62,
    "y": 42.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0180"
  },
  {
    "x": 4820.7,
    "y": 45.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0207"
  },
  {
    "x": 3774.99,
    "y": 36.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0212"
  },
  {
    "x": 599.27,
    "y": 31.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0289"
  },
  {
    "x": 4085.89,
    "y": 50.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0344"
  },
  {
    "x": 4231.33,
    "y": 62.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0359"
  },
  {
    "x": 2036.08,
    "y": 34.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0451"
  },
  {
    "x": 621.83,
    "y": 29.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0577"
  },
  {
    "x": 4381.75,
    "y": 29.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0590"
  },
  {
    "x": 1314.04,
    "y": 49.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0732"
  },
  {
    "x": 1991.25,
    "y": 52.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0759"
  },
  {
    "x": 2927.89,
    "y": 35.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0778"
  },
  {
    "x": 3785.45,
    "y": 57.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0788"
  },
  {
    "x": 1622.77,
    "y": 56.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0811"
  },
  {
    "x": 4219.17,
    "y": 52.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0826"
  },
  {
    "x": 2304.42,
    "y": 51.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0855"
  },
  {
    "x": 2641.21,
    "y": 51.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0870"
  },
  {
    "x": 1303.51,
    "y": 36.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-0911"
  },
  {
    "x": 3823.94,
    "y": 61.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-1006"
  },
  {
    "x": 1533.57,
    "y": 34.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-1065"
  },
  {
    "x": 3326.36,
    "y": 58.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-1086"
  },
  {
    "x": 3621.88,
    "y": 37.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-1104"
  },
  {
    "x": 4202.33,
    "y": 39.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-1127"
  },
  {
    "x": 3565.41,
    "y": 52.0,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-1150"
  },
  {
    "x": 3488.04,
    "y": 47.5,
    "type": "Payment",
    "isAnomaly": true,
    "label": "BANK-TRX-1156"
  }
,
  {
    "x": 24.46,
    "y": 32.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0001"
  },
  {
    "x": 66.84,
    "y": 38.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0003"
  },
  {
    "x": 9.48,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0006"
  },
  {
    "x": 3.99,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0007"
  },
  {
    "x": 46.95,
    "y": 35.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0009"
  },
  {
    "x": 12.93,
    "y": 21.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0014"
  },
  {
    "x": 11.03,
    "y": 49.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0015"
  },
  {
    "x": 19.14,
    "y": 40.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0017"
  },
  {
    "x": 38.2,
    "y": 31.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0018"
  },
  {
    "x": 29.28,
    "y": 26.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0019"
  },
  {
    "x": 18.21,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0020"
  },
  {
    "x": 8.51,
    "y": 35.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0022"
  },
  {
    "x": 18.28,
    "y": 56.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0023"
  },
  {
    "x": 31.45,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0025"
  },
  {
    "x": 12.14,
    "y": 56.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0027"
  },
  {
    "x": 37.1,
    "y": 61.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0028"
  },
  {
    "x": 45.88,
    "y": 35.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-NORM-0029"
  },
  {
    "x": 12.5,
    "y": 43.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-GL-002"
  },
  {
    "x": 35.0,
    "y": 47.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-GL-004"
  },
  {
    "x": 54.2,
    "y": 36.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "TRX-GL-005"
  },
  {
    "x": 24.46,
    "y": 32.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0001"
  },
  {
    "x": 66.84,
    "y": 38.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0003"
  },
  {
    "x": 9.48,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0006"
  },
  {
    "x": 3.99,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0007"
  },
  {
    "x": 46.95,
    "y": 35.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0009"
  },
  {
    "x": 12.93,
    "y": 21.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0014"
  },
  {
    "x": 11.03,
    "y": 49.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0015"
  },
  {
    "x": 19.14,
    "y": 40.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0017"
  },
  {
    "x": 38.2,
    "y": 31.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0018"
  },
  {
    "x": 29.28,
    "y": 26.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0019"
  },
  {
    "x": 18.21,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0020"
  },
  {
    "x": 8.51,
    "y": 35.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0022"
  },
  {
    "x": 18.28,
    "y": 56.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0023"
  },
  {
    "x": 31.45,
    "y": 44.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0025"
  },
  {
    "x": 12.14,
    "y": 56.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0027"
  },
  {
    "x": 37.1,
    "y": 61.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0028"
  },
  {
    "x": 45.88,
    "y": 35.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-NORM-0029"
  },
  {
    "x": 12.5,
    "y": 43.5,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-TRX-002"
  },
  {
    "x": 35.0,
    "y": 47.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-TRX-004"
  },
  {
    "x": 54.2,
    "y": 36.0,
    "type": "Transaction",
    "isAnomaly": false,
    "label": "GL-TRX-005"
  }
]
