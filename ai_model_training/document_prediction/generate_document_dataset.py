import os
import numpy as np
import pandas as pd

np.random.seed(42)

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
N_SAMPLES = 1400

# 7 Risk Categories
CATEGORIES = ['Financial', 'Operations', 'Technology', 'Human Resources', 'Governance', 'Compliance', 'Strategic']

# Templates for text generation across categories and sentiments
TEXT_TEMPLATES = {
    'Financial': {
        'Negative': [
            "Ditemukan kelemahan signifikan dalam prosedur otorisasi pembayaran kas besar senilai Rp {amt} Miliar tanpa persetujuan bertingkat.",
            "Terdapat selisih kas fisik sebesar Rp {amt} Juta pada rekonsiliasi bank bulanan yang belum dapat dijelaskan oleh tim akuntansi.",
            "Prosedur pencatatan utang usaha terlambat hingga 45 hari, berpotensi menimbulkan klaim penalti vendor dan denda administrasi.",
            "Laporan piutang ragu-ragu tidak memperhitungkan cadangan kerugian penurunan nilai secara akurat sesuai standar PSAK.",
            "Terjadi pencatatan ganda atas klaim pengeluaran operasional cabang senilai Rp {amt} Juta karena tidak adanya validasi sistemic."
        ],
        'Positive': [
            "Prosedur rekonsiliasi keuangan dan kas harian telah berjalan sangat tertib dan sesuai dengan SOP standar manajemen risiko keuangan.",
            "Sistem otomatisasi pencatatan faktur pajak dan transaksi kas berhasil menekan angka kesalahan akuntansi hingga 0.01%.",
            "Pengelolaan anggaran belanja operasional cabang berhasil mencapai efisiensi biaya sebesar 12% tanpa mengurangi kualitas layanan.",
            "Verifikasi dokumen pembayaran kas besar telah menerapkan otorisasi bertingkat yang transparan dan terdokumentasi dengan baik.",
            "Proses audit internal keuangan semesteran menunjukkan kepatuhan penuh terhadap kebijakan pembukuan perusahaan."
        ],
        'Neutral': [
            "Telah dilakukan evaluasi berkala terhadap laporan arus kas bulanan unit usaha untuk periode kuartal II tahun berjalan.",
            "Tim audit mencatat implementasi modul akuntansi baru pada sistem ERP utama yang mulai berlaku efektif sejak bulan lalu.",
            "Laporan posisi kas harian telah disampaikan kepada jajaran direksi keuangan secara rutin sesuai jadwal operasional.",
            "Dilakukan penyesuaian format penyajian laporan keuangan konsolidasi untuk menyesuaikan standar pelaporan induk perusahaan.",
            "Manajemen telah memfinalisasi pemutakhiran bagan akun standar (chart of accounts) untuk seluruh cabang operasional."
        ]
    },
    'Operations': {
        'Negative': [
            "Jadwal pemeliharaan rutin mesin produksi utama terlambat 3 bulan, menyebabkan penurunan kapasitas produksi hingga 25%.",
            "Pengelolaan stok persediaan barang jadi di gudang utama mengalami selisih fisik sejumlah {amt} unit tanpa catatan penyesuaian.",
            "Prosedur penanganan keluhan pelanggan belum memiliki SLA yang jelas, mengakibatkan tingkat kepuasan nasabah turun.",
            "Terdapat kecenderungan keterlambatan pengiriman barang ke jaringan distributor hingga melebihi batas toleransi 5 hari.",
            "Sistem manajemen rantai pasok belum terintegrasi dengan vendor utama, memicu akumulasi *bottleneck* persediaan."
        ],
        'Positive': [
            "Proses pemeliharaan berkala fasilitas pabrik berjalan tepat waktu dan meningkatkan efisiensi penggunaan bahan bakar hingga 8%.",
            "Tingkat ketepatan pengiriman barang kepada pelanggan mencapai 99.2% sepanjang semester I, melebihi target KPI yang ditetapkan.",
            "Penerapan standar K3 di area operasional gudang berhasil mempertahankan pencapaian *zero accident* selama 3 tahun berturut-turut.",
            "Sistem pelacakan persediaan berbasis *barcode* beroperasi optimal dan memangkas durasi *stock opname* hingga 50%.",
            "Penyelesaian tiket keluhan pelanggan berhasil dituntaskan dalam kurun waktu kurang dari 24 jam secara terukur."
        ],
        'Neutral': [
            "Pemeriksaan rutin terhadap alur proses distribusi barang antar gudang cabang telah diselesaikan sesuai jadwal kerja audit.",
            "Manajemen operasional menyampaikan pemutakhiran panduan prosedur standar operasional (SOP) distribusi logistik.",
            "Telah dilaksanakan simulasi tanggap darurat keselamatan kerja di area pabrik bersama tim K3 internal.",
            "Laporan kinerja operasional bulanan menunjukkan volume output produksi yang stabil dibandingkan periode sebelumnya.",
            "Pihak pengelola gudang sedang melakukan penataan ulang tata letak rak barang untuk mengoptimalkan ruang penyimpanan."
        ]
    },
    'Technology': {
        'Negative': [
            "Hasil tes penetrasi mengungkapkan 3 kerentanan kritis (SQL Injection & RCE) pada aplikasi e-banking yang belum diperbaiki.",
            "Prosedur *backup* basis data server utama tidak berhasil dijalankan secara penuh selama 2 minggu berturut-turut.",
            "Akses akun administrator sistem inti tidak menerapkan otentikasi ganda (2FA), berisiko tinggi terhadap akses tidak sah.",
            "Sistem pemulihan bencana (*Disaster Recovery Plan*) belum pernah diuji cobakan secara menyeluruh dalam 18 bulan terakhir.",
            "Ditemukan insiden kebocoran kredensial pengguna akibat lemahnya enkripsi pada lalu lintas data API microservice."
        ],
        'Positive': [
            "Implementasi arsitektur keamanan *Zero Trust* pada jaringan internal berhasil menggagalkan 100% ancaman *cyber attack* external.",
            "Proses migrasi server basis data ke infrastruktur *cloud* berjalan lancar tanpa mengalami *downtime* operasional.",
            "Pembaruan patch keamanan pada seluruh *endpoint* komputer karyawan telah tuntas 100% sesuai kebijakan TI.",
            "Sistem pemantauan server berbasis AI secara proaktif mendeteksi dan mengisolasi potensi gangguan performa secara otomatis.",
            "Uji keandalan pemulihan bencana (*DRP drill*) berhasil memulihkan seluruh sistem utama dalam durasi kurang dari 15 menit."
        ],
        'Neutral': [
            "Tim IT telah menyelesaikan pemutakhiran versi perangkat lunak server web ke versi stabil terbaru.",
            "Telah dilakukan pemetaan alokasi kapasitas *bandwidth* jaringan internet untuk mendukung kegiatan operasional kantor cabang.",
            "Laporan pemantauan statistik penggunaan pemrosesan CPU server utama mencatat rata-rata beban kerja sebesar 45%.",
            "Manajemen TI mempublikasikan dokumen pedoman penggunaan aset komputer dan perangkat keras bagi pegawai baru.",
            "Pemeriksaan berkala atas daftar inventaris lisensi perangkat lunak perusahaan telah dilaksanakan sesuai jadwal."
        ]
    },
    'Human Resources': {
        'Negative': [
            "Tingkat perputaran (*turnover*) karyawan pada divisi IT mencapai 32% per tahun, mengganggu kelancaran proyek-proyek strategis.",
            "Proses rekrutmen tidak menjalankan verifikasi latar belakang (*background check*) secara lengkap untuk posisi manajerial kunci.",
            "Ditemukan ketidaksesuaian perhitungan jam lembur pada 45 karyawan cabang yang berpotensi memicu perselisihan hubungan industrial.",
            "Program pelatihan wajib K3 dan *cybersecurity awareness* baru diikuti oleh kurang dari 40% total pegawai perusahaan.",
            "Dokumen kontrak kerja bagi 20 karyawan kontrak belum ditandatangani secara sah meskipun masa kerja telah berjalan 3 bulan."
        ],
        'Positive': [
            "Program pengembangan talenta dan *succession planning* berhasil mengisi 85% posisi kepemimpinan dari internal perusahaan.",
            "Hasil survei kepuasan dan keterikatan pegawai (*employee engagement score*) meningkat signifikan menjadi 88.5%.",
            "Pelaksanaan evaluasi kinerja berkala (KPI review) telah dilakukan 100% tepat waktu dengan umpan balik transparan.",
            "Program pelatihan keselamatan dan kesehatan kerja berhasil mencapai tingkat kelulusan peserta sebesar 98%.",
            "Sistem penggajian otomatis (*payroll system*) beroperasi tanpa kendala dan memproses hak karyawan secara presisi."
        ],
        'Neutral': [
            "Tim HRD menyampaikan laporan rekapitulasi data keikutsertaan karyawan dalam program asuransi kesehatan tahunan.",
            "Telah dilaksanakan pemutakhiran data mandiri karyawan pada portal *employee self-service* untuk pembaruan alamat.",
            "Manajemen sumber daya manusia merilis kalender kegiatan pelatihan dan pengembangan keterampilan untuk semester depan.",
            "Laporan rasio kecukupan tenaga kerja pada unit bisnis operasional mencatat kondisi pemenuhan yang stabil.",
            "Proses administrasi pengajuan cuti tahunan pegawai berjalan sesuai regulasi internal yang berlaku."
        ]
    },
    'Governance': {
        'Negative': [
            "Komite audit belum menyelenggarakan rapat berkala kuartalan sesuai ketentuan tata kelola perusahaan yang baik (GCG).",
            "Kebijakan penanganan benturan kepentingan (*conflict of interest*) tidak diperbarui dalam 5 tahun terakhir.",
            "Dokumen risalah rapat Direksi dan Dewan Komisaris tidak terdokumentasi secara rapi dan sulit diakses saat audit.",
            "Prosedur pengungkapan informasi transaksi pihak berelasi tidak memenuhi standar keterbukaan informasi yang dipersyaratkan.",
            "Terdapat insiden pelanggaran pedoman etika bisnis oleh oknum manajemen yang tidak ditindaklanjuti secara tegas."
        ],
        'Positive': [
            "Penerapan prinsip *Good Corporate Governance* (GCG) meraih penilaian predikat 'Sangat Baik' dari lembaga pemeringkat independen.",
            "Dewan Komisaris dan Direksi secara konsisten menyelenggarakan rapat evaluasi kinerja strategis setiap bulan.",
            "Kebijakan *whistleblowing system* dikelola secara independen oleh pihak ketiga dan menjamin perlindungan pelapor 100%.",
            "Struktur komite pendukung tata kelola telah berfungsi efektif dalam memberikan rekomendasi independen bagi manajemen.",
            "Seluruh jajaran manajemen dan pegawai telah menandatangani pakta integritas dan komitmen anti-korupsi secara berkala."
        ],
        'Neutral': [
            "Telah dilaksanakan pemutakhiran struktur organisasi dan tata kerja pada tingkat direktorat operasional.",
            "Sekretariat perusahaan menyampaikan laporan tahunan pelaksanaan kegiatan tata kelola kepada otoritas terkait.",
            "Manajemen meninjau ulang susunan piagam (*charter*) audit internal sesuai dengan perkembangan standar profesi audit.",
            "Rapat umum pemegang saham telah mengesahkan laporan keuangan audited untuk tahun buku yang lalu.",
            "Dilakukan dokumentasi berkala atas matriks kewenangan keputusanjatan di lingkungan jajaran eksekutif."
        ]
    },
    'Compliance': {
        'Negative': [
            "Pelaporan kewajiban transaksi keuangan mencurigakan (LTKM) ke PPATK mengalami keterlambatan melebihi batas waktu 3 hari.",
            "Perusahaan belum mengimplementasikan penyesuaian regulasi perlindungan data pribadi (UU PDP) secara menyeluruh.",
            "Ditemukan 4 kantor cabang pembantu yang beroperasi tanpa kelengkapan perizinan usaha daerah yang valid.",
            "Prosedur *Anti-Money Laundering* (APU-PPT) tidak melakukan pemutakhiran profil risiko nasabah secara berkala.",
            "Terdapat ketidakpatuhan terhadap batas maksimum pemberian kredit (BMPK) pada 2 debitur grup usaha terkait."
        ],
        'Positive': [
            "Hasil penilaian kepatuhan dari regulator menunjukkan angka kepatuhan 100% tanpa adanya catatan pelanggaran.",
            "Sistem pemantauan kepatuhan regulasi secara proaktif menyesuaikan seluruh prosedur internal dengan aturan OJK terbaru.",
            "Seluruh kewajiban pelaporan berkala kepada otoritas keuangan diselesaikan tepat waktu dengan akurasi data yang baik.",
            "Implementasi program edukasi anti-pencucian uang diikuti oleh 100% pegawai unit kerja bisnis dan operasional.",
            "Prosedur mitigasi risiko hukum berhasil menyelesaikan seluruh potensi klaim sengketa secara damai."
        ],
        'Neutral': [
            "Tim *compliance* telah menyelesaikan inventarisasi ketentuan dan peraturan perundang-undangan terbaru yang relevan.",
            "Telah disampaikan laporan pemantauan kepatuhan triwulanan kepada Direktur Kepatuhan dan Komite Risiko.",
            "Manajemen menyelenggarakan sesi sosialisasi aturan regulasi baru yang diterbitkan oleh instansi pembina sektor usaha.",
            "Dilakukan evaluasi rutin terhadap kelengkapan berkas perizinan operasional di seluruh unit kerja perusahaan.",
            "Pemeriksaan berkala atas tingkat kepatuhan pelaksanaan prosedur operasional standar diselesaikan sesuai rencana kerja."
        ]
    },
    'Strategic': {
        'Negative': [
            "Proyek ekspansi pembukaan 10 cabang baru mengalami penundaan 9 bulan dan pembengkakan anggaran mencapai 40%.",
            "Strategi peluncuran produk digital baru tidak didukung oleh riset pasar yang memadai, mengakibatkan adopsi pengguna amat rendah.",
            "Kemitraan strategis dengan penyedia teknologi utama dihentikan sepihak karena ketidakmampuan mencapai target kerjasama.",
            "Transformasi sistem bisnis inti mengalami kendala integrasi berat yang mengancam pencapaian target jangka panjang.",
            "Rencana diversifikasi portofolio investasi mengalami kerugian akibat lemahnya analisis mitigasi risiko pasar."
        ],
        'Positive': [
            "Pencapaian target pendapatan dari segmen bisnis baru melampaui rencana strategis perusahaan hingga 115%.",
            "Peluncuran inovasi layanan berbasis aplikasi seluler berhasil menggaet 200,000 pengguna aktif baru dalam 3 bulan.",
            "Kerjasama strategis dengan jaringan ritel nasional berhasil memperluas pangsa pasar produk perusahaan secara signifikan.",
            "Eksekusi rencana investasi teknologi informasi berjalan sesuai anggaran dan tepat waktu (*on time and on budget*).",
            "Transformasi budaya digital perusahaan menunjukkan dampak positif pada efisiensi biaya dan kecepatan inovasi."
        ],
        'Neutral': [
            "Manajemen strategis menyelenggarakan lokakarya peninjauan ulang rencana jangka panjang perusahaan (RJPP).",
            "Telah dilakukan pemetaan analisis SWOT berkala untuk mengidentifikasi peluang dan tantangan bisnis mendatang.",
            "Laporan pencapaian inisiatif strategis kuartal II dipresentasikan dalam rapat koordinasi pimpinan tingkat atas.",
            "Tim perencanaan bisnis menyusun studi kelayakan awal untuk rencana pengembangan area operasional baru.",
            "Dilakukan evaluasi terhadap indikator pencapaian sasaran strategis tahunan pada masing-masing unit bisnis."
        ]
    }
}

# Generate 1400 balanced text samples
rows = []
category_counts = {c: 0 for c in CATEGORIES}

for i in range(N_SAMPLES):
    cat = CATEGORIES[i % len(CATEGORIES)]

    # Sentiments: 50% Negative, 25% Positive, 25% Neutral
    p_sent = np.random.choice(['Negative', 'Positive', 'Neutral'], p=[0.50, 0.25, 0.25])

    # Pick template
    template_list = TEXT_TEMPLATES[cat][p_sent]
    template = np.random.choice(template_list)

    # Format template if amount placeholder exists
    amt_val = np.round(np.random.uniform(10, 800), 1)
    text = template.replace('{amt}', str(amt_val))
    text_with_num = f"{i+1}. {text}"

    # Impact (1-5): Driven by domain severity & sentiment, independent from Likelihood
    if p_sent == 'Negative':
        impact = int(np.random.choice([3, 4, 5], p=[0.20, 0.50, 0.30]))
    elif p_sent == 'Positive':
        impact = int(np.random.choice([1, 2], p=[0.60, 0.40]))
    else:  # Neutral
        impact = int(np.random.choice([2, 3, 4], p=[0.30, 0.50, 0.20]))

    # Likelihood (1-5): Driven by violation frequency / historical occurrence
    if p_sent == 'Negative':
        likelihood = int(np.random.choice([2, 3, 4, 5], p=[0.15, 0.35, 0.35, 0.15]))
    elif p_sent == 'Positive':
        likelihood = int(np.random.choice([1, 2, 3], p=[0.50, 0.40, 0.10]))
    else:  # Neutral
        likelihood = int(np.random.choice([1, 2, 3, 4], p=[0.25, 0.40, 0.25, 0.10]))

    rows.append({
        'Teks Input (Kutipan dari Laporan Audit)': text_with_num,
        'TARGET: Kategori Risiko': cat,
        'TARGET: Sentimen': p_sent,
        'TARGET: Impact (1-5)': impact,
        'TARGET: Likelihood (1-5)': likelihood
    })

df_doc = pd.DataFrame(rows)
df_doc.to_csv(os.path.join(CURRENT_DIR, 'document_data.csv'), index=False)

print(f"[SUCCESS] Generated enriched document_data.csv with {len(df_doc)} rows!")
print("\n--- Target Sentiment Value Counts ---")
print(df_doc['TARGET: Sentimen'].value_counts())

print("\n--- Target Impact (1-5) Value Counts ---")
print(df_doc['TARGET: Impact (1-5)'].value_counts().sort_index())

print("\n--- Target Likelihood (1-5) Value Counts ---")
print(df_doc['TARGET: Likelihood (1-5)'].value_counts().sort_index())

print("\n--- Impact vs Likelihood Equality ---")
eq = (df_doc['TARGET: Impact (1-5)'] == df_doc['TARGET: Likelihood (1-5)']).mean()
print(f"Equality: {eq*100:.2f}%")
