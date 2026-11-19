package services

// ORIGINAL AuditSphere COSO 2013 point-of-focus catalogue (version 2013.1).
//
// The wording below was written for AuditSphere from the publicly known COSO
// 2013 points of focus, in English and Indonesian. It contains no text, no
// numbering scheme specifics, regulation numbers or company names taken from
// any third-party working paper. Numbering follows the common
// component.principle.pof.element style (for example 1.1.1.1).
//
// Weights are NOT written here: BuildCosoPofCatalogue derives them from the
// fixed component weights (25 / 25 / 20 / 17.5 / 12.5) and the principle
// weights below, splitting a principle over its points of focus and a point of
// focus over its elements as evenly as 0.001 allows, so the sums always match.
// A later catalogue version may hand-tune weights; only the invariants matter
// (ValidateCosoPofCatalogue).

const (
	pofMethodsDK = "D,K"
)

type elemDef struct {
	Type    string // PRESENT or FUNCTION
	EN, ID  string
	Methods string
	Indep   bool
}

type pofDef struct {
	NameEN, NameID string
	Elements       []elemDef
}

type principleDef struct {
	Code   string
	NameID string
	Weight int64 // thousandths of a point
	Owners string
	Pofs   []pofDef
}

type componentDef struct {
	Code       string
	NameID     string
	Weight     int64 // thousandths of a point
	Principles []principleDef
}

func pr(en, id, methods string) elemDef {
	return elemDef{Type: "PRESENT", EN: en, ID: id, Methods: methods}
}
func fn(en, id, methods string) elemDef {
	return elemDef{Type: "FUNCTION", EN: en, ID: id, Methods: methods}
}
func fi(en, id, methods string) elemDef {
	return elemDef{Type: "FUNCTION", EN: en, ID: id, Methods: methods, Indep: true}
}
func pri(en, id, methods string) elemDef {
	return elemDef{Type: "PRESENT", EN: en, ID: id, Methods: methods, Indep: true}
}
func pf(en, id string, e ...elemDef) pofDef { return pofDef{NameEN: en, NameID: id, Elements: e} }

var cosoPofComponentDefs = []componentDef{
	{Code: "CE", NameID: "Lingkungan Pengendalian", Weight: 25000, Principles: []principleDef{
		{Code: "P01", NameID: "Menunjukkan komitmen pada integritas dan nilai etika", Weight: 5500, Owners: "EXECUTIVE", Pofs: []pofDef{
			pf("Tone at the top", "Teladan dari pimpinan",
				pr("Leadership has issued messages or directives on integrity and ethics.", "Pimpinan telah menerbitkan pesan atau arahan tentang integritas dan etika.", "D"),
				fn("Leaders' decisions and conduct are consistent with the stated values.", "Keputusan dan perilaku pimpinan konsisten dengan nilai yang dinyatakan.", "K,W"),
				fn("Staff perceive that leadership acts with integrity.", "Pegawai menilai pimpinan bertindak dengan integritas.", "K")),
			pf("Standards of conduct", "Standar perilaku",
				pr("A code of conduct covering ethics, conflicts of interest and gifts is approved.", "Kode etik yang mengatur etika, benturan kepentingan, dan gratifikasi telah disahkan.", "D"),
				fn("The code is communicated to all staff and acknowledged in writing on a regular basis.", "Kode etik disosialisasikan ke seluruh pegawai dan diakui tertulis secara berkala.", "D,K"),
				fn("Third parties such as vendors and contractors are informed of the expected standards.", "Pihak ketiga seperti vendor dan kontraktor diberi tahu standar yang diharapkan.", "D")),
			pf("Evaluating adherence", "Evaluasi kepatuhan",
				pr("A process exists to assess adherence to the standards of conduct (surveys, declarations, reviews).", "Terdapat proses menilai kepatuhan terhadap standar perilaku (survei, deklarasi, reviu).", "D"),
				fn("The assessment was carried out in the period and its results reported to leadership.", "Penilaian dilakukan pada periode berjalan dan hasilnya dilaporkan ke pimpinan.", "D,W")),
			pf("Addressing deviations", "Penanganan penyimpangan",
				pr("A channel for reporting violations that protects reporters is available.", "Tersedia kanal pelaporan pelanggaran yang melindungi pelapor.", "D"),
				fn("Reported deviations are investigated and sanctioned in a timely and consistent way.", "Penyimpangan yang dilaporkan diselidiki dan diberi sanksi tepat waktu dan konsisten.", "D,W"),
				fn("Outcomes are tracked to closure and analysed for recurring themes.", "Hasil penanganan dipantau hingga selesai dan dianalisis pola berulangnya.", "D")),
		}},
		{Code: "P02", NameID: "Dewan pengawas menjalankan tanggung jawab pengawasan", Weight: 5000, Owners: "AUDIT_COMMITTEE,CORPORATE_SECRETARY", Pofs: []pofDef{
			pf("Oversight responsibilities", "Tanggung jawab pengawasan",
				pr("A charter or terms of reference define the oversight duties of the board and audit committee.", "Piagam atau kerangka acuan mendefinisikan tugas pengawasan dewan dan komite audit.", "D"),
				fn("Oversight duties are carried out through regular meetings with minutes.", "Tugas pengawasan dijalankan melalui rapat berkala yang tercatat risalahnya.", "D,W")),
			pf("Relevant expertise", "Keahlian yang relevan",
				pr("The skills needed by overseers (finance, risk, technology, industry) are defined.", "Keahlian yang dibutuhkan pengawas (keuangan, risiko, teknologi, industri) ditetapkan.", "D"),
				fn("The composition of the oversight body covers these skills and gaps are addressed.", "Komposisi badan pengawas memenuhi keahlian tersebut dan kekurangannya diatasi.", "D,W")),
			pf("Independence", "Independensi",
				pr("Criteria for independent members and conflict-of-interest rules for overseers exist.", "Terdapat kriteria anggota independen dan aturan benturan kepentingan bagi pengawas.", "D"),
				fn("Independent members take part in oversight decisions and challenge management.", "Anggota independen berperan dalam keputusan pengawasan dan menantang manajemen.", "D,W")),
			pf("Oversight of the internal control system", "Pengawasan sistem pengendalian intern",
				pr("The agenda of the oversight body covers control environment, risk, control activities, information and monitoring.", "Agenda badan pengawas mencakup lingkungan pengendalian, risiko, aktivitas pengendalian, informasi, dan pemantauan.", "D"),
				fn("The body reviews reports of internal audit and other assurance providers and follows up.", "Badan pengawas mereviu laporan audit internal dan penyedia asurans lain dan menindaklanjutinya.", "D,W"),
				fn("The possibility that management overrides controls is considered.", "Kemungkinan manajemen mengabaikan pengendalian dipertimbangkan.", "W")),
		}},
		{Code: "P03", NameID: "Menetapkan struktur, wewenang, dan tanggung jawab", Weight: 4500, Owners: "EXECUTIVE", Pofs: []pofDef{
			pf("Organisational structures", "Struktur organisasi",
				pr("An approved organisation chart shows units, functions and outsourced services.", "Bagan organisasi yang disahkan menunjukkan unit, fungsi, dan layanan alih daya.", "D"),
				fn("Structures are reviewed when objectives or operations change.", "Struktur ditinjau ulang saat tujuan atau operasi berubah.", "D,W")),
			pf("Reporting lines", "Jalur pelaporan",
				pr("Reporting lines, including dotted lines, are documented for each unit.", "Jalur pelaporan, termasuk garis koordinasi, didokumentasikan untuk setiap unit.", "D"),
				fn("Reporting lines work in practice and allow escalation without obstruction.", "Jalur pelaporan berjalan dalam praktik dan memungkinkan eskalasi tanpa hambatan.", "W,K")),
			pf("Authority and responsibility", "Wewenang dan tanggung jawab",
				pr("A delegation of authority matrix with limits and approval thresholds is in place.", "Matriks delegasi wewenang beserta batas dan ambang persetujuan tersedia.", "D"),
				fn("Transactions are approved within the delegated limits.", "Transaksi disetujui dalam batas wewenang yang didelegasikan.", "D,O"),
				fn("Control responsibilities are in job descriptions and understood by staff.", "Tanggung jawab pengendalian tercantum dalam uraian jabatan dan dipahami pegawai.", "D,K")),
		}},
		{Code: "P04", NameID: "Menunjukkan komitmen pada kompetensi", Weight: 5000, Owners: "HEAD_OF_HR", Pofs: []pofDef{
			pf("Policies and practices", "Kebijakan dan praktik",
				pr("Policies set the competence requirements of each role.", "Kebijakan menetapkan persyaratan kompetensi tiap peran.", "D"),
				fn("Recruitment and promotion follow the competence requirements.", "Rekrutmen dan promosi mengikuti persyaratan kompetensi.", "D,W")),
			pf("Evaluating competence", "Evaluasi kompetensi",
				pr("A method to assess competence exists (appraisal, skills matrix).", "Metode penilaian kompetensi tersedia (penilaian kinerja, matriks keterampilan).", "D"),
				fn("Gaps found are addressed through training or other actions.", "Kesenjangan yang ditemukan ditangani melalui pelatihan atau tindakan lain.", "D,W")),
			pf("Attracting, developing and retaining people", "Menarik, mengembangkan, dan mempertahankan pegawai",
				pr("Training and development plans and a retention approach exist.", "Rencana pelatihan dan pengembangan serta pendekatan retensi tersedia.", "D"),
				fn("Training is delivered as planned and its effect is evaluated.", "Pelatihan dilaksanakan sesuai rencana dan efektivitasnya dievaluasi.", "D,K")),
			pf("Succession", "Suksesi",
				pr("Succession plans exist for key positions and for the oversight body.", "Rencana suksesi tersedia untuk posisi kunci dan badan pengawas.", "D"),
				fn("Successors are prepared and cover arrangements work when key people leave.", "Calon pengganti dipersiapkan dan pengganti sementara berjalan ketika personel kunci pergi.", "D,W")),
		}},
		{Code: "P05", NameID: "Menegakkan akuntabilitas", Weight: 5000, Owners: "EXECUTIVE", Pofs: []pofDef{
			pf("Accountability through structures", "Akuntabilitas melalui struktur",
				pr("Accountability for internal control is assigned to managers in writing.", "Akuntabilitas pengendalian intern ditetapkan tertulis kepada manajer.", "D"),
				fn("Managers report on their control responsibilities and are held to account.", "Manajer melaporkan tanggung jawab pengendalian dan dimintai pertanggungjawaban.", "D,W")),
			pf("Performance measures and incentives", "Ukuran kinerja dan insentif",
				pr("Performance measures and incentives are defined for each level.", "Ukuran kinerja dan insentif ditetapkan untuk tiap level.", "D"),
				fn("Measures cover conduct and control, not only financial results.", "Ukuran mencakup perilaku dan pengendalian, bukan hanya hasil keuangan.", "D,K")),
			pf("Ongoing relevance of measures", "Relevansi ukuran yang berkelanjutan",
				pr("A periodic review of the measures is scheduled.", "Reviu berkala atas ukuran kinerja dijadwalkan.", "D"),
				fn("Measures are adjusted when objectives or risks change.", "Ukuran disesuaikan ketika tujuan atau risiko berubah.", "D,W")),
			pf("Pressures", "Tekanan",
				pr("Workload and target-setting guidelines address excessive pressure.", "Pedoman beban kerja dan penetapan target mengatasi tekanan berlebihan.", "D"),
				fn("Pressure indicators (overtime, turnover, survey results) are monitored and acted on.", "Indikator tekanan (lembur, perputaran pegawai, hasil survei) dipantau dan ditindaklanjuti.", "D,K")),
			pf("Rewards and discipline", "Penghargaan dan sanksi",
				pr("Reward and disciplinary procedures exist.", "Prosedur penghargaan dan sanksi tersedia.", "D"),
				fn("Rewards and sanctions are applied consistently to the performance evaluated.", "Penghargaan dan sanksi diterapkan konsisten sesuai kinerja yang dievaluasi.", "D,W")),
		}},
	}},
	{Code: "RA", NameID: "Penilaian Risiko", Weight: 25000, Principles: []principleDef{
		{Code: "P06", NameID: "Menetapkan tujuan yang sesuai", Weight: 6000, Owners: "RISK_MANAGEMENT,DEPARTMENT_HEAD", Pofs: []pofDef{
			pf("Operations objectives reflect management choices", "Tujuan operasi mencerminkan pilihan manajemen",
				pr("Operational objectives are documented and linked to strategy.", "Tujuan operasional didokumentasikan dan terkait dengan strategi.", "D"),
				fn("Objectives are communicated to the staff responsible for them.", "Tujuan dikomunikasikan kepada pegawai yang bertanggung jawab.", "D,K")),
			pf("Risk tolerance for operations", "Toleransi risiko untuk tujuan operasi",
				pr("Acceptable variation (tolerance) is defined for key operational objectives.", "Variasi yang dapat diterima (toleransi) ditetapkan untuk tujuan operasi utama.", "D"),
				fn("Tolerances are used in monitoring and exceedances are escalated.", "Toleransi dipakai dalam pemantauan dan pelampauannya dieskalasi.", "D,W")),
			pf("Resources aligned with objectives", "Sumber daya selaras dengan tujuan",
				pr("Budget and resource plans refer to the objectives.", "Anggaran dan rencana sumber daya merujuk pada tujuan.", "D"),
				fn("Resource allocation is reviewed against the achievement of objectives.", "Alokasi sumber daya ditinjau terhadap pencapaian tujuan.", "D,W")),
			pf("Financial reporting objectives and standards", "Tujuan pelaporan keuangan dan standar",
				pr("Applicable accounting standards are identified and documented.", "Standar akuntansi yang berlaku diidentifikasi dan didokumentasikan.", "D"),
				fn("Financial statements are prepared in line with these standards.", "Laporan keuangan disusun sesuai standar tersebut.", "D")),
			pf("Materiality in reporting", "Materialitas pelaporan",
				pr("Materiality levels for reporting are set and approved.", "Tingkat materialitas pelaporan ditetapkan dan disetujui.", "D"),
				fn("Materiality is applied when assessing misstatements and revisited each period.", "Materialitas diterapkan saat menilai salah saji dan ditinjau tiap periode.", "D,W")),
			pf("Non-financial and internal reporting objectives", "Tujuan pelaporan nonkeuangan dan internal",
				pr("Reporting objectives with the required precision are defined for non-financial and internal reports.", "Tujuan pelaporan dengan tingkat ketelitian yang disyaratkan ditetapkan untuk laporan nonkeuangan dan internal.", "D"),
				fn("Reports issued meet the stated precision and timeliness.", "Laporan yang diterbitkan memenuhi ketelitian dan ketepatan waktu yang ditetapkan.", "D,K")),
			pf("Compliance objectives", "Tujuan kepatuhan",
				pr("A register of applicable laws and regulations exists.", "Daftar peraturan perundang-undangan yang berlaku tersedia.", "D"),
				fn("Compliance objectives are derived from the register and updated when regulations change.", "Tujuan kepatuhan diturunkan dari daftar tersebut dan diperbarui saat regulasi berubah.", "D,W")),
			pf("Alignment of objectives across levels", "Keselarasan tujuan antar level",
				pr("Unit objectives cascade from entity objectives.", "Tujuan unit diturunkan dari tujuan entitas.", "D"),
				fn("Unit managers can explain how their objectives support those of the entity.", "Manajer unit dapat menjelaskan kaitan tujuan unitnya dengan tujuan entitas.", "W,K")),
		}},
		{Code: "P07", NameID: "Mengidentifikasi dan menganalisis risiko", Weight: 7000, Owners: "RISK_MANAGEMENT,DEPARTMENT_HEAD", Pofs: []pofDef{
			pf("Risk identification across levels", "Identifikasi risiko di seluruh level",
				pr("A risk management policy and process cover entity, division and unit levels.", "Kebijakan dan proses manajemen risiko mencakup level entitas, divisi, dan unit.", "D"),
				fn("Risk registers exist for each unit and were updated in the period.", "Daftar risiko tersedia untuk tiap unit dan diperbarui pada periode berjalan.", "D")),
			pf("Internal and external factors", "Faktor internal dan eksternal",
				pr("A method to consider internal and external factors in risk identification is defined.", "Metode mempertimbangkan faktor internal dan eksternal dalam identifikasi risiko ditetapkan.", "D"),
				fn("Identification uses input on the economy, regulation, technology, people and processes.", "Identifikasi memakai masukan tentang ekonomi, regulasi, teknologi, SDM, dan proses.", "D,W")),
			pf("Management involvement", "Keterlibatan manajemen",
				pr("Roles in risk identification and analysis are assigned to management levels.", "Peran dalam identifikasi dan analisis risiko ditetapkan pada tingkat manajemen.", "D"),
				fn("Managers at the appropriate levels take part in risk workshops or reviews.", "Manajer pada tingkat yang tepat berpartisipasi dalam lokakarya atau reviu risiko.", "D,W")),
			pf("Significance of risks", "Signifikansi risiko",
				pr("Criteria for likelihood and impact are defined.", "Kriteria kemungkinan dan dampak ditetapkan.", "D"),
				fn("Each identified risk is rated by the criteria and ranked.", "Setiap risiko yang teridentifikasi dinilai menurut kriteria dan diurutkan.", "D")),
			pf("Risk responses", "Respons terhadap risiko",
				pr("Response options (accept, reduce, share, avoid) are defined in the policy.", "Pilihan respons (menerima, mengurangi, membagi, menghindari) ditetapkan dalam kebijakan.", "D"),
				fn("A response and an owner are recorded for risks above appetite.", "Respons dan pemilik dicatat untuk risiko di atas selera risiko.", "D,W")),
		}},
		{Code: "P08", NameID: "Menilai risiko kecurangan", Weight: 6000, Owners: "RISK_MANAGEMENT,DEPARTMENT_HEAD", Pofs: []pofDef{
			pf("Types of fraud", "Jenis kecurangan",
				pr("Fraud types (fraudulent reporting, misappropriation of assets, corruption) are listed in a fraud risk assessment.", "Jenis kecurangan (pelaporan palsu, penyalahgunaan aset, korupsi) dicantumkan dalam penilaian risiko kecurangan.", "D"),
				fn("The assessment is updated periodically and covers the processes of the unit.", "Penilaian diperbarui berkala dan mencakup proses unit.", "D,W")),
			pf("Incentives and pressures", "Insentif dan tekanan",
				pr("The assessment considers incentives and pressures that may drive fraud.", "Penilaian mempertimbangkan insentif dan tekanan yang dapat mendorong kecurangan.", "D"),
				fn("Targets and compensation were examined for such pressures.", "Target dan kompensasi ditelaah dari sisi tekanan tersebut.", "D,W")),
			pf("Opportunities", "Peluang",
				pr("The assessment considers opportunities such as access to assets, override of controls and weak segregation of duties.", "Penilaian mempertimbangkan peluang seperti akses ke aset, pengabaian pengendalian, dan lemahnya pemisahan tugas.", "D"),
				fn("Controls were designed or strengthened to close the opportunities found.", "Pengendalian dirancang atau diperkuat untuk menutup peluang yang ditemukan.", "D,W")),
			pf("Attitudes and rationalisation", "Sikap dan rasionalisasi",
				pr("The assessment considers attitudes and rationalisations that may justify improper acts.", "Penilaian mempertimbangkan sikap dan rasionalisasi yang dapat membenarkan tindakan tidak patut.", "D"),
				fn("Fraud awareness training is given and its results feed back into the assessment.", "Pelatihan kesadaran antikecurangan diberikan dan hasilnya menjadi masukan penilaian.", "D,K")),
		}},
		{Code: "P09", NameID: "Mengidentifikasi dan menganalisis perubahan signifikan", Weight: 6000, Owners: "RISK_MANAGEMENT,DEPARTMENT_HEAD", Pofs: []pofDef{
			pf("Changes in the external environment", "Perubahan lingkungan eksternal",
				pr("A process exists to monitor changes in regulation, the economy and the physical environment.", "Terdapat proses memantau perubahan regulasi, ekonomi, dan lingkungan fisik.", "D"),
				fn("Significant external changes were assessed for their effect on objectives and controls.", "Perubahan eksternal yang signifikan dinilai dampaknya terhadap tujuan dan pengendalian.", "D,W")),
			pf("Changes in the business model", "Perubahan model bisnis",
				pr("Criteria identify changes in the business model (new products, acquisitions, technology).", "Kriteria mengidentifikasi perubahan model bisnis (produk baru, akuisisi, teknologi).", "D"),
				fn("Changes were assessed before implementation and controls were adjusted.", "Perubahan dinilai sebelum diimplementasikan dan pengendalian disesuaikan.", "D,W")),
			pf("Changes in leadership", "Perubahan kepemimpinan",
				pr("A process addresses the effect of changes in leadership and key personnel.", "Terdapat proses menangani dampak pergantian pimpinan dan personel kunci.", "D"),
				fn("Handovers and attitudes towards control after leadership changes were reviewed.", "Serah terima dan sikap terhadap pengendalian setelah pergantian pimpinan ditinjau.", "D,W")),
		}},
	}},
	{Code: "CA", NameID: "Aktivitas Pengendalian", Weight: 20000, Principles: []principleDef{
		{Code: "P10", NameID: "Memilih dan mengembangkan aktivitas pengendalian", Weight: 7000, Owners: "DEPARTMENT_HEAD", Pofs: []pofDef{
			pf("Integration with risk assessment", "Integrasi dengan penilaian risiko",
				pr("Controls are mapped to the risks they address (a control matrix or similar).", "Pengendalian dipetakan ke risiko yang ditanganinya (matriks pengendalian atau sejenisnya).", "D"),
				fn("Each risk above appetite has at least one control that operates.", "Setiap risiko di atas selera risiko memiliki minimal satu pengendalian yang berjalan.", "D,W")),
			pf("Entity-specific factors", "Faktor spesifik entitas",
				pr("Control design considers the environment, complexity and scope of the entity's operations.", "Desain pengendalian mempertimbangkan lingkungan, kompleksitas, dan cakupan operasi entitas.", "D"),
				fn("Controls were adapted where operations differ (branches, products).", "Pengendalian disesuaikan di tempat operasi berbeda (cabang, produk).", "D,W")),
			pf("Relevant business processes", "Proses bisnis yang relevan",
				pr("Significant business processes and their controls are documented.", "Proses bisnis signifikan dan pengendaliannya didokumentasikan.", "D"),
				fn("Process owners know the controls in their process and perform them.", "Pemilik proses mengetahui dan menjalankan pengendalian dalam prosesnya.", "W,O")),
			pf("Mix of control types", "Kombinasi jenis pengendalian",
				pr("Controls include preventive and detective, manual and automated types.", "Pengendalian mencakup jenis preventif dan detektif, manual dan otomatis.", "D"),
				fn("The mix was reviewed for balance and cost.", "Kombinasi ditinjau keseimbangan dan biayanya.", "D,W")),
			pf("Level of application", "Tingkat penerapan",
				pr("Controls are placed at appropriate levels (entity, unit, transaction).", "Pengendalian ditempatkan pada tingkat yang tepat (entitas, unit, transaksi).", "D"),
				fn("Higher-level reviews complement transaction-level controls.", "Reviu tingkat atas melengkapi pengendalian tingkat transaksi.", "D,O")),
			pf("Segregation of duties", "Pemisahan tugas",
				pr("Incompatible duties (authorise, record, custody) are identified.", "Tugas yang tidak boleh digabung (otorisasi, pencatatan, penyimpanan) diidentifikasi.", "D"),
				fn("Duties are separated, or compensating controls exist where separation is impractical.", "Tugas dipisahkan atau tersedia pengendalian kompensasi bila pemisahan tidak praktis.", "D,O")),
		}},
		{Code: "P11", NameID: "Memilih dan mengembangkan pengendalian umum atas teknologi", Weight: 6500, Owners: "HEAD_OF_IT", Pofs: []pofDef{
			pf("Dependency on technology", "Ketergantungan pada teknologi",
				pr("The link between business processes, automated controls and technology is documented.", "Hubungan antara proses bisnis, pengendalian otomatis, dan teknologi didokumentasikan.", "D"),
				fn("Technology dependencies are used when scoping reviews of IT controls.", "Ketergantungan teknologi dipakai saat menentukan cakupan reviu pengendalian TI.", "D,W")),
			pf("Technology infrastructure controls", "Pengendalian infrastruktur teknologi",
				pr("Controls for operations, backup, recovery and monitoring of infrastructure are defined.", "Pengendalian operasi, pencadangan, pemulihan, dan pemantauan infrastruktur ditetapkan.", "D"),
				fn("Backups and recovery tests are performed and logged.", "Pencadangan dan uji pemulihan dilakukan dan dicatat.", "D,O")),
			pf("Security management", "Manajemen keamanan",
				pr("An information security policy covers access, authentication and segregation in systems.", "Kebijakan keamanan informasi mencakup akses, autentikasi, dan pemisahan di dalam sistem.", "D"),
				fn("Access is granted by approval and reviewed periodically; leavers are removed in time.", "Akses diberikan dengan persetujuan dan ditinjau berkala; pegawai yang keluar dihapus tepat waktu.", "D,O")),
			pf("Acquisition, development and maintenance", "Akuisisi, pengembangan, dan pemeliharaan",
				pr("A change and development procedure with testing and approval exists.", "Prosedur perubahan dan pengembangan dengan pengujian dan persetujuan tersedia.", "D"),
				fn("Changes to systems follow the procedure and can be traced.", "Perubahan sistem mengikuti prosedur dan dapat ditelusuri.", "D,O")),
		}},
		{Code: "P12", NameID: "Menerapkan melalui kebijakan dan prosedur", Weight: 6500, Owners: "DEPARTMENT_HEAD", Pofs: []pofDef{
			pf("Policies supporting management directives", "Kebijakan pendukung arahan manajemen",
				pr("Policies and procedures exist for significant processes and tie to objectives and risks.", "Kebijakan dan prosedur tersedia untuk proses signifikan dan terkait tujuan dan risiko.", "D"),
				fn("They are accessible and known to the people who apply them.", "Mudah diakses dan diketahui pihak yang menerapkannya.", "K,W")),
			pf("Responsibility and accountability", "Tanggung jawab dan akuntabilitas",
				pr("Each policy names the unit or role responsible for executing it.", "Setiap kebijakan menyebut unit atau peran yang bertanggung jawab melaksanakannya.", "D"),
				fn("The responsible people carry out the policy and answer for its execution.", "Pihak yang bertanggung jawab melaksanakan kebijakan dan mempertanggungjawabkan pelaksanaannya.", "W,O")),
			pf("Timely performance", "Pelaksanaan tepat waktu",
				pr("Timing requirements for performing controls are stated.", "Persyaratan waktu pelaksanaan pengendalian dinyatakan.", "D"),
				fn("Controls are performed on time according to the samples examined.", "Pengendalian dilaksanakan tepat waktu berdasarkan sampel yang diperiksa.", "D,O")),
			pf("Corrective action", "Tindakan korektif",
				pr("A procedure for investigating and resolving exceptions exists.", "Prosedur menyelidiki dan menyelesaikan pengecualian tersedia.", "D"),
				fn("Exceptions are followed up with documented corrective action.", "Pengecualian ditindaklanjuti dengan tindakan korektif yang terdokumentasi.", "D,W")),
			pf("Competent personnel", "Personel yang kompeten",
				pr("The competence needed to perform controls is defined.", "Kompetensi yang diperlukan untuk melaksanakan pengendalian ditetapkan.", "D"),
				fn("People who perform controls have the competence, time and authority to do so.", "Pelaksana pengendalian memiliki kompetensi, waktu, dan wewenang yang cukup.", "W,K")),
			pf("Reassessing policies", "Penilaian ulang kebijakan",
				pr("A review cycle for policies and procedures is defined.", "Siklus reviu kebijakan dan prosedur ditetapkan.", "D"),
				fn("Reviews took place in the cycle and updates were approved.", "Reviu dilakukan sesuai siklus dan pembaruan disetujui.", "D")),
		}},
	}},
	{Code: "IC", NameID: "Informasi dan Komunikasi", Weight: 17500, Principles: []principleDef{
		{Code: "P13", NameID: "Menggunakan informasi yang relevan", Weight: 6000, Owners: "DEPARTMENT_HEAD,HEAD_OF_IT", Pofs: []pofDef{
			pf("Information requirements", "Kebutuhan informasi",
				pr("The information needed for each key process and control is identified.", "Informasi yang dibutuhkan untuk tiap proses dan pengendalian utama diidentifikasi.", "D"),
				fn("Information requirements are revisited when objectives change.", "Kebutuhan informasi ditinjau ulang saat tujuan berubah.", "D,W")),
			pf("Internal and external data sources", "Sumber data internal dan eksternal",
				pr("Sources of data (systems, regulators, customers, market) are inventoried.", "Sumber data (sistem, regulator, pelanggan, pasar) diinventarisasi.", "D"),
				fn("Data is captured from these sources regularly.", "Data diambil dari sumber tersebut secara berkala.", "D,O")),
			pf("Processing data into information", "Pengolahan data menjadi informasi",
				pr("Reports and dashboards turn data into information for decision-makers.", "Laporan dan dasbor mengolah data menjadi informasi bagi pengambil keputusan.", "D"),
				fn("Decision-makers use this information in their decisions.", "Pengambil keputusan menggunakan informasi tersebut dalam keputusan.", "W,K")),
			pf("Quality of information", "Kualitas informasi",
				pr("Information quality criteria (accurate, complete, timely, accessible) and data ownership are defined.", "Kriteria kualitas informasi (akurat, lengkap, tepat waktu, mudah diakses) dan kepemilikan data ditetapkan.", "D"),
				fn("Data quality checks (reconciliation, validation) run and issues are corrected.", "Pemeriksaan kualitas data (rekonsiliasi, validasi) berjalan dan masalah dikoreksi.", "D,O")),
			pf("Cost and benefit", "Biaya dan manfaat",
				pr("The cost and benefit of information systems and reports are considered.", "Biaya dan manfaat sistem informasi dan laporan dipertimbangkan.", "D"),
				fn("Reports that no longer serve a purpose are retired or revised.", "Laporan yang tidak lagi diperlukan dihentikan atau diperbaiki.", "W")),
		}},
		{Code: "P14", NameID: "Berkomunikasi secara internal", Weight: 6000, Owners: "DEPARTMENT_HEAD,HEAD_OF_IT", Pofs: []pofDef{
			pf("Internal control information", "Informasi pengendalian intern",
				pr("Policies explain control objectives and responsibilities to staff.", "Kebijakan menjelaskan tujuan pengendalian dan tanggung jawab kepada pegawai.", "D"),
				fn("Staff understand their control responsibilities.", "Pegawai memahami tanggung jawab pengendaliannya.", "K")),
			pf("Communication with the board", "Komunikasi dengan dewan",
				pr("Regular reporting from management to the oversight body on control matters is defined.", "Pelaporan berkala dari manajemen ke badan pengawas mengenai pengendalian ditetapkan.", "D"),
				fn("Reports are submitted on schedule and discussed.", "Laporan disampaikan sesuai jadwal dan dibahas.", "D,W")),
			pf("Separate communication lines", "Jalur komunikasi terpisah",
				pr("Anonymous or independent lines (hotline) are available when normal lines fail.", "Jalur anonim atau independen (hotline) tersedia ketika jalur normal gagal.", "D"),
				fn("Staff know and trust these lines and the reports received are handled.", "Pegawai mengetahui dan memercayai jalur ini dan laporan yang diterima ditangani.", "K,D")),
			pf("Methods of communication", "Metode komunikasi",
				pr("Communication methods (meetings, memos, intranet) are chosen by audience and urgency.", "Metode komunikasi (rapat, memo, intranet) dipilih menurut audiens dan urgensi.", "D"),
				fn("Messages reach the intended audience in time.", "Pesan sampai ke audiens yang dituju tepat waktu.", "K,W")),
		}},
		{Code: "P15", NameID: "Berkomunikasi secara eksternal", Weight: 5500, Owners: "CORPORATE_COMMUNICATION,CORPORATE_SECRETARY", Pofs: []pofDef{
			pf("Communication to external parties", "Komunikasi ke pihak eksternal",
				pr("A policy governs communication with regulators, customers, shareholders and suppliers.", "Kebijakan mengatur komunikasi dengan regulator, pelanggan, pemegang saham, dan pemasok.", "D"),
				fn("External communications are authorised and consistent.", "Komunikasi eksternal diotorisasi dan konsisten.", "D,W")),
			pf("Inbound communications", "Komunikasi masuk",
				pr("Channels exist for customers, suppliers and others to give feedback and complaints.", "Kanal tersedia bagi pelanggan, pemasok, dan pihak lain untuk memberi masukan dan keluhan.", "D"),
				fn("Inbound items are logged, handled and brought to the attention of management.", "Masukan dicatat, ditangani, dan disampaikan kepada manajemen.", "D,O")),
			pf("External information reaches the board", "Informasi eksternal sampai ke dewan",
				pr("A process passes relevant external feedback and regulator results to the oversight body.", "Proses meneruskan masukan eksternal dan hasil pemeriksaan regulator ke badan pengawas.", "D"),
				fn("Such information reached the oversight body in the period.", "Informasi tersebut telah sampai ke badan pengawas pada periode berjalan.", "D,W")),
			pf("Separate external communication lines", "Jalur komunikasi eksternal terpisah",
				pr("Independent external reporting channels (such as a whistleblower portal) are publicly known.", "Kanal pelaporan eksternal independen (seperti portal pelaporan pelanggaran) dipublikasikan.", "D"),
				fn("Reports through these channels are handled according to procedure.", "Laporan melalui kanal ini ditangani sesuai prosedur.", "D,W")),
			pf("Methods of external communication", "Metode komunikasi eksternal",
				pr("Methods and timing for external communication are defined, including legal limits.", "Metode dan waktu komunikasi eksternal ditetapkan, termasuk batasan hukum.", "D"),
				fn("Disclosures are accurate and on time.", "Pengungkapan akurat dan tepat waktu.", "D")),
		}},
	}},
	{Code: "MA", NameID: "Aktivitas Pemantauan", Weight: 12500, Principles: []principleDef{
		{Code: "P16", NameID: "Melakukan evaluasi berkelanjutan dan/atau terpisah", Weight: 6500, Owners: "MANAGEMENT", Pofs: []pofDef{
			pf("Mix of ongoing and separate evaluations", "Kombinasi evaluasi berkelanjutan dan terpisah",
				pr("A monitoring plan combines ongoing monitoring and periodic separate evaluations.", "Rencana pemantauan memadukan pemantauan berkelanjutan dan evaluasi terpisah berkala.", "D"),
				fn("The plan was carried out in the period.", "Rencana dilaksanakan pada periode berjalan.", "D,W")),
			pf("Rate of change", "Laju perubahan",
				pr("Monitoring frequency takes the pace of change in operations and risk into account.", "Frekuensi pemantauan mempertimbangkan laju perubahan operasi dan risiko.", "D"),
				fn("Frequency was increased in areas with significant change.", "Frekuensi ditingkatkan pada area dengan perubahan signifikan.", "D,W")),
			pf("Baseline understanding", "Pemahaman awal (baseline)",
				pr("A baseline of the design and status of the control system exists.", "Baseline desain dan status sistem pengendalian tersedia.", "D"),
				fn("Evaluations compare the current state with the baseline.", "Evaluasi membandingkan kondisi saat ini dengan baseline.", "D")),
			pf("Knowledgeable personnel", "Personel yang berpengetahuan",
				pr("Competence requirements for evaluators are defined.", "Persyaratan kompetensi evaluator ditetapkan.", "D"),
				fn("Evaluators understand what they evaluate and are independent of the area.", "Evaluator memahami yang dievaluasi dan independen dari area tersebut.", "D,W")),
			pf("Integration with business processes", "Integrasi dengan proses bisnis",
				pr("Ongoing monitoring (exception reports, key risk indicators, supervisory review) is built into processes.", "Pemantauan berkelanjutan (laporan eksepsi, indikator risiko utama, reviu berjenjang) menyatu dalam proses.", "D"),
				fn("Monitoring outputs are reviewed and acted on in the normal course of work.", "Hasil pemantauan ditinjau dan ditindaklanjuti dalam kegiatan rutin.", "D,O")),
			pf("Scope and frequency", "Cakupan dan frekuensi",
				pr("Criteria to adjust scope and frequency by risk are defined.", "Kriteria menyesuaikan cakupan dan frekuensi berdasarkan risiko ditetapkan.", "D"),
				fn("Scope and frequency were adjusted according to the criteria.", "Cakupan dan frekuensi disesuaikan menurut kriteria.", "D,W")),
			pf("Objective evaluation", "Evaluasi objektif",
				pr("Separate evaluations are done by units independent of the area evaluated (internal audit, quality assurance, external parties).", "Evaluasi terpisah dilakukan unit yang independen dari area yang dievaluasi (audit internal, penjaminan kualitas, pihak eksternal).", "D"),
				fn("Evaluation results are objective and not influenced by those evaluated.", "Hasil evaluasi objektif dan tidak dipengaruhi pihak yang dievaluasi.", "W,K")),
			pf("Evaluation of the internal audit function", "Evaluasi fungsi audit internal",
				pri("A quality assurance and improvement programme for the internal audit function exists.", "Program penjaminan dan peningkatan kualitas fungsi audit internal tersedia.", "D"),
				fi("An internal or external assessment of the internal audit function was performed by someone other than the chief audit executive.", "Penilaian internal atau eksternal atas fungsi audit internal dilakukan oleh pihak selain kepala audit internal.", "D,W")),
		}},
		{Code: "P17", NameID: "Mengevaluasi dan mengomunikasikan kekurangan", Weight: 6000, Owners: "MANAGEMENT", Pofs: []pofDef{
			pf("Assessing results", "Penilaian hasil",
				pr("Criteria to evaluate monitoring results and rate deficiencies exist.", "Kriteria menilai hasil pemantauan dan menetapkan tingkat kekurangan tersedia.", "D"),
				fn("Management assessed the results of the period against the criteria.", "Manajemen menilai hasil periode berjalan menurut kriteria.", "D,W")),
			pf("Communicating deficiencies", "Komunikasi kekurangan",
				pr("Escalation rules exist for deficiencies to responsible parties, senior management and the oversight body.", "Aturan eskalasi kekurangan kepada pihak bertanggung jawab, manajemen senior, dan badan pengawas tersedia.", "D"),
				fn("Deficiencies were reported to the responsible parties and escalated as required.", "Kekurangan dilaporkan ke pihak yang bertanggung jawab dan dieskalasi sesuai ketentuan.", "D,W")),
			pf("Monitoring corrective actions", "Pemantauan tindakan korektif",
				pr("Remediation plans with owners and due dates are recorded for deficiencies.", "Rencana perbaikan dengan penanggung jawab dan tenggat dicatat untuk setiap kekurangan.", "D"),
				fn("Corrective actions are tracked to closure and overdue items are escalated.", "Tindakan korektif dipantau hingga selesai dan yang terlambat dieskalasi.", "D,W")),
		}},
	}},
}
