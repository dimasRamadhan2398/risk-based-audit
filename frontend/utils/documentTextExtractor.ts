/**
 * Client-Side Document Text Extractor
 * Extracts genuine document text from PDF, DOCX, XLSX, CSV, and text files
 * directly in the browser using standard Web Streams (DecompressionStream).
 */

function decodeHexPdfString(hex: string, cmap?: Map<number, string>): string {
  let clean = hex.replace(/[^0-9A-Fa-f]/g, '')
  if (!clean) return ''
  if (clean.length % 2 !== 0) clean += '0'

  // 1. If ToUnicode CMap is available, try 2-byte and 1-byte glyph mapping first
  if (cmap && cmap.size > 0) {
    if (clean.length % 4 === 0) {
      let mapped2 = ''
      let hits2 = 0
      const total2 = clean.length / 4
      for (let i = 0; i < clean.length; i += 4) {
        const code = parseInt(clean.slice(i, i + 4), 16)
        const ch = cmap.get(code)
        if (ch !== undefined) {
          mapped2 += ch
          hits2++
        } else if (code >= 0x20 && code <= 0x7e) {
          mapped2 += String.fromCharCode(code)
        }
      }
      if (hits2 > 0 && hits2 >= total2 * 0.4) return mapped2
    }
    let mapped1 = ''
    let hits1 = 0
    const total1 = clean.length / 2
    for (let i = 0; i < clean.length; i += 2) {
      const code = parseInt(clean.slice(i, i + 2), 16)
      const ch = cmap.get(code)
      if (ch !== undefined) {
        mapped1 += ch
        hits1++
      } else if (code >= 0x20 && code <= 0x7e) {
        mapped1 += String.fromCharCode(code)
      }
    }
    if (hits1 > 0 && hits1 >= total1 * 0.4) return mapped1
  }

  const bytes: number[] = []
  for (let i = 0; i < clean.length; i += 2) {
    bytes.push(parseInt(clean.slice(i, i + 2), 16))
  }

  // 2. Check UTF-16BE BOM (FE FF) or UTF-16BE pattern (alternating 0x00 bytes)
  if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    let out = ''
    for (let i = 2; i + 1 < bytes.length; i += 2) {
      const code = (bytes[i] << 8) | bytes[i + 1]
      if (code >= 0x20 || code === 0x09 || code === 0x0a) out += String.fromCharCode(code)
    }
    return out
  }
  if (bytes.length >= 4 && bytes.length % 2 === 0) {
    let zeroEven = 0
    for (let i = 0; i < bytes.length; i += 2) {
      if (bytes[i] === 0) zeroEven++
    }
    if (zeroEven >= (bytes.length / 2) * 0.7) {
      let out = ''
      for (let i = 0; i + 1 < bytes.length; i += 2) {
        const code = (bytes[i] << 8) | bytes[i + 1]
        if (code >= 0x20 && code <= 0x024f) out += String.fromCharCode(code)
      }
      return out
    }
  }

  try {
    return new TextDecoder('utf-8', { fatal: false }).decode(new Uint8Array(bytes)).replace(/\0/g, '')
  } catch {
    return String.fromCharCode(...bytes.filter(b => b >= 32 && b <= 126))
  }
}

function decodePdfLiteralString(str: string, cmap?: Map<number, string>): string {
  const unescaped = str
    .replace(/\\([0-7]{1,3})/g, (_, oct) => String.fromCharCode(parseInt(oct, 8)))
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\r')
    .replace(/\\t/g, '\t')
    .replace(/\\\(/g, '(')
    .replace(/\\\)/g, ')')
    .replace(/\\\\/g, '\\')

  // Handle UTF-16BE BOM inside literal string (\xFE\xFF...)
  if (unescaped.length >= 2 && unescaped.charCodeAt(0) === 0xfe && unescaped.charCodeAt(1) === 0xff) {
    let out = ''
    for (let i = 2; i + 1 < unescaped.length; i += 2) {
      const code = (unescaped.charCodeAt(i) << 8) | unescaped.charCodeAt(i + 1)
      if (code >= 0x20) out += String.fromCharCode(code)
    }
    return out.trim()
  }

  if (cmap && cmap.size > 0) {
    let mapped = ''
    let hits = 0
    for (let i = 0; i < unescaped.length; i++) {
      const code = unescaped.charCodeAt(i)
      const ch = cmap.get(code)
      if (ch !== undefined) {
        mapped += ch
        hits++
      } else if (code >= 0x20 && code <= 0x7e) {
        mapped += unescaped[i]
      }
    }
    if (hits > 0 && hits >= unescaped.length * 0.4) {
      return mapped.trim()
    }
  }

  return unescaped.replace(/\0/g, '').trim()
}

const PDF_SYNTAX_NOISE = new Set([
  'obj', 'endobj', 'stream', 'endstream', 'xref', 'trailer', 'startxref',
  'flatedecode', 'fontdescriptor', 'basefont', 'tounicode', 'cidfont',
  'identity', 'mediabox', 'cropbox', 'resources', 'procset', 'extgstate',
  'bitspercomponent', 'colorspace', 'devicergb', 'devicegray', 'xobject',
  'subtype', 'parent', 'catalog', 'outlines', 'predictor', 'decodeparms'
])

/**
 * Validates that an extracted string is genuine human-readable language (Indonesian/English)
 * and NOT binary stream noise, font tables, or PDF syntax operators.
 */
export function isReadableHumanText(text: string): boolean {
  if (!text) return false
  const clean = text.replace(/\s+/g, ' ').trim()
  if (clean.length < 6) return false

  let alpha = 0
  let vowels = 0
  let digits = 0
  let spaces = 0
  let weird = 0

  for (let i = 0; i < clean.length; i++) {
    const code = clean.charCodeAt(i)
    const ch = clean[i]
    if ((code >= 65 && code <= 90) || (code >= 97 && code <= 122)) {
      alpha++
      if ('aeiouAEIOU'.includes(ch)) vowels++
    } else if (code >= 48 && code <= 57) {
      digits++
    } else if (ch === ' ') {
      spaces++
    } else if (!'.,;:-_/()[]\'"?!%&+@#Rp$€¥\n\r\t|'.includes(ch)) {
      weird++
    }
  }

  if (weird / clean.length > 0.12) return false
  if ((alpha + digits) / clean.length < 0.5) return false
  if (alpha >= 8) {
    const vowelRatio = vowels / alpha
    if (vowelRatio < 0.15 || vowelRatio > 0.82) return false
  }

  const words = clean.toLowerCase().match(/[a-z]{2,}/g) || []
  if (words.length === 0 && digits < 4) return false

  let pdfNoiseHits = 0
  for (const w of words) {
    if (PDF_SYNTAX_NOISE.has(w)) pdfNoiseHits++
  }
  if (words.length > 0 && pdfNoiseHits / words.length > 0.2) return false

  return true
}

function parseToUnicodeCMap(streamText: string, cmap: Map<number, string>) {
  // Parse beginbfchar ... endbfchar
  for (const block of streamText.matchAll(/beginbfchar([\s\S]*?)endbfchar/g)) {
    for (const m of block[1].matchAll(/<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>/g)) {
      const src = parseInt(m[1], 16)
      const dstHex = m[2]
      let dst = ''
      for (let i = 0; i + 3 < dstHex.length; i += 4) {
        dst += String.fromCharCode(parseInt(dstHex.slice(i, i + 4), 16))
      }
      if (!dst && dstHex.length >= 2) {
        dst = String.fromCharCode(parseInt(dstHex, 16))
      }
      if (!Number.isNaN(src) && dst) cmap.set(src, dst)
    }
  }
  // Parse beginbfrange ... endbfrange
  for (const block of streamText.matchAll(/beginbfrange([\s\S]*?)endbfrange/g)) {
    for (const m of block[1].matchAll(/<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>/g)) {
      const start = parseInt(m[1], 16)
      const end = parseInt(m[2], 16)
      let target = parseInt(m[3], 16)
      if (!Number.isNaN(start) && !Number.isNaN(end) && !Number.isNaN(target) && end >= start && end - start <= 512) {
        for (let c = start; c <= end; c++) {
          cmap.set(c, String.fromCharCode(target++))
        }
      }
    }
  }
}

async function decompressZlibDeflate(bytes: Uint8Array): Promise<string | null> {
  let raw = bytes
  // If standard zlib header (0x78), strip 2-byte header and 4-byte Adler32 checksum for raw deflate
  if (bytes[0] === 0x78 && (bytes[1] === 0x01 || bytes[1] === 0x9c || bytes[1] === 0xda || bytes[1] === 0x5e)) {
    raw = bytes.subarray(2, bytes.length - 4)
  }

  for (const fmt of ['deflate-raw', 'deflate'] as const) {
    try {
      if (typeof DecompressionStream === 'undefined') break
      const ds = new DecompressionStream(fmt)
      const writer = ds.writable.getWriter()
      writer.write(raw as any)
      writer.close()
      const reader = ds.readable.getReader()
      const chunks: Uint8Array[] = []
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        if (value) chunks.push(value)
      }
      const totalLen = chunks.reduce((a, b) => a + b.length, 0)
      const out = new Uint8Array(totalLen)
      let off = 0
      for (const c of chunks) {
        out.set(c, off)
        off += c.length
      }
      return new TextDecoder('latin1').decode(out)
    } catch {
      // try next format
    }
  }
  return null
}

async function extractDocxOrXlsxFromZip(arrayBuffer: ArrayBuffer, targetFile: string): Promise<string> {
  const bytes = new Uint8Array(arrayBuffer)
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
  let pos = 0
  const extractedPieces: string[] = []

  while (pos < bytes.length - 30) {
    if (view.getUint32(pos, true) === 0x04034b50) {
      const compression = view.getUint16(pos + 8, true)
      const compSize = view.getUint32(pos + 18, true)
      const nameLen = view.getUint16(pos + 26, true)
      const extraLen = view.getUint16(pos + 28, true)

      const fileNameBytes = bytes.subarray(pos + 30, pos + 30 + nameLen)
      const fileName = new TextDecoder('utf-8').decode(fileNameBytes)
      const dataStart = pos + 30 + nameLen + extraLen

      const shouldExtract = (targetFile === 'docx' && (fileName === 'word/document.xml' || fileName.startsWith('word/header') || fileName.startsWith('word/footer'))) ||
        (targetFile === 'xlsx' && (fileName === 'xl/sharedStrings.xml' || fileName.includes('sheet')))

      if (shouldExtract && dataStart + compSize <= bytes.length) {
        const fileData = bytes.subarray(dataStart, dataStart + compSize)
        let xmlStr = ''
        if (compression === 0) {
          xmlStr = new TextDecoder('utf-8').decode(fileData)
        } else if (compression === 8 && typeof DecompressionStream !== 'undefined') {
          try {
            const ds = new DecompressionStream('deflate-raw')
            const writer = ds.writable.getWriter()
            writer.write(fileData as any)
            writer.close()
            const reader = ds.readable.getReader()
            const chunks: Uint8Array[] = []
            while (true) {
              const { done, value } = await reader.read()
              if (done) break
              if (value) chunks.push(value)
            }
            const totalLen = chunks.reduce((a, b) => a + b.length, 0)
            const out = new Uint8Array(totalLen)
            let off = 0
            for (const c of chunks) {
              out.set(c, off)
              off += c.length
            }
            xmlStr = new TextDecoder('utf-8').decode(out)
          } catch {
            // decompression fallback
          }
        }

        if (xmlStr) {
          const text = xmlStr
            .replace(/<\/w:p>/g, '\n')
            .replace(/<\/w:tr>/g, '\n')
            .replace(/<\/w:tc>/g, ' | ')
            .replace(/<\/row>/g, '\n')
            .replace(/<\/c>/g, ' | ')
            .replace(/<[^>]+>/g, '')
            .replace(/&lt;/g, '<')
            .replace(/&gt;/g, '>')
            .replace(/&amp;/g, '&')
            .replace(/&quot;/g, '"')
            .replace(/&apos;/g, "'")
            .trim()
          if (text) extractedPieces.push(text)
        }
      }
      pos = dataStart + compSize
    } else {
      pos++
    }
  }

  return extractedPieces.join('\n').replace(/\n{3,}/g, '\n\n').trim()
}

async function extractPdfTextFromBuffer(arrayBuffer: ArrayBuffer): Promise<string> {
  const bytes = new Uint8Array(arrayBuffer)
  const latin = new TextDecoder('latin1').decode(bytes)
  const extractedChunks: string[] = []
  const decompressedStreams: string[] = []
  const cmap = new Map<number, string>()

  // 1. Decompress Flate streams and first collect any ToUnicode CMap mappings
  const streamRegex = /stream[\r\n]+([\s\S]*?)[\r\n]+endstream/g
  let match: RegExpExecArray | null
  while ((match = streamRegex.exec(latin)) !== null) {
    const sStart = latin.indexOf('stream', match.index)
    let sIdx = sStart + 6
    while (sIdx < bytes.length && (bytes[sIdx] === 10 || bytes[sIdx] === 13)) sIdx++
    const eStart = latin.indexOf('endstream', sIdx)
    let eIdx = eStart
    while (eIdx > sIdx && (bytes[eIdx - 1] === 10 || bytes[eIdx - 1] === 13)) eIdx--

    if (sIdx > -1 && eIdx > sIdx) {
      // Skip obvious binary image/font objects by inspecting object header before 'stream'
      const headerSlice = latin.slice(Math.max(0, sStart - 220), sStart)
      if (/\/Subtype\s*\/Image|\/FontFile[23]?|\/Type\s*\/XRef|\/Type\s*\/Metadata/i.test(headerSlice)) {
        continue
      }
      const streamBytes = bytes.subarray(sIdx, eIdx)
      const decompressed = await decompressZlibDeflate(streamBytes)
      const streamText = decompressed || new TextDecoder('latin1').decode(streamBytes)
      if (streamText.includes('beginbfchar') || streamText.includes('beginbfrange')) {
        parseToUnicodeCMap(streamText, cmap)
        continue
      }
      decompressedStreams.push(streamText)
    }
  }

  // 2. Extract text operators from content streams using decoded CMap / UTF-16BE / Latin
  for (const streamText of decompressedStreams) {
    if (!streamText.includes('Tj') && !streamText.includes('TJ') && !streamText.includes('BT')) {
      continue
    }
    const streamPieces: string[] = []

    // Literal strings (...) Tj
    for (const m of streamText.matchAll(/\(([^()\\]*(?:\\.[^()\\]*)*)\)\s*Tj/g)) {
      const decoded = decodePdfLiteralString(m[1], cmap)
      if (decoded) streamPieces.push(decoded)
    }
    // Array literal & hex strings [...] TJ
    for (const m of streamText.matchAll(/\[(.*?)\]\s*TJ/gs)) {
      const tokenMatches = m[1].matchAll(/\(([^()\\]*(?:\\.[^()\\]*)*)\)|<([0-9A-Fa-f\s]+)>/g)
      let linePart = ''
      for (const tk of tokenMatches) {
        if (tk[1] !== undefined) {
          linePart += decodePdfLiteralString(tk[1], cmap)
        } else if (tk[2] !== undefined) {
          linePart += decodeHexPdfString(tk[2], cmap)
        }
      }
      if (linePart.trim()) streamPieces.push(linePart.trim())
    }
    // Hex strings <...> Tj
    for (const m of streamText.matchAll(/<([0-9A-Fa-f\s]+)>\s*Tj/g)) {
      const decoded = decodeHexPdfString(m[1], cmap)
      if (decoded && decoded.trim()) streamPieces.push(decoded.trim())
    }

    const joinedStream = streamPieces.join(' ').replace(/\s+/g, ' ').trim()
    if (isReadableHumanText(joinedStream)) {
      extractedChunks.push(joinedStream)
    }
  }

  const finalResult = extractedChunks.join(' ').replace(/\s+/g, ' ').trim()
  return isReadableHumanText(finalResult) ? finalResult : ''
}

export async function extractTextFromClientFile(file: File): Promise<{
  text: string
  method: string
  charCount: number
}> {
  const ext = file.name.split('.').pop()?.toUpperCase() || ''

  // 1. Plain Text / CSV / JSON / Markdown
  if (['TXT', 'CSV', 'LOG', 'JSON', 'MD', 'XML', 'HTML'].includes(ext)) {
    try {
      const text = (await file.text()).trim()
      return {
        text,
        method: 'Direct Client Text Reader',
        charCount: text.length
      }
    } catch {
      // fallback
    }
  }

  const arrayBuffer = await file.arrayBuffer()

  // 2. Word Document (DOCX)
  if (ext === 'DOCX') {
    try {
      const text = await extractDocxOrXlsxFromZip(arrayBuffer, 'docx')
      if (text && isReadableHumanText(text)) {
        return {
          text,
          method: 'Client Word (DOCX) OOXML Parser',
          charCount: text.length
        }
      }
    } catch (err) {
      console.warn('DOCX extraction notice:', err)
    }
    return { text: '', method: 'Server OOXML Word Extractor', charCount: 0 }
  }

  // 3. Excel Spreadsheet (XLSX)
  if (ext === 'XLSX') {
    try {
      const text = await extractDocxOrXlsxFromZip(arrayBuffer, 'xlsx')
      if (text && text.length > 5) {
        return {
          text,
          method: 'Client Excel (XLSX) Spreadsheet Parser',
          charCount: text.length
        }
      }
    } catch (err) {
      console.warn('XLSX extraction notice:', err)
    }
    return { text: '', method: 'Server Spreadsheet Extractor', charCount: 0 }
  }

  // 4. PDF Document — return clean human-readable text or delegate to backend pypdf/OCR
  if (ext === 'PDF') {
    try {
      const text = await extractPdfTextFromBuffer(arrayBuffer)
      if (text && isReadableHumanText(text)) {
        return {
          text,
          method: 'Client PDF Stream & Text Engine',
          charCount: text.length
        }
      }
    } catch (err) {
      console.warn('PDF extraction notice:', err)
    }
    // Delegate to backend python-ai pypdf + OCR engine instead of scanning binary bytes
    return {
      text: '',
      method: 'PDF Text & OCR Engine',
      charCount: 0
    }
  }

  // 5. Image files — always delegate to backend OCR engine (never scan binary image bytes!)
  if (['PNG', 'JPG', 'JPEG', 'WEBP', 'TIFF', 'BMP'].includes(ext)) {
    return {
      text: '',
      method: 'Optical Character Recognition (OCR)',
      charCount: 0
    }
  }

  return {
    text: '',
    method: 'Server Document & OCR Extractor',
    charCount: 0
  }
}

