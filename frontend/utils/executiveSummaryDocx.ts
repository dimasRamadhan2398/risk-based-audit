/**
 * Native Microsoft Word (.docx) generator for Executive Summary Compilation
 * ("Executive Summary Compilation Hasil Audit Triwulanan") and Executive Summary Individual.
 *
 * Rule: A Word (.docx) document can ONLY be generated when the summary status is 'Approved'.
 * Draft or Rejected summaries are rejected before any file is built.
 */

import type { ExecutiveSummary, FollowUpRow, MatriksRow, SignificantFinding } from '~/stores/executive-summary'

export type ExecutiveSummaryDocxVariant = 'compilation' | 'individual'

export interface ExecutiveSummaryDocxContext {
  variant?: ExecutiveSummaryDocxVariant
  reportTitle?: string
  auditObject?: string
  department?: string
  reportDate?: string
  formatDepartmentName?: (val?: string) => string
}

export const isCompilationDocNumber = (nomorDokumen?: string | null): boolean => {
  const n = String(nomorDokumen ?? '').trim().toUpperCase()
  return n.startsWith('ESC-') || n.startsWith('EXSUM-KOMP-') || n.startsWith('KOMPILASI-')
}

export const resolveExecutiveSummaryVariant = (
  summary: Pick<ExecutiveSummary, 'nomorDokumen'>,
  explicit?: ExecutiveSummaryDocxVariant
): ExecutiveSummaryDocxVariant => {
  if (explicit === 'compilation' || explicit === 'individual') return explicit
  return isCompilationDocNumber(summary.nomorDokumen) ? 'compilation' : 'individual'
}

export const canGenerateExecutiveSummaryDocx = (
  summary?: Pick<ExecutiveSummary, 'status'> | null
): boolean => String(summary?.status ?? '').trim() === 'Approved'

const xmlEsc = (value: unknown): string =>
  String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')

const riskLabelId = (raw?: string | null): string => {
  const v = String(raw ?? '').trim()
  if (!v) return '-'
  const lower = v.toLowerCase()
  if (lower === 'very significant' || lower === 'significant' || lower === 'high' || lower === 'moderate-high' || lower === 'tinggi') {
    return 'Tinggi'
  }
  if (lower === 'quite significant' || lower === 'moderate' || lower === 'sedang') {
    return 'Sedang'
  }
  if (lower === 'not significant' || lower === 'low' || lower === 'low-moderate' || lower === 'rendah') {
    return 'Rendah'
  }
  return v
}

const quarterMonthsLabel = (quarter: number): string => {
  switch (Number(quarter)) {
    case 1: return 'JANUARI – MARET (TRIWULAN I)'
    case 2: return 'APRIL – JUNI (TRIWULAN II)'
    case 3: return 'JULI – SEPTEMBER (TRIWULAN III)'
    case 4: return 'OKTOBER – DESEMBER (TRIWULAN IV)'
    default: return 'TRIWULAN I'
  }
}

const formatIdDate = (raw?: string | null, fallbackYear?: number): string => {
  const s = String(raw ?? '').trim()
  if (!s) return `................ ${fallbackYear || 2026}`
  const m = s.match(/^(\d{4})-(\d{1,2})-(\d{1,2})/)
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ]
  if (m) {
    const year = Number(m[1])
    const month = months[Number(m[2]) - 1] || m[2]
    const day = Number(m[3])
    return `${day} ${month} ${year}`
  }
  return s
}

const headingPara = (text: string): string =>
  `<w:p><w:pPr><w:spacing w:before="220" w:after="80"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="22"/><w:color w:val="003366"/></w:rPr><w:t xml:space="preserve">${xmlEsc(text)}</w:t></w:r></w:p>`

const centerTitlePara = (text: string, sizeHalfPt = 26, bold = true): string =>
  `<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:before="40" w:after="40"/></w:pPr><w:r><w:rPr>${bold ? '<w:b/>' : ''}<w:sz w:val="${sizeHalfPt}"/></w:rPr><w:t xml:space="preserve">${xmlEsc(text)}</w:t></w:r></w:p>`

const textPara = (text: string, bold = false): string => {
  const cleaned = String(text ?? '').trim() || '-'
  const lines = cleaned.replace(/\r\n/g, '\n').split('\n').map(l => l.trim()).filter(Boolean)
  const rPr = bold ? '<w:rPr><w:b/><w:sz w:val="20"/></w:rPr>' : '<w:rPr><w:sz w:val="20"/></w:rPr>'
  return (lines.length ? lines : ['-'])
    .map(line => `<w:p><w:pPr><w:spacing w:after="80"/></w:pPr><w:r>${rPr}<w:t xml:space="preserve">${xmlEsc(line)}</w:t></w:r></w:p>`)
    .join('')
}

const tableOpen = (widths: number[], totalWidth = 9600): string => {
  const borders = ['top', 'left', 'bottom', 'right', 'insideH', 'insideV']
    .map(side => `<w:${side} w:val="single" w:sz="4" w:space="0" w:color="808080"/>`)
    .join('')
  const grid = widths.map(w => `<w:gridCol w:w="${w}"/>`).join('')
  return `<w:tbl><w:tblPr><w:tblW w:w="${totalWidth}" w:type="dxa"/><w:tblBorders>${borders}</w:tblBorders><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="50" w:type="dxa"/><w:left w:w="80" w:type="dxa"/><w:bottom w:w="50" w:type="dxa"/><w:right w:w="80" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>${grid}</w:tblGrid>`
}

const tableCell = (width: number, text: unknown, opts: { bold?: boolean, shade?: string, align?: 'left' | 'center' | 'right', fontSize?: number } = {}): string => {
  const shd = opts.shade ? `<w:shd w:val="clear" w:color="auto" w:fill="${opts.shade}"/>` : ''
  const jc = opts.align ? `<w:jc w:val="${opts.align}"/>` : ''
  const sz = opts.fontSize ?? 18
  const rPr = `<w:rPr>${opts.bold ? '<w:b/>' : ''}<w:sz w:val="${sz}"/></w:rPr>`
  const raw = String(text ?? '').trim()
  const val = raw === '' ? '-' : raw
  return `<w:tc><w:tcPr><w:tcW w:w="${width}" w:type="dxa"/>${shd}</w:tcPr><w:p><w:pPr>${jc}<w:spacing w:after="40"/></w:pPr><w:r>${rPr}<w:t xml:space="preserve">${xmlEsc(val)}</w:t></w:r></w:p></w:tc>`
}

const tableRow = (widths: number[], cells: unknown[], opts: { header?: boolean, bold?: boolean, shade?: string, aligns?: Array<'left' | 'center' | 'right'>, fontSize?: number } = {}): string => {
  const trPr = opts.header ? '<w:trPr><w:tblHeader/></w:trPr>' : ''
  const inner = widths.map((w, i) =>
    tableCell(w, cells[i] ?? '-', {
      bold: opts.bold ?? opts.header,
      shade: opts.shade ?? (opts.header ? 'D9E2F3' : undefined),
      align: opts.aligns?.[i],
      fontSize: opts.fontSize
    })
  ).join('')
  return `<w:tr>${trPr}${inner}</w:tr>`
}

const buildFollowUpTableXml = (rows: readonly FollowUpRow[] | null | undefined, highRiskCount: number): string => {
  const list = Array.isArray(rows) && rows.length > 0
    ? rows
    : [
        { status: 'Closed' as const, jumlah: 0, persentase: 0, keterangan: '' },
        { status: 'In Progress' as const, jumlah: 0, persentase: 0, keterangan: '' },
        { status: 'Overdue' as const, jumlah: 0, persentase: 0, keterangan: '' }
      ]
  const total = list.reduce((sum, r) => sum + (Number(r?.jumlah) || 0), 0)
  const widths = [2200, 1400, 1400, 4600]
  const defaultNote = (status: string, jumlah: number): string => {
    if (status === 'Closed') return jumlah > 0 ? 'Telah divalidasi SPI' : '-'
    if (status === 'In Progress') return jumlah > 0 ? 'On track pelaksanaan tindak lanjut' : '-'
    if (status === 'Overdue') return jumlah > 0 && highRiskCount > 0 ? `${Math.min(jumlah, highRiskCount)} Risiko Tinggi` : (jumlah > 0 ? 'Melewati batas waktu' : '-')
    return '-'
  }

  // Show the 3 core rows (Closed, In Progress, Overdue) and No Action if non-zero
  const visibleRows = list.filter(r => r.status !== 'No Action' || (Number(r.jumlah) || 0) > 0)
  let xml = tableOpen(widths)
  xml += tableRow(widths, ['Status', 'Jumlah', '%', 'Keterangan'], {
    header: true,
    aligns: ['left', 'center', 'center', 'left']
  })
  for (const r of visibleRows) {
    const count = Number(r.jumlah) || 0
    const pct = total > 0 ? ((count / total) * 100).toFixed(1) : (Number(r.persentase) || 0).toFixed(1)
    const note = String(r.keterangan ?? '').trim() || defaultNote(r.status, count)
    xml += tableRow(widths, [r.status, String(count), `${pct}%`, note], {
      aligns: ['left', 'center', 'center', 'left']
    })
  }
  xml += tableRow(widths, ['Total', String(total), total > 0 ? '100%' : '0%', '-'], {
    bold: true,
    shade: 'F2F4F8',
    aligns: ['left', 'center', 'center', 'left']
  })
  xml += '</w:tbl><w:p><w:pPr><w:spacing w:after="100"/></w:pPr></w:p>'
  return xml
}

const buildTopFindingsTableXml = (
  findings: readonly SignificantFinding[] | null | undefined,
  fmtDept: (val?: string) => string
): string => {
  const widths = [600, 1600, 2800, 1300, 1400, 1900]
  const list = Array.isArray(findings) ? findings : []
  let xml = tableOpen(widths)
  xml += tableRow(widths, ['No', 'Unit Division', 'Judul Temuan', 'Risiko', 'Status TL', 'Usulan'], {
    header: true,
    aligns: ['center', 'left', 'left', 'center', 'center', 'left']
  })
  if (list.length === 0) {
    xml += tableRow(widths, ['-', '-', 'Tidak ada temuan signifikan pada periode ini', '-', '-', '-'], {
      aligns: ['center', 'left', 'left', 'center', 'center', 'left']
    })
  } else {
    list.forEach((f, idx) => {
      const div = fmtDept(f.unitDivision) || f.unitDivision || '-'
      const statusLabel = f.statusTL === 'Overdue' ? 'Overdue/Telat' : (f.statusTL || '-')
      xml += tableRow(widths, [
        String(idx + 1),
        div,
        f.judulTemuan || '-',
        riskLabelId(f.risiko),
        statusLabel,
        f.usulan || '-'
      ], {
        aligns: ['center', 'left', 'left', 'center', 'center', 'left']
      })
    })
  }
  xml += '</w:tbl><w:p><w:pPr><w:spacing w:after="100"/></w:pPr></w:p>'
  return xml
}

const buildCompilationMatrixTableXml = (
  rows: readonly MatriksRow[] | null | undefined,
  fmtDept: (val?: string) => string
): string => {
  const widths = [900, 700, 1100, 900, 1300, 700, 1100, 750, 750, 550, 700, 750]
  const totalWidth = widths.reduce((a, b) => a + b, 0)
  const list = Array.isArray(rows) ? rows : []
  let xml = tableOpen(widths, totalWidth)
  xml += tableRow(
    widths,
    ['Nomor', 'Division', 'Unit Kerja', 'Proses Bisnis', 'Judul Temuan', 'Nilai Risiko', 'Rekomendasi', 'Due Date', 'PIC Unit', '% Progres', 'Status', 'Bukti TL'],
    { header: true, fontSize: 15 }
  )
  xml += tableRow(
    widths,
    ['(A)', '(B)', '(C)', '(D)', '(E)', '(F)', '(G)', '(H)', '(I)', '(J)', '(K)', '(L)'],
    { header: true, shade: 'EDF2F8', aligns: widths.map(() => 'center'), fontSize: 15 }
  )
  if (list.length === 0) {
    xml += tableRow(
      widths,
      ['-', '-', '-', '-', '-', '-', '-', '-', '-', '0%', '-', '-'],
      { fontSize: 15 }
    )
  } else {
    for (const r of list) {
      const divCode = r.division || '-'
      const unitKerja = fmtDept(r.unitKerja || r.division) || r.unitKerja || r.division || '-'
      xml += tableRow(
        widths,
        [
          r.nomor || '-',
          divCode,
          unitKerja,
          r.prosesBisnis || '-',
          r.judulTemuan || '-',
          riskLabelId(r.nilaiRisiko),
          r.rekomendasi || '-',
          r.dueDate || '-',
          r.picUnit || '-',
          `${Number(r.progres) || 0}%`,
          r.status || 'No Action',
          r.buktiTL || '-'
        ],
        { fontSize: 15 }
      )
    }
  }
  xml += '</w:tbl><w:p><w:pPr><w:spacing w:after="100"/></w:pPr></w:p>'
  return xml
}

const buildSignatureBlockXml = (summary: ExecutiveSummary): string => {
  const place = String(summary.signatureTempat || 'Jakarta').trim()
  const dateLabel = formatIdDate(summary.signatureTanggal || summary.approvedAt, summary.tahun)
  const headName = String(summary.signatureNamaKepala || (summary as any).signatureName || '[Nama Kepala SPI]').trim()
  const nik = String(summary.signatureNIK || '........................').trim()

  return [
    '<w:p><w:pPr><w:spacing w:before="240" w:after="40"/></w:pPr></w:p>',
    `<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="40"/></w:pPr><w:r><w:rPr><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">${xmlEsc(`${place}, ${dateLabel}`)}</w:t></w:r></w:p>`,
    `<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="600"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">Kepala SPI,</w:t></w:r></w:p>`,
    `<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="40"/></w:pPr><w:r><w:rPr><w:b/><w:u w:val="single"/><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">${xmlEsc(headName)}</w:t></w:r></w:p>`,
    `<w:p><w:pPr><w:ind w:left="5400"/><w:spacing w:after="120"/></w:pPr><w:r><w:rPr><w:sz w:val="20"/></w:rPr><w:t xml:space="preserve">${xmlEsc(`NIK. ${nik}`)}</w:t></w:r></w:p>`
  ].join('')
}

/**
 * Builds the inner `word/document.xml` string for an Approved Executive Summary
 * (either Compilation "Executive Summary Compilation Hasil Audit Triwulanan" or Individual).
 */
export const buildExecutiveSummaryDocumentXml = (
  summary: ExecutiveSummary,
  context: ExecutiveSummaryDocxContext = {}
): string => {
  if (!canGenerateExecutiveSummaryDocx(summary)) {
    throw new Error('Word (.docx) can only be generated for an Approved Executive Summary')
  }

  const variant = resolveExecutiveSummaryVariant(summary, context.variant)
  const fmtDept = context.formatDepartmentName || ((v?: string) => v || '-')
  const year = Number(summary.tahun) || 2026
  const quarter = Number(summary.quarter) || 1
  const highRisk = (Number(summary.risikoHigh) || 0) + (Number(summary.risikoModerateHigh) || 0)
  const medRisk = Number(summary.risikoModerate) || 0
  const lowRisk = (Number(summary.risikoLow) || 0) + (Number(summary.risikoLowModerate) || 0)
  const totalFindings = highRisk + medRisk + lowRisk || summary.matriksKompilasi?.length || summary.topFindings?.length || 0
  const recCount = Number(summary.jumlahRekomendasi) || totalFindings
  const statusUntilDate = formatIdDate(summary.approvedAt || summary.signatureTanggal, year)

  let body = ''

  if (variant === 'compilation') {
    const periodHeader = summary.periodeBulan
      ? `${String(summary.periodeBulan).toUpperCase()} (${quarterMonthsLabel(quarter)}) TAHUN ${year}`
      : `BULAN ${quarterMonthsLabel(quarter)} TAHUN ${year}`
    const docNum = summary.nomorDokumen || `LKA/B/00${quarter}/SPI/KAI/${year}`
    const lhaCount = Number(summary.jumlahLaporan) || 0

    // Header
    body += centerTitlePara('EXECUTIVE SUMMARY COMPILATION', 28, true)
    body += centerTitlePara('HASIL AUDIT TRIWULANAN', 26, true)
    body += centerTitlePara(`PERIODE: ${periodHeader}`, 22, true)
    body += centerTitlePara(`Nomor: ${docNum}`, 22, true)
    body += '<w:p><w:pPr><w:spacing w:after="120"/></w:pPr></w:p>'

    // I. EXECUTIVE SUMMARY
    body += headingPara('I. EXECUTIVE SUMMARY')
    body += textPara(summary.narrative)

    // II. STATISTIK KOMPILASI
    body += headingPara('II. STATISTIK KOMPILASI')
    body += textPara(`Jumlah Laporan: Internal ${lhaCount} LHA.`)
    body += textPara(`Jumlah Temuan: ${totalFindings} temuan. Rincian: Risiko Tinggi ${highRisk}, Sedang ${medRisk}, Rendah ${lowRisk}.`)
    body += textPara(`Jumlah Rekomendasi: ${recCount} rekomendasi.`)

    // III. STATUS TINDAK LANJUT
    body += headingPara(`III. STATUS TINDAK LANJUT s.d ${statusUntilDate}`)
    body += buildFollowUpTableXml(summary.followUpTable, highRisk)

    // IV. TOP 5 TEMUAN SIGNIFIKAN
    body += headingPara('IV. TOP 5 TEMUAN SIGNIFIKAN')
    body += buildTopFindingsTableXml(summary.topFindings, fmtDept)

    // V. ANALISIS TEMUAN BERULANG
    body += headingPara('V. ANALISIS TEMUAN BERULANG')
    const defaultAkar = summary.topFindings?.length
      ? `Tema berulang periode ini: ${summary.topFindings.slice(0, 2).map((f, i) => `${i + 1}. ${f.judulTemuan}`).join('; ')}.\nAkar masalah: SOP belum update dan pengawasan pengendalian internal perlu diperkuat. Usulan: Revisi kebijakan dan penguatan monitoring.`
      : 'Tema berulang periode ini: -\nAkar masalah: -'
    body += textPara(summary.akarMasalah || defaultAkar)

    // VI. TREN & GRAFIK
    body += headingPara('VI. TREN & GRAFIK')
    const fuTotal = (summary.followUpTable || []).reduce((s, r) => s + (Number(r.jumlah) || 0), 0)
    const closedRow = (summary.followUpTable || []).find(r => r.status === 'Closed')
    const onTimePct = fuTotal > 0 ? (((Number(closedRow?.jumlah) || 0) / fuTotal) * 100).toFixed(1) : '0.0'
    body += textPara('[Lampirkan Grafik: Jumlah Temuan per Bulan, Aging Rekomendasi, % On-time]')
    body += textPara(`Ringkasan Tren Periode Triwulan ${quarter} Tahun ${year}: Total ${lhaCount} LHA diterbitkan dengan ${totalFindings} temuan (${highRisk} Risiko Tinggi, ${medRisk} Sedang, ${lowRisk} Rendah) dan tingkat penyelesaian (On-time/Closed) sebesar ${onTimePct}%.`)

    // VII. KESIMPULAN & REKOMENDASI UNTUK MANAJEMEN
    body += headingPara('VII. KESIMPULAN & REKOMENDASI UNTUK MANAJEMEN')
    const defaultKesimpulan = [
      'Secara umum pengendalian internal terus diperkuat dan tindak lanjut rekomendasi dipantau secara berkala.',
      `Mohon arahan Direksi untuk percepatan penyelesaian rekomendasi risiko tinggi.`,
      'Usulan penguatan koordinasi lintas unit untuk mencegah temuan berulang.'
    ].join('\n')
    body += textPara(summary.kesimpulan || defaultKesimpulan)
    if (summary.executiveNote && summary.executiveNote.trim()) {
      body += textPara(`Catatan Executive: ${summary.executiveNote.trim()}`)
    }

    // Signature
    body += buildSignatureBlockXml(summary)

    // Page break + Lampiran Matriks Induk Kompilasi Temuan
    body += '<w:p><w:r><w:br w:type="page"/></w:r></w:p>'
    body += textPara('Lampiran.', true)
    body += centerTitlePara('MATRIKS INDUK KOMPILASI TEMUAN', 24, true)
    body += centerTitlePara(`LAPORAN KOMPILASI HASIL AUDIT PERIODE: ${periodHeader}`, 20, true)
    body += centerTitlePara(`Nomor: ${docNum}`, 20, true)
    body += '<w:p><w:pPr><w:spacing w:after="80"/></w:pPr></w:p>'
    body += buildCompilationMatrixTableXml(summary.matriksKompilasi, fmtDept)
    body += textPara('Keterangan:', true)
    body += textPara('Kolom C: Unit Kerja\nKolom F: Tinggi; Sedang; Rendah\nKolom K: Closed; In Progress; Overdue')
  } else {
    // INDIVIDUAL EXECUTIVE SUMMARY TEMPLATE (adapted strictly to Individual feature data)
    const periodLabel = summary.periodeBulan
      ? `${String(summary.periodeBulan).toUpperCase()} ${String(summary.periodeBulan).includes(String(year)) ? '' : `TAHUN ${year}`}`.trim()
      : `TAHUN ${year}`
    const lhaNum = summary.nomorDokumen || '-'
    const stNum = summary.assignmentLetterId || '-'
    const reportTitle = context.reportTitle || context.auditObject || `Laporan Hasil Audit (${lhaNum})`
    const divisionLabel = fmtDept(context.department || summary.topFindings?.[0]?.unitDivision || '-')

    body += centerTitlePara('EXECUTIVE SUMMARY INDIVIDUAL', 28, true)
    body += centerTitlePara('LAPORAN HASIL AUDIT (LHA)', 26, true)
    body += centerTitlePara(`PERIODE: ${periodLabel}`, 22, true)
    body += centerTitlePara(`Nomor LHA: ${lhaNum}`, 22, true)
    if (stNum && stNum !== '-') {
      body += centerTitlePara(`Surat Tugas: ${stNum}`, 20, false)
    }
    body += '<w:p><w:pPr><w:spacing w:after="120"/></w:pPr></w:p>'

    // I. EXECUTIVE SUMMARY
    body += headingPara('I. EXECUTIVE SUMMARY')
    body += textPara(summary.narrative)

    // II. INFORMASI & STATISTIK LHA INDIVIDUAL
    body += headingPara('II. INFORMASI & STATISTIK LHA INDIVIDUAL')
    body += textPara(`Nomor LHA: ${lhaNum}`)
    body += textPara(`Nomor Surat Tugas: ${stNum}`)
    body += textPara(`Judul Laporan / Objek Audit: ${reportTitle}`)
    body += textPara(`Unit / Divisi: ${divisionLabel}`)
    body += textPara(`Jumlah Temuan: ${totalFindings} temuan. Rincian: Risiko Tinggi (Signifikan) ${highRisk}, Sedang (Cukup Signifikan) ${medRisk}, Rendah (Tidak Signifikan) ${lowRisk}.`)
    body += textPara(`Jumlah Rekomendasi: ${recCount} rekomendasi.`)

    // III. STATUS TINDAK LANJUT
    body += headingPara(`III. STATUS TINDAK LANJUT s.d ${statusUntilDate}`)
    body += buildFollowUpTableXml(summary.followUpTable, highRisk)

    // IV. TOP 5 TEMUAN SIGNIFIKAN
    body += headingPara('IV. TOP 5 TEMUAN SIGNIFIKAN')
    body += buildTopFindingsTableXml(summary.topFindings, fmtDept)

    // V. ANALISIS TEMUAN BERULANG & AKAR MASALAH
    body += headingPara('V. ANALISIS TEMUAN BERULANG & AKAR MASALAH')
    const defaultAkar = summary.topFindings?.length
      ? `Analisis temuan berulang mengidentifikasi kelemahan pada kepatuhan SOP dan monitoring: ${summary.topFindings.slice(0, 2).map(f => f.judulTemuan).join('; ')}.`
      : '-'
    body += textPara(summary.akarMasalah || defaultAkar)

    // VI. KESIMPULAN & REKOMENDASI UNTUK MANAJEMEN
    body += headingPara('VI. KESIMPULAN & REKOMENDASI UNTUK MANAJEMEN')
    const defaultKesimpulan = `Tata kelola dan pengendalian internal pada ${reportTitle} berjalan cukup memadai dengan prioritas penyelesaian rekomendasi temuan signifikan.`
    body += textPara(summary.kesimpulan || defaultKesimpulan)
    if (summary.executiveNote && summary.executiveNote.trim()) {
      body += textPara(`Catatan Executive: ${summary.executiveNote.trim()}`)
    }

    // Signature
    body += buildSignatureBlockXml(summary)
  }

  return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    ${body}
    <w:sectPr>
      <w:pgSz w:w="11906" w:h="16838"/>
      <w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134" w:header="708" w:footer="708" w:gutter="0"/>
    </w:sectPr>
  </w:body>
</w:document>`
}

// --- Minimal Store-mode ZIP builder for browser/unit-test .docx creation ---
const CRC_TABLE = (() => {
  const table = new Uint32Array(256)
  for (let n = 0; n < 256; n++) {
    let c = n
    for (let k = 0; k < 8; k++) {
      c = (c & 1) ? (0xedb88320 ^ (c >>> 1)) : (c >>> 1)
    }
    table[n] = c >>> 0
  }
  return table
})()

const crc32 = (bytes: Uint8Array): number => {
  let c = 0xffffffff
  for (let i = 0; i < bytes.length; i++) {
    c = (CRC_TABLE[(c ^ bytes[i]!) & 0xff]! ^ (c >>> 8)) >>> 0
  }
  return (c ^ 0xffffffff) >>> 0
}

const buildStoreZip = (files: Array<{ name: string, data: Uint8Array }>): Uint8Array => {
  const encoder = new TextEncoder()
  const localChunks: Uint8Array[] = []
  const centralChunks: Uint8Array[] = []
  let offset = 0

  for (const file of files) {
    const nameBytes = encoder.encode(file.name)
    const data = file.data
    const crc = crc32(data)

    const localHeader = new Uint8Array(30 + nameBytes.length)
    const lv = new DataView(localHeader.buffer)
    lv.setUint32(0, 0x04034b50, true)
    lv.setUint16(4, 20, true)
    lv.setUint16(6, 0x0800, true) // UTF-8 flag
    lv.setUint16(8, 0, true) // store method
    lv.setUint16(10, 0, true)
    lv.setUint16(12, 0, true)
    lv.setUint32(14, crc, true)
    lv.setUint32(18, data.length, true)
    lv.setUint32(22, data.length, true)
    lv.setUint16(26, nameBytes.length, true)
    lv.setUint16(28, 0, true)
    localHeader.set(nameBytes, 30)

    localChunks.push(localHeader, data)

    const centralHeader = new Uint8Array(46 + nameBytes.length)
    const cv = new DataView(centralHeader.buffer)
    cv.setUint32(0, 0x02014b50, true)
    cv.setUint16(4, 20, true)
    cv.setUint16(6, 20, true)
    cv.setUint16(8, 0x0800, true)
    cv.setUint16(10, 0, true)
    cv.setUint16(12, 0, true)
    cv.setUint16(14, 0, true)
    cv.setUint32(16, crc, true)
    cv.setUint32(20, data.length, true)
    cv.setUint32(24, data.length, true)
    cv.setUint16(28, nameBytes.length, true)
    cv.setUint16(30, 0, true)
    cv.setUint16(32, 0, true)
    cv.setUint16(34, 0, true)
    cv.setUint16(36, 0, true)
    cv.setUint32(38, 0, true)
    cv.setUint32(42, offset, true)
    centralHeader.set(nameBytes, 46)
    centralChunks.push(centralHeader)

    offset += localHeader.length + data.length
  }

  const centralSize = centralChunks.reduce((s, c) => s + c.length, 0)
  const eocd = new Uint8Array(22)
  const ev = new DataView(eocd.buffer)
  ev.setUint32(0, 0x06054b50, true)
  ev.setUint16(4, 0, true)
  ev.setUint16(6, 0, true)
  ev.setUint16(8, files.length, true)
  ev.setUint16(10, files.length, true)
  ev.setUint32(12, centralSize, true)
  ev.setUint32(16, offset, true)
  ev.setUint16(20, 0, true)

  const totalLen = offset + centralSize + eocd.length
  const out = new Uint8Array(totalLen)
  let pos = 0
  for (const chunk of localChunks) {
    out.set(chunk, pos)
    pos += chunk.length
  }
  for (const chunk of centralChunks) {
    out.set(chunk, pos)
    pos += chunk.length
  }
  out.set(eocd, pos)
  return out
}

export const generateExecutiveSummaryDocxBytes = (
  summary: ExecutiveSummary,
  context: ExecutiveSummaryDocxContext = {}
): Uint8Array => {
  const documentXml = buildExecutiveSummaryDocumentXml(summary, context)
  const encoder = new TextEncoder()

  const contentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`

  const rels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

  const docRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

  const styles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Arial" w:hAnsi="Arial" w:cs="Arial"/>
        <w:sz w:val="20"/>
        <w:color w:val="2A2A2A"/>
      </w:rPr>
    </w:rPrDefault>
  </w:docDefaults>
</w:styles>`

  return buildStoreZip([
    { name: '[Content_Types].xml', data: encoder.encode(contentTypes) },
    { name: '_rels/.rels', data: encoder.encode(rels) },
    { name: 'word/_rels/document.xml.rels', data: encoder.encode(docRels) },
    { name: 'word/styles.xml', data: encoder.encode(styles) },
    { name: 'word/document.xml', data: encoder.encode(documentXml) }
  ])
}

export const executiveSummaryDocxFilename = (
  summary: Pick<ExecutiveSummary, 'nomorDokumen' | 'quarter' | 'tahun'>,
  variant?: ExecutiveSummaryDocxVariant
): string => {
  const resolved = resolveExecutiveSummaryVariant(summary, variant)
  const cleanNum = String(summary.nomorDokumen || (resolved === 'compilation' ? `Q${summary.quarter || 1}_${summary.tahun || 2026}` : 'LHA'))
    .replace(/[/\\:*?"<>|\s]+/g, '_')
  return resolved === 'compilation'
    ? `Executive_Summary_Compilation_${cleanNum}.docx`
    : `Executive_Summary_Individual_${cleanNum}.docx`
}
