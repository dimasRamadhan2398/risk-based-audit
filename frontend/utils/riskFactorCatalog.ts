// Seeded standard risk factors (backend/risk-service/cmd/seed_risk_universe.go), keyed by the
// numeric suffix of their seeded UUID. `en` mirrors the seed so we only translate rows the user
// has not edited; `g` lists guideline texts for scores 5 down to 1.
export interface RiskFactorCatalogText {
  name: string
  description: string
  g: string[]
}

export const RISK_FACTOR_CATALOG: Record<string, { en: RiskFactorCatalogText; id: RiskFactorCatalogText }> = {
  "1": {
    "en": {
      "name": "Financial Materiality",
      "description": "Size of financial exposure",
      "g": [
        "High – High revenue, expenditure, assets, or transaction values; major contributor to enterprise financial results",
        "Medium to High – Financial value affecting several business units or major operations",
        "Medium - Financial exposure with limited enterprise-wide impact",
        "Low to Medium - Value operations with limited financial significance",
        "Low - Financial activities or administrative support functions"
      ]
    },
    "id": {
      "name": "Materialitas Keuangan",
      "description": "Besarnya eksposur keuangan",
      "g": [
        "Tinggi – Pendapatan, belanja, aset, atau nilai transaksi besar; kontributor utama hasil keuangan perusahaan",
        "Sedang ke Tinggi – Nilai keuangan yang memengaruhi beberapa unit bisnis atau operasi utama",
        "Sedang - Eksposur keuangan dengan dampak terbatas pada tingkat perusahaan",
        "Rendah ke Sedang - Operasi bernilai dengan signifikansi keuangan yang terbatas",
        "Rendah - Aktivitas keuangan atau fungsi pendukung administratif"
      ]
    }
  },
  "2": {
    "en": {
      "name": "Strategic Importance",
      "description": "Contribution to organizational goals and objectives",
      "g": [
        "High - Directly supports enterprise strategy, Board priorities, or mission-critical objectives",
        "Medium to High - Supports major strategic programs or key business capabilities",
        "Medium - Contributes indirectly to strategic objectives",
        "Low to Medium - Primarily supports internal operations with limited strategic influence",
        "Low - Administrative or routine activities with minimal strategic impact"
      ]
    },
    "id": {
      "name": "Kepentingan Strategis",
      "description": "Kontribusi terhadap tujuan dan sasaran organisasi",
      "g": [
        "Tinggi - Mendukung langsung strategi perusahaan, prioritas Dewan, atau sasaran yang kritis bagi misi",
        "Sedang ke Tinggi - Mendukung program strategis utama atau kapabilitas bisnis kunci",
        "Sedang - Berkontribusi secara tidak langsung terhadap sasaran strategis",
        "Rendah ke Sedang - Terutama mendukung operasi internal dengan pengaruh strategis terbatas",
        "Rendah - Aktivitas administratif atau rutin dengan dampak strategis minimal"
      ]
    }
  },
  "3": {
    "en": {
      "name": "Inherent Risk",
      "description": "Natural level of risk before controls",
      "g": [
        "High – High value transactions, complex operations, significant fraud potential, heavy regulation, critical technology dependence",
        "Medium to High - Complexity with several inherent risk drivers",
        "Medium - Operational and financial exposure",
        "Low to Medium - Complexity and financial impact",
        "Low - Routine, repetitive, and standardized activities with minimal exposure"
      ]
    },
    "id": {
      "name": "Risiko Inheren",
      "description": "Tingkat risiko alami sebelum adanya pengendalian",
      "g": [
        "Tinggi – Transaksi bernilai tinggi, operasi kompleks, potensi fraud signifikan, regulasi ketat, ketergantungan teknologi yang kritis",
        "Sedang ke Tinggi - Kompleksitas dengan beberapa pemicu risiko inheren",
        "Sedang - Eksposur operasional dan keuangan",
        "Rendah ke Sedang - Kompleksitas dan dampak keuangan",
        "Rendah - Aktivitas rutin, berulang, dan terstandar dengan eksposur minimal"
      ]
    }
  },
  "4": {
    "en": {
      "name": "Internal Control Effectiveness",
      "description": "Effectiveness of existing controls",
      "g": [
        "High - Controls are ineffective or absent; frequent incidents; risk remains well above tolerance.",
        "Medium to High - Controls only partially effective; several unresolved issues; elevated exposure.",
        "Medium - Controls generally effective with some improvement opportunities.",
        "Low to Medium - Strong controls with only minor residual exposure.",
        "Low - Highly effective controls and continuous monitoring; residual risk is minimal."
      ]
    },
    "id": {
      "name": "Efektivitas Pengendalian Internal",
      "description": "Efektivitas pengendalian yang ada",
      "g": [
        "Tinggi - Pengendalian tidak efektif atau tidak ada; insiden sering terjadi; risiko tetap jauh di atas toleransi.",
        "Sedang ke Tinggi - Pengendalian hanya sebagian efektif; beberapa isu belum terselesaikan; eksposur meningkat.",
        "Sedang - Pengendalian umumnya efektif dengan beberapa peluang perbaikan.",
        "Rendah ke Sedang - Pengendalian kuat dengan eksposur residual yang kecil.",
        "Rendah - Pengendalian sangat efektif dan dipantau berkelanjutan; risiko residual minimal."
      ]
    }
  },
  "5": {
    "en": {
      "name": "Regulatory & Compliance Impact",
      "description": "Exposure to regulatory violations",
      "g": [
        "High - Highly regulated activity, frequent regulatory examinations, severe penalties for non-compliance",
        "Medium to High - Significant compliance obligations with regular reporting requirements",
        "Medium - Moderate regulatory requirements and periodic oversight",
        "Low to Medium - Limited compliance obligations with low enforcement risk",
        "Low - Minimal regulatory exposure beyond standard business requirements"
      ]
    },
    "id": {
      "name": "Dampak Regulasi & Kepatuhan",
      "description": "Eksposur terhadap pelanggaran regulasi",
      "g": [
        "Tinggi - Aktivitas sangat diregulasi, pemeriksaan regulator sering, sanksi berat atas ketidakpatuhan",
        "Sedang ke Tinggi - Kewajiban kepatuhan signifikan dengan kewajiban pelaporan rutin",
        "Sedang - Persyaratan regulasi moderat dan pengawasan berkala",
        "Rendah ke Sedang - Kewajiban kepatuhan terbatas dengan risiko penegakan rendah",
        "Rendah - Eksposur regulasi minimal di luar persyaratan bisnis standar"
      ]
    }
  },
  "6": {
    "en": {
      "name": "Operational Complexity",
      "description": "Difficulty of business processes",
      "g": [
        "High - Enterprise-wide operations, multiple systems, global locations, numerous stakeholders, highly integrated processes.",
        "Medium to High - Multi-department processes, significant technology integration, moderate geographic spread.",
        "Medium - Several interconnected processes within one business unit.",
        "Low to Medium - Limited process complexity and organizational dependencies.",
        "Low - Routine administrative activities with standardized procedures."
      ]
    },
    "id": {
      "name": "Kompleksitas Operasional",
      "description": "Tingkat kesulitan proses bisnis",
      "g": [
        "Tinggi - Operasi lintas perusahaan, banyak sistem, lokasi global, banyak pemangku kepentingan, proses sangat terintegrasi.",
        "Sedang ke Tinggi - Proses lintas departemen, integrasi teknologi signifikan, sebaran geografis moderat.",
        "Sedang - Beberapa proses yang saling terhubung dalam satu unit bisnis.",
        "Rendah ke Sedang - Kompleksitas proses dan ketergantungan organisasi terbatas.",
        "Rendah - Aktivitas administratif rutin dengan prosedur terstandar."
      ]
    }
  },
  "7": {
    "en": {
      "name": "Fraud Exposure",
      "description": "Opportunity for fraud",
      "g": [
        "High - High-value transactions, extensive management judgment, frequent manual overrides, prior fraud incidents, weak control environment, or elevated corruption risk.",
        "Medium to High - Significant cash/assets, complex transactions, extensive third-party involvement, or previous control weaknesses.",
        "Medium - Some exposure due to manual processes, moderate transaction volume, or limited segregation of duties.",
        "Low to Medium - Limited fraud opportunities with effective preventive and detective controls.",
        "Low - Strong controls, minimal cash handling, limited opportunity for fraud, no history of fraud."
      ]
    },
    "id": {
      "name": "Eksposur Fraud",
      "description": "Peluang terjadinya fraud",
      "g": [
        "Tinggi - Transaksi bernilai tinggi, pertimbangan manajemen yang luas, override manual yang sering, insiden fraud sebelumnya, lingkungan pengendalian lemah, atau risiko korupsi tinggi.",
        "Sedang ke Tinggi - Kas/aset signifikan, transaksi kompleks, keterlibatan pihak ketiga yang luas, atau kelemahan pengendalian sebelumnya.",
        "Sedang - Terdapat eksposur akibat proses manual, volume transaksi moderat, atau pemisahan tugas yang terbatas.",
        "Rendah ke Sedang - Peluang fraud terbatas dengan pengendalian preventif dan detektif yang efektif.",
        "Rendah - Pengendalian kuat, penanganan kas minimal, peluang fraud terbatas, tidak ada riwayat fraud."
      ]
    }
  },
  "8": {
    "en": {
      "name": "Information Security",
      "description": "Cyber security threat",
      "g": [
        "High - Mission-critical systems, highly sensitive or regulated data, frequent cyber threats, complex IT environment, or previous cybersecurity incidents.",
        "Medium to High - Critical business systems, significant confidential data, internet-facing applications, or known security weaknesses.",
        "Medium - Moderate reliance on IT systems, some sensitive information, standard security controls in place.",
        "Low to Medium - Limited exposure with effective cybersecurity governance and monitoring.",
        "Low - Minimal use of sensitive information, strong security controls, limited external connectivity."
      ]
    },
    "id": {
      "name": "Keamanan Informasi",
      "description": "Ancaman keamanan siber",
      "g": [
        "Tinggi - Sistem kritis bagi misi, data sangat sensitif atau diatur regulasi, ancaman siber sering, lingkungan TI kompleks, atau insiden keamanan siber sebelumnya.",
        "Sedang ke Tinggi - Sistem bisnis kritis, data rahasia dalam jumlah signifikan, aplikasi yang terhubung ke internet, atau kelemahan keamanan yang diketahui.",
        "Sedang - Ketergantungan moderat pada sistem TI, beberapa informasi sensitif, pengendalian keamanan standar tersedia.",
        "Rendah ke Sedang - Eksposur terbatas dengan tata kelola dan pemantauan keamanan siber yang efektif.",
        "Rendah - Penggunaan informasi sensitif minimal, pengendalian keamanan kuat, konektivitas eksternal terbatas."
      ]
    }
  },
  "9": {
    "en": {
      "name": "Business Continuity",
      "description": "Operational disruption potential",
      "g": [
        "High - Mission-critical process essential for organizational survival; disruption would cause severe financial, operational, legal, or reputational consequences.",
        "Medium to High - Critical process with strict recovery requirements; disruption significantly affects operations, customers, or compliance.",
        "Medium - Important process requiring documented continuity procedures; moderate financial or operational impact.",
        "Low to Medium - Limited operational impact, alternate procedures available, low customer impact.",
        "Low - Non-critical process, minimal business impact if interrupted, recovery within one business day without formal plans."
      ]
    },
    "id": {
      "name": "Keberlangsungan Bisnis",
      "description": "Potensi gangguan operasional",
      "g": [
        "Tinggi - Proses kritis bagi misi yang esensial bagi kelangsungan organisasi; gangguan akan menimbulkan konsekuensi keuangan, operasional, hukum, atau reputasi yang parah.",
        "Sedang ke Tinggi - Proses kritis dengan persyaratan pemulihan ketat; gangguan memengaruhi operasi, nasabah, atau kepatuhan secara signifikan.",
        "Sedang - Proses penting yang membutuhkan prosedur keberlangsungan terdokumentasi; dampak keuangan atau operasional moderat.",
        "Rendah ke Sedang - Dampak operasional terbatas, tersedia prosedur alternatif, dampak ke nasabah rendah.",
        "Rendah - Proses non-kritis, dampak bisnis minimal jika terhenti, pemulihan dalam satu hari kerja tanpa rencana formal."
      ]
    }
  },
  "10": {
    "en": {
      "name": "Change Management",
      "description": "Organizational or process changes",
      "g": [
        "High - Enterprise-wide transformation, multiple concurrent critical initiatives, major system implementation, or merger with high execution uncertainty and substantial operational impact.",
        "Medium to High - Significant organizational or technology changes affecting multiple functions; increased implementation risk.",
        "Medium - Several moderate initiatives with manageable implementation risk and adequate governance.",
        "Low to Medium - Limited changes affecting a small portion of the organization; changes are well managed.",
        "Low - Stable environment with minimal organizational or technology changes and mature change governance."
      ]
    },
    "id": {
      "name": "Manajemen Perubahan",
      "description": "Perubahan organisasi atau proses",
      "g": [
        "Tinggi - Transformasi lintas perusahaan, beberapa inisiatif kritis berjalan bersamaan, implementasi sistem besar, atau merger dengan ketidakpastian eksekusi tinggi dan dampak operasional substansial.",
        "Sedang ke Tinggi - Perubahan organisasi atau teknologi signifikan yang memengaruhi beberapa fungsi; risiko implementasi meningkat.",
        "Sedang - Beberapa inisiatif moderat dengan risiko implementasi yang terkelola dan tata kelola memadai.",
        "Rendah ke Sedang - Perubahan terbatas yang memengaruhi sebagian kecil organisasi; perubahan dikelola dengan baik.",
        "Rendah - Lingkungan stabil dengan perubahan organisasi atau teknologi minimal dan tata kelola perubahan yang matang."
      ]
    }
  },
  "11": {
    "en": {
      "name": "Third-Party Dependency",
      "description": "Dependency on external providers",
      "g": [
        "High - Mission-critical operations depend on one or more external providers with limited alternatives; failure of these providers could severely disrupt operations, regulatory compliance, customer service, or financial performance.",
        "Medium to High - Significant reliance on external providers for critical operations, technology, or supply chain activities; disruption would materially affect business performance.",
        "Medium - Moderate dependence on external providers supporting important, but not mission-critical, activities.",
        "Low to Medium - Limited use of third parties with effective oversight and readily available alternatives.",
        "Low - Minimal reliance on external parties; services are non-critical and easily replaceable."
      ]
    },
    "id": {
      "name": "Ketergantungan Pihak Ketiga",
      "description": "Ketergantungan pada penyedia eksternal",
      "g": [
        "Tinggi - Operasi kritis bagi misi bergantung pada satu atau lebih penyedia eksternal dengan alternatif terbatas; kegagalan penyedia dapat mengganggu operasi, kepatuhan regulasi, layanan nasabah, atau kinerja keuangan secara parah.",
        "Sedang ke Tinggi - Ketergantungan signifikan pada penyedia eksternal untuk operasi, teknologi, atau rantai pasok yang kritis; gangguan akan memengaruhi kinerja bisnis secara material.",
        "Sedang - Ketergantungan moderat pada penyedia eksternal yang mendukung aktivitas penting, namun tidak kritis bagi misi.",
        "Rendah ke Sedang - Penggunaan pihak ketiga terbatas dengan pengawasan efektif dan alternatif yang mudah tersedia.",
        "Rendah - Ketergantungan minimal pada pihak eksternal; layanan bersifat non-kritis dan mudah digantikan."
      ]
    }
  },
  "12": {
    "en": {
      "name": "Reputational Impact",
      "description": "Potential damage to reputation",
      "g": [
        "High - Severe and long-lasting reputational damage with major public attention, stakeholder distrust, litigation, or executive accountability.",
        "Medium to High - Significant damage to customer trust, investor confidence, or regulatory relationships. Likely national media coverage.",
        "Medium - Potential negative perception among customers, regulators, or business partners. Moderate media attention possible.",
        "Low to Medium - Limited impact affecting only a small group of stakeholders. Easily managed internally.",
        "Low - Minimal or no impact on organizational reputation. Issues are unlikely to become publicly known."
      ]
    },
    "id": {
      "name": "Dampak Reputasi",
      "description": "Potensi kerusakan reputasi",
      "g": [
        "Tinggi - Kerusakan reputasi parah dan berkepanjangan dengan perhatian publik besar, ketidakpercayaan pemangku kepentingan, litigasi, atau pertanggungjawaban eksekutif.",
        "Sedang ke Tinggi - Kerusakan signifikan pada kepercayaan nasabah, keyakinan investor, atau hubungan dengan regulator. Kemungkinan diliput media nasional.",
        "Sedang - Potensi persepsi negatif di kalangan nasabah, regulator, atau mitra bisnis. Perhatian media moderat dimungkinkan.",
        "Rendah ke Sedang - Dampak terbatas hanya pada sekelompok kecil pemangku kepentingan. Mudah dikelola secara internal.",
        "Rendah - Dampak minimal atau tidak ada pada reputasi organisasi. Isu kecil kemungkinan menjadi pengetahuan publik."
      ]
    }
  },
  "13": {
    "en": {
      "name": "Customer Impact",
      "description": "Risk of damage to customer engagement",
      "g": [
        "High - Severe impact on a large customer base, widespread service failure, customer harm, substantial complaints, or long-term loss of customer trust.",
        "Medium to High - Significant customer disruption, financial impact, or major decline in customer satisfaction affecting many customers.",
        "Medium - Noticeable customer dissatisfaction or moderate service disruption affecting a defined customer group.",
        "Low to Medium - Limited customer effect involving a small number of customers. Minor inconvenience.",
        "Low - No direct customer interaction. Failures have little or no customer impact."
      ]
    },
    "id": {
      "name": "Dampak terhadap Nasabah",
      "description": "Risiko kerusakan hubungan dengan nasabah",
      "g": [
        "Tinggi - Dampak parah pada basis nasabah yang besar, kegagalan layanan meluas, kerugian nasabah, keluhan substansial, atau hilangnya kepercayaan nasabah jangka panjang.",
        "Sedang ke Tinggi - Gangguan nasabah signifikan, dampak keuangan, atau penurunan besar kepuasan nasabah yang memengaruhi banyak nasabah.",
        "Sedang - Ketidakpuasan nasabah yang terasa atau gangguan layanan moderat yang memengaruhi kelompok nasabah tertentu.",
        "Rendah ke Sedang - Efek terbatas pada sejumlah kecil nasabah. Ketidaknyamanan ringan.",
        "Rendah - Tidak ada interaksi langsung dengan nasabah. Kegagalan berdampak kecil atau tidak ada pada nasabah."
      ]
    }
  },
  "14": {
    "en": {
      "name": "Previous Audit Findings",
      "description": "Prior audit issues",
      "g": [
        "High - Numerous High or Critical findings, unresolved recommendations, repeated deficiencies, or ineffective management response.",
        "Medium to High - Minor findings identified. Corrective actions completed on time. Strong management response.",
        "Medium - Several moderate findings or isolated overdue actions. Some improvement opportunities remain.",
        "Low to Medium - Minor findings identified. Corrective actions completed on time. Strong management response.",
        "Low - No significant findings in previous audits. Recommendations fully implemented with no repeat issues."
      ]
    },
    "id": {
      "name": "Temuan Audit Sebelumnya",
      "description": "Isu audit sebelumnya",
      "g": [
        "Tinggi - Banyak temuan Tinggi atau Kritis, rekomendasi belum terselesaikan, kelemahan berulang, atau tanggapan manajemen tidak efektif.",
        "Sedang ke Tinggi - Temuan minor teridentifikasi. Tindakan korektif diselesaikan tepat waktu. Tanggapan manajemen kuat.",
        "Sedang - Beberapa temuan moderat atau tindak lanjut yang terlambat secara terisolasi. Masih ada peluang perbaikan.",
        "Rendah ke Sedang - Temuan minor teridentifikasi. Tindakan korektif diselesaikan tepat waktu. Tanggapan manajemen kuat.",
        "Rendah - Tidak ada temuan signifikan pada audit sebelumnya. Rekomendasi diimplementasikan sepenuhnya tanpa isu berulang."
      ]
    }
  },
  "15": {
    "en": {
      "name": "Time Since Last Audit",
      "description": "Length of time since the last review",
      "g": [
        "High – More than 5 years or never audited.",
        "Medium to High – 3-5 years.",
        "Medium – 2-3 years.",
        "Low to Medium – 1-2 years.",
        "Low - Less than 1 year."
      ]
    },
    "id": {
      "name": "Waktu Sejak Audit Terakhir",
      "description": "Lamanya waktu sejak review terakhir",
      "g": [
        "Tinggi – Lebih dari 5 tahun atau belum pernah diaudit.",
        "Sedang ke Tinggi – 3-5 tahun.",
        "Sedang – 2-3 tahun.",
        "Rendah ke Sedang – 1-2 tahun.",
        "Rendah - Kurang dari 1 tahun."
      ]
    }
  },
  "16": {
    "en": {
      "name": "Management Concern",
      "description": "Management's perception of risk",
      "g": [
        "High – Critical management concern requiring immediate independent assurance due to significant uncertainty, major incidents, or substantial business risk.",
        "Medium to High – Significant management concern regarding governance, compliance, operational performance, or strategic execution. Management requests priority audit coverage.",
        "Medium – Moderate concerns regarding process efficiency, control effectiveness, or emerging risks. Management requests periodic assurance.",
        "Low to Medium – Minor concerns expressed by management, with no significant operational or control issues.",
        "Low - No significant management concerns. Controls are considered effective and the area is operating as expected."
      ]
    },
    "id": {
      "name": "Perhatian Manajemen",
      "description": "Persepsi manajemen terhadap risiko",
      "g": [
        "Tinggi – Perhatian manajemen yang kritis dan memerlukan asurans independen segera karena ketidakpastian signifikan, insiden besar, atau risiko bisnis substansial.",
        "Sedang ke Tinggi – Perhatian manajemen yang signifikan terkait tata kelola, kepatuhan, kinerja operasional, atau eksekusi strategi. Manajemen meminta cakupan audit prioritas.",
        "Sedang – Perhatian moderat terkait efisiensi proses, efektivitas pengendalian, atau risiko yang muncul. Manajemen meminta asurans berkala.",
        "Rendah ke Sedang – Perhatian minor disampaikan manajemen, tanpa isu operasional atau pengendalian yang signifikan.",
        "Rendah - Tidak ada perhatian manajemen yang signifikan. Pengendalian dinilai efektif dan area beroperasi sesuai harapan."
      ]
    }
  },
  "17": {
    "en": {
      "name": "HSE",
      "description": "Frequency of incidents (Health, Safety, and Environment)",
      "g": [
        "High – Frequent or critical incidents causing major operational disruption, financial loss, regulatory action, safety events, cybersecurity breaches, fraud, or significant reputational damage.",
        "Medium to High – Multiple significant incidents affecting operations, customers, compliance, or financial performance. Repeat incidents indicate control weaknesses.",
        "Medium – Several moderate incidents or an increasing incident trend. Some recurring issues identified.",
        "Low to Medium – Few minor incidents with limited impact. Corrective actions were timely and effective.",
        "Low - No significant incidents during the assessment period. Strong incident prevention and response processes."
      ]
    },
    "id": {
      "name": "K3L (HSE)",
      "description": "Frekuensi insiden (Kesehatan, Keselamatan, dan Lingkungan)",
      "g": [
        "Tinggi – Insiden sering atau kritis yang menyebabkan gangguan operasional besar, kerugian keuangan, tindakan regulator, kejadian keselamatan, pelanggaran keamanan siber, fraud, atau kerusakan reputasi signifikan.",
        "Sedang ke Tinggi – Beberapa insiden signifikan yang memengaruhi operasi, nasabah, kepatuhan, atau kinerja keuangan. Insiden berulang menunjukkan kelemahan pengendalian.",
        "Sedang – Beberapa insiden moderat atau tren insiden yang meningkat. Teridentifikasi beberapa isu yang berulang.",
        "Rendah ke Sedang – Sedikit insiden minor dengan dampak terbatas. Tindakan korektif tepat waktu dan efektif.",
        "Rendah - Tidak ada insiden signifikan selama periode penilaian. Proses pencegahan dan respons insiden kuat."
      ]
    }
  },
  "18": {
    "en": {
      "name": "Process Automation Level",
      "description": "Complexity of process automation",
      "g": [
        "High – Predominantly manual processes with weak controls, or highly complex automation lacking effective governance, monitoring, cybersecurity, or change management.",
        "Medium to High – Significant manual processing or highly complex automation with governance weaknesses, increasing operational or technology risk.",
        "Medium – Mixed manual and automated processing. Some reliance on manual controls or moderate system complexity.",
        "Low to Medium – Mostly automated with effective controls and only minor manual activities.",
        "Low - Well-controlled automation with strong governance, effective IT controls, reliable monitoring, and minimal manual intervention."
      ]
    },
    "id": {
      "name": "Tingkat Otomasi Proses",
      "description": "Kompleksitas otomasi proses",
      "g": [
        "Tinggi – Proses didominasi manual dengan pengendalian lemah, atau otomasi sangat kompleks tanpa tata kelola, pemantauan, keamanan siber, atau manajemen perubahan yang efektif.",
        "Sedang ke Tinggi – Pemrosesan manual yang signifikan atau otomasi sangat kompleks dengan kelemahan tata kelola, meningkatkan risiko operasional atau teknologi.",
        "Sedang – Pemrosesan campuran manual dan otomatis. Ada ketergantungan pada pengendalian manual atau kompleksitas sistem moderat.",
        "Rendah ke Sedang – Sebagian besar otomatis dengan pengendalian efektif dan hanya sedikit aktivitas manual.",
        "Rendah - Otomasi terkendali dengan tata kelola kuat, pengendalian TI efektif, pemantauan andal, dan intervensi manual minimal."
      ]
    }
  },
  "19": {
    "en": {
      "name": "Competency of Employee",
      "description": "Competency fit of employees",
      "g": [
        "High – Critical competency gaps, lack of qualified personnel, ineffective succession planning, or widespread skill deficiencies causing significant operational and control risks.",
        "Medium to High – Significant competency deficiencies, skill shortages, high turnover, or inadequate training affecting operational performance.",
        "Medium – Some competency gaps exist or moderate dependence on key personnel. Additional training is required.",
        "Low to Medium – Competency requirements are generally met with only minor development needs.",
        "Low - Highly competent workforce with appropriate qualifications, certifications, continuous training, and low turnover."
      ]
    },
    "id": {
      "name": "Kompetensi Pegawai",
      "description": "Kesesuaian kompetensi pegawai",
      "g": [
        "Tinggi – Kesenjangan kompetensi kritis, kurangnya personel berkualifikasi, perencanaan suksesi tidak efektif, atau defisiensi keterampilan meluas yang menyebabkan risiko operasional dan pengendalian signifikan.",
        "Sedang ke Tinggi – Defisiensi kompetensi signifikan, kekurangan keterampilan, perputaran pegawai tinggi, atau pelatihan tidak memadai yang memengaruhi kinerja operasional.",
        "Sedang – Terdapat beberapa kesenjangan kompetensi atau ketergantungan moderat pada personel kunci. Diperlukan pelatihan tambahan.",
        "Rendah ke Sedang – Persyaratan kompetensi umumnya terpenuhi dengan kebutuhan pengembangan yang minor.",
        "Rendah - Tenaga kerja sangat kompeten dengan kualifikasi, sertifikasi, pelatihan berkelanjutan yang sesuai, dan perputaran rendah."
      ]
    }
  },
  "20": {
    "en": {
      "name": "Data Sensitivity",
      "description": "Level of data sensitivity",
      "g": [
        "High – Handles highly confidential, mission-critical, or regulated information where unauthorized disclosure, alteration, or loss would have severe legal, financial, operational, or reputational consequences.",
        "Medium to High – Processes significant volumes of sensitive customer, financial, employee, or proprietary information subject to regulatory requirements.",
        "Medium – Handles confidential business information or moderate volumes of personal data requiring standard protection.",
        "Low to Medium – Processes limited internal-use information with relatively low confidentiality requirements.",
        "Low - Handles only public or non-confidential information with minimal business impact if disclosed."
      ]
    },
    "id": {
      "name": "Sensitivitas Data",
      "description": "Tingkat sensitivitas data",
      "g": [
        "Tinggi – Menangani informasi sangat rahasia, kritis bagi misi, atau diatur regulasi yang jika diungkap, diubah, atau hilang tanpa izin akan berdampak hukum, keuangan, operasional, atau reputasi yang parah.",
        "Sedang ke Tinggi – Memproses informasi nasabah, keuangan, pegawai, atau milik perusahaan yang sensitif dalam volume besar dan tunduk pada persyaratan regulasi.",
        "Sedang – Menangani informasi bisnis rahasia atau data pribadi dalam volume moderat yang memerlukan perlindungan standar.",
        "Rendah ke Sedang – Memproses informasi untuk penggunaan internal yang terbatas dengan kebutuhan kerahasiaan relatif rendah.",
        "Rendah - Hanya menangani informasi publik atau tidak rahasia dengan dampak bisnis minimal jika terungkap."
      ]
    }
  },
  "21": {
    "en": {
      "name": "Project Criticality",
      "description": "The importance and potential impact of a project to the organization",
      "g": [
        "High – Mission-critical enterprise project; failure would significantly affect strategy, operations, or financial performance.",
        "Medium to High – High-impact project involving multiple business units with substantial financial or operational implications.",
        "Medium – Moderate-impact project affecting one or more departments with manageable risks.",
        "Low to Medium – Limited-scope project with minimal organizational impact.",
        "Low - Routine or minor project with negligible impact on business objectives."
      ]
    },
    "id": {
      "name": "Kritikalitas Proyek",
      "description": "Pentingnya dan potensi dampak proyek bagi organisasi",
      "g": [
        "Tinggi – Proyek perusahaan yang kritis bagi misi; kegagalan akan memengaruhi strategi, operasi, atau kinerja keuangan secara signifikan.",
        "Sedang ke Tinggi – Proyek berdampak tinggi yang melibatkan beberapa unit bisnis dengan implikasi keuangan atau operasional substansial.",
        "Sedang – Proyek berdampak moderat yang memengaruhi satu atau lebih departemen dengan risiko yang terkelola.",
        "Rendah ke Sedang – Proyek berlingkup terbatas dengan dampak organisasi minimal.",
        "Rendah - Proyek rutin atau kecil dengan dampak yang dapat diabaikan terhadap sasaran bisnis."
      ]
    }
  },
  "22": {
    "en": {
      "name": "Legal Exposure",
      "description": "Exposure to legal violations",
      "g": [
        "High – Extensive legal exposure with major litigation, high-value contractual obligations, or activities where legal failure could materially impact the organization.",
        "Medium to High – Significant legal exposure due to complex contracts, international operations, or recurring legal matters.",
        "Medium – Moderate legal exposure involving important contracts or occasional legal disputes.",
        "Low to Medium – Routine legal obligations with limited contractual or litigation exposure",
        "Low - Minimal legal obligations and very limited exposure to litigation or contractual disputes."
      ]
    },
    "id": {
      "name": "Eksposur Hukum",
      "description": "Eksposur terhadap pelanggaran hukum",
      "g": [
        "Tinggi – Eksposur hukum yang luas dengan litigasi besar, kewajiban kontraktual bernilai tinggi, atau aktivitas yang kegagalan hukumnya dapat berdampak material bagi organisasi.",
        "Sedang ke Tinggi – Eksposur hukum signifikan akibat kontrak kompleks, operasi internasional, atau perkara hukum berulang.",
        "Sedang – Eksposur hukum moderat yang melibatkan kontrak penting atau sengketa hukum sesekali.",
        "Rendah ke Sedang – Kewajiban hukum rutin dengan eksposur kontraktual atau litigasi terbatas",
        "Rendah - Kewajiban hukum minimal dan eksposur yang sangat terbatas terhadap litigasi atau sengketa kontrak."
      ]
    }
  },
  "23": {
    "en": {
      "name": "ESG Impact",
      "description": "Impact to the environment, social, and governance",
      "g": [
        "High – Critical ESG impact where failures could materially affect sustainability performance, stakeholder trust, regulatory compliance, or long-term organizational value.",
        "Medium to High – Significant impact on ESG performance with substantial regulatory, operational, or reputational implications.",
        "Medium – Moderate influence on sustainability objectives or stakeholder expectations.",
        "Low to Medium – Limited ESG impact with relatively minor stakeholder implications.",
        "Low - Minimal influence on environmental, social, or governance outcomes."
      ]
    },
    "id": {
      "name": "Dampak ESG",
      "description": "Dampak terhadap lingkungan, sosial, dan tata kelola",
      "g": [
        "Tinggi – Dampak ESG kritis yang kegagalannya dapat memengaruhi kinerja keberlanjutan, kepercayaan pemangku kepentingan, kepatuhan regulasi, atau nilai organisasi jangka panjang secara material.",
        "Sedang ke Tinggi – Dampak signifikan pada kinerja ESG dengan implikasi regulasi, operasional, atau reputasi yang substansial.",
        "Sedang – Pengaruh moderat terhadap sasaran keberlanjutan atau ekspektasi pemangku kepentingan.",
        "Rendah ke Sedang – Dampak ESG terbatas dengan implikasi yang relatif kecil bagi pemangku kepentingan.",
        "Rendah - Pengaruh minimal terhadap hasil lingkungan, sosial, atau tata kelola."
      ]
    }
  },
  "24": {
    "en": {
      "name": "Emerging Risk",
      "description": "The extent to which an auditable entity is exposed to new, evolving, or rapidly changing risks",
      "g": [
        "High – Significant exposure to rapidly evolving risks with potentially severe strategic, operational, financial, or regulatory consequences. Governance and controls are still developing.",
        "Medium to High – High exposure to emerging risks requiring enhanced monitoring and management attention.",
        "Medium – Moderate exposure with identifiable risks and partially established controls.",
        "Low to Medium – Limited exposure to emerging risks with manageable uncertainty.",
        "Low - Minimal exposure to new or evolving risks; operating environment is relatively stable."
      ]
    },
    "id": {
      "name": "Risiko yang Muncul (Emerging Risk)",
      "description": "Sejauh mana entitas audit terpapar risiko baru, berkembang, atau berubah dengan cepat",
      "g": [
        "Tinggi – Eksposur signifikan terhadap risiko yang berkembang pesat dengan konsekuensi strategis, operasional, keuangan, atau regulasi yang berpotensi parah. Tata kelola dan pengendalian masih berkembang.",
        "Sedang ke Tinggi – Eksposur tinggi terhadap risiko yang muncul sehingga memerlukan pemantauan dan perhatian manajemen yang lebih intensif.",
        "Sedang – Eksposur moderat dengan risiko yang dapat diidentifikasi dan pengendalian yang sebagian sudah terbentuk.",
        "Rendah ke Sedang – Eksposur terbatas terhadap risiko yang muncul dengan ketidakpastian yang dapat dikelola.",
        "Rendah - Eksposur minimal terhadap risiko baru atau berkembang; lingkungan operasi relatif stabil."
      ]
    }
  }
}
