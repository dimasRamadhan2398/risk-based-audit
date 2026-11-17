/**
 * Client-Side Document Text Extractor
 * Extracts genuine document text from PDF, DOCX, XLSX, CSV, and text files
 * directly in the browser using standard Web Streams (DecompressionStream).
 */

function decodeHexPdfString(hex: string): string {
  let clean = hex.replace(/[^0-9A-Fa-f]/g, '')
  if (clean.length % 2 !== 0) clean += '0'
  const bytes: number[] = []
  for (let i = 0; i < clean.length; i += 2) {
    bytes.push(parseInt(clean.slice(i, i + 2), 16))
  }
  try {
    return new TextDecoder('utf-8').decode(new Uint8Array(bytes))
  } catch {
    return String.fromCharCode(...bytes)
  }
}

function decodePdfLiteralString(str: string): string {
  return str
    .replace(/\\([0-7]{1,3})/g, (_, oct) => String.fromCharCode(parseInt(oct, 8)))
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\r')
    .replace(/\\t/g, '\t')
    .replace(/\\\(/g, '(')
    .replace(/\\\)/g, ')')
    .replace(/\\\\/g, '\\')
    .trim()
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

  // 1. Search for stream ... endstream blocks and decompress Flate streams
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
      const streamBytes = bytes.subarray(sIdx, eIdx)
      const decompressed = await decompressZlibDeflate(streamBytes)
      const streamText = decompressed || new TextDecoder('latin1').decode(streamBytes)

      // Literal strings (...) Tj
      for (const m of streamText.matchAll(/\(([^()\\]*(?:\\.[^()\\]*)*)\)\s*Tj/g)) {
        const decoded = decodePdfLiteralString(m[1])
        if (decoded) extractedChunks.push(decoded)
      }
      // Array literal strings [...] TJ
      for (const m of streamText.matchAll(/\[(.*?)\]\s*TJ/gs)) {
        const parts = m[1].matchAll(/\(([^()\\]*(?:\\.[^()\\]*)*)\)/g)
        for (const p of parts) {
          const decoded = decodePdfLiteralString(p[1])
          if (decoded) extractedChunks.push(decoded)
        }
      }
      // Hex strings <...> Tj
      for (const m of streamText.matchAll(/<([0-9A-Fa-f\s]+)>\s*Tj/g)) {
        const decoded = decodeHexPdfString(m[1])
        if (decoded && decoded.trim()) extractedChunks.push(decoded.trim())
      }
      // Array hex strings [...] TJ
      for (const m of streamText.matchAll(/\[(.*?)\]\s*TJ/gs)) {
        const hexParts = m[1].matchAll(/<([0-9A-Fa-f\s]+)>/g)
        for (const hp of hexParts) {
          const decoded = decodeHexPdfString(hp[1])
          if (decoded && decoded.trim()) extractedChunks.push(decoded.trim())
        }
      }
    }
  }

  if (extractedChunks.length > 0) {
    return extractedChunks.join(' ').replace(/\s+/g, ' ').trim()
  }

  // 2. Fallback: Search for printable text in uncompressed portions
  const parenMatches = latin.matchAll(/\(([^()\\]{4,250})\)/g)
  for (const m of parenMatches) {
    const cand = decodePdfLiteralString(m[1])
    if (/[A-Za-z]{3,}/.test(cand) && !cand.startsWith('PDF-') && !cand.startsWith('Adobe') && !cand.startsWith('Font')) {
      extractedChunks.push(cand)
    }
  }

  return extractedChunks.join(' ').replace(/\s+/g, ' ').trim()
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
      const text = await file.text()
      return {
        text: text.trim(),
        method: 'Direct Client Text Reader',
        charCount: text.trim().length
      }
    } catch {
      // fallback to FileReader
    }
  }

  const arrayBuffer = await file.arrayBuffer()

  // 2. Word Document (DOCX)
  if (ext === 'DOCX') {
    try {
      const text = await extractDocxOrXlsxFromZip(arrayBuffer, 'docx')
      if (text && text.length > 10) {
        return {
          text,
          method: 'Client Word (DOCX) OOXML Parser',
          charCount: text.length
        }
      }
    } catch (err) {
      console.warn('DOCX extraction notice:', err)
    }
  }

  // 3. Excel Spreadsheet (XLSX)
  if (ext === 'XLSX') {
    try {
      const text = await extractDocxOrXlsxFromZip(arrayBuffer, 'xlsx')
      if (text && text.length > 10) {
        return {
          text,
          method: 'Client Excel (XLSX) Spreadsheet Parser',
          charCount: text.length
        }
      }
    } catch (err) {
      console.warn('XLSX extraction notice:', err)
    }
  }

  // 4. PDF Document
  if (ext === 'PDF') {
    try {
      const text = await extractPdfTextFromBuffer(arrayBuffer)
      if (text && text.length > 10) {
        return {
          text,
          method: 'Client PDF Stream & Text Engine',
          charCount: text.length
        }
      }
    } catch (err) {
      console.warn('PDF extraction notice:', err)
    }
  }

  // 5. General printable text fallback
  const bytes = new Uint8Array(arrayBuffer)
  let rawStr = ''
  const maxScan = Math.min(bytes.length, 150000)
  for (let i = 0; i < maxScan; i++) {
    const b = bytes[i]
    if ((b >= 32 && b <= 126) || b === 10 || b === 13) {
      rawStr += String.fromCharCode(b)
    } else {
      rawStr += ' '
    }
  }
  const words = rawStr.match(/[A-Za-z0-9_.,;:!?'"/\-()]{3,}/g) || []
  const text = words.slice(0, 500).join(' ').trim()

  return {
    text,
    method: 'Binary Stream Text Scanner',
    charCount: text.length
  }
}

