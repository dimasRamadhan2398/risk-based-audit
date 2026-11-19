import io
import os
import re
import csv
import zlib
import base64
import zipfile
import xml.etree.ElementTree as ET
from typing import Dict, Any, Optional, List


def detect_file_type(file_name: Optional[str] = None, raw_bytes: Optional[bytes] = None) -> str:
    ext = ""
    if file_name:
        ext = os.path.splitext(file_name.strip().lower())[1]

    if ext == ".pdf":
        return "PDF"
    if ext in (".docx", ".doc"):
        return "DOCX" if ext == ".docx" else "DOC"
    if ext in (".xlsx", ".xls"):
        return "XLSX" if ext == ".xlsx" else "XLS"
    if ext == ".csv":
        return "CSV"
    if ext in (".png", ".jpg", ".jpeg", ".tiff", ".bmp"):
        return "IMAGE"
    if ext in (".txt", ".md", ".json", ".log"):
        return "TXT"

    if raw_bytes:
        if raw_bytes.startswith(b"%PDF"):
            return "PDF"
        if raw_bytes.startswith(b"PK\x03\x04"):
            try:
                with zipfile.ZipFile(io.BytesIO(raw_bytes)) as zf:
                    names = zf.namelist()
                    if any(n.startswith("word/") for n in names):
                        return "DOCX"
                    if any(n.startswith("xl/") for n in names):
                        return "XLSX"
            except Exception:
                pass
        if raw_bytes.startswith(b"\x89PNG") or raw_bytes.startswith(b"\xff\xd8\xff"):
            return "IMAGE"

    return "TXT"


def _extract_docx_native(raw_bytes: bytes) -> str:
    """Pure Python OOXML (.docx) paragraph & table text extractor."""
    paragraphs: List[str] = []
    try:
        with zipfile.ZipFile(io.BytesIO(raw_bytes)) as zf:
            xml_targets = [
                n for n in zf.namelist()
                if n == "word/document.xml" or n.startswith("word/header") or n.startswith("word/footer")
            ]
            for target in xml_targets:
                xml_content = zf.read(target)
                root = ET.fromstring(xml_content)
                # Iterate over paragraphs <w:p>
                for elem in root.iter():
                    if elem.tag.endswith("}p") or elem.tag == "w:p":
                        runs = []
                        for child in elem.iter():
                            if child.tag.endswith("}t") or child.tag == "w:t":
                                if child.text:
                                    runs.append(child.text)
                        line = "".join(runs).strip()
                        if line:
                            paragraphs.append(line)
    except Exception:
        pass
    return "\n".join(paragraphs).strip()


def _extract_xlsx_native(raw_bytes: bytes) -> str:
    """Pure Python OOXML (.xlsx) sharedStrings & worksheet cell extractor."""
    lines: List[str] = []
    try:
        with zipfile.ZipFile(io.BytesIO(raw_bytes)) as zf:
            shared_strings: List[str] = []
            if "xl/sharedStrings.xml" in zf.namelist():
                ss_xml = zf.read("xl/sharedStrings.xml")
                ss_root = ET.fromstring(ss_xml)
                for si in ss_root.iter():
                    if si.tag.endswith("}si") or si.tag == "si":
                        texts = [
                            t.text for t in si.iter()
                            if (t.tag.endswith("}t") or t.tag == "t") and t.text
                        ]
                        shared_strings.append("".join(texts).strip())

            sheet_files = sorted([n for n in zf.namelist() if n.startswith("xl/worksheets/sheet") and n.endswith(".xml")])
            for sf in sheet_files:
                sh_xml = zf.read(sf)
                sh_root = ET.fromstring(sh_xml)
                for row in sh_root.iter():
                    if row.tag.endswith("}row") or row.tag == "row":
                        row_vals: List[str] = []
                        for cell in row:
                            if not (cell.tag.endswith("}c") or cell.tag == "c"):
                                continue
                            cell_type = cell.attrib.get("t", "")
                            val_text = ""
                            for sub in cell.iter():
                                if (sub.tag.endswith("}v") or sub.tag == "v") and sub.text:
                                    if cell_type == "s":
                                        idx = int(sub.text)
                                        if 0 <= idx < len(shared_strings):
                                            val_text = shared_strings[idx]
                                    else:
                                        val_text = sub.text
                                elif (sub.tag.endswith("}t") or sub.tag == "t") and sub.text:
                                    val_text = sub.text
                            if val_text.strip():
                                row_vals.append(val_text.strip())
                        if row_vals:
                            lines.append(" | ".join(row_vals))
            if not lines and shared_strings:
                lines = [s for s in shared_strings if s]
    except Exception:
        pass
    return "\n".join(lines).strip()


PDF_SYNTAX_NOISE = {
    "obj", "endobj", "stream", "endstream", "xref", "trailer", "startxref",
    "flatedecode", "fontdescriptor", "basefont", "tounicode", "cidfont",
    "identity", "mediabox", "cropbox", "resources", "procset", "extgstate",
    "bitspercomponent", "colorspace", "devicergb", "devicegray", "xobject",
    "subtype", "parent", "catalog", "outlines", "predictor", "decodeparms",
    "ihdr", "idat", "iend", "jfif", "exif"
}


def _is_readable_human_text(text: str) -> bool:
    """
    Validates that extracted text is genuine human-readable language (Indonesian/English)
    and NOT binary stream bytes, PDF object operators, or font glyph noise.
    """
    if not text:
        return False
    clean = re.sub(r"\s+", " ", text).strip()
    if len(clean) < 6:
        return False

    alpha = 0
    vowels = 0
    digits = 0
    weird = 0
    allowed_punct = set(".,;:-_/()[]'\"?!%&+@#Rp$€¥\n\r\t| ")

    for ch in clean:
        o = ord(ch)
        if (65 <= o <= 90) or (97 <= o <= 122):
            alpha += 1
            if ch in "aeiouAEIOU":
                vowels += 1
        elif 48 <= o <= 57:
            digits += 1
        elif ch not in allowed_punct:
            weird += 1

    if weird / len(clean) > 0.12:
        return False
    if (alpha + digits) / len(clean) < 0.50:
        return False
    if alpha >= 8:
        vowel_ratio = vowels / alpha
        if vowel_ratio < 0.15 or vowel_ratio > 0.82:
            return False

    words = re.findall(r"[a-z]{2,}", clean.lower())
    if not words and digits < 4:
        return False

    if words:
        noise_hits = sum(1 for w in words if w in PDF_SYNTAX_NOISE)
        if noise_hits / len(words) > 0.20:
            return False

    return True


def _decode_pdf_literal_string(s: str, cmap: Optional[Dict[int, str]] = None) -> str:
    # Unescape octal sequences first
    def _oct_repl(m):
        try:
            return chr(int(m.group(1), 8))
        except Exception:
            return " "

    unescaped = re.sub(r"\\([0-7]{1,3})", _oct_repl, s)
    unescaped = (
        unescaped.replace(r"\(", "(")
        .replace(r"\)", ")")
        .replace(r"\n", "\n")
        .replace(r"\r", "\r")
        .replace(r"\t", "\t")
        .replace(r"\\", "\\")
    )

    if len(unescaped) >= 2 and ord(unescaped[0]) == 0xFE and ord(unescaped[1]) == 0xFF:
        out = []
        for i in range(2, len(unescaped) - 1, 2):
            code = (ord(unescaped[i]) << 8) | ord(unescaped[i + 1])
            if code >= 0x20:
                out.append(chr(code))
        return "".join(out).strip()

    if cmap:
        mapped = []
        hits = 0
        for ch in unescaped:
            code = ord(ch)
            if code in cmap:
                mapped.append(cmap[code])
                hits += 1
            elif 0x20 <= code <= 0x7E:
                mapped.append(ch)
        if hits > 0 and hits >= len(unescaped) * 0.4:
            return "".join(mapped).strip()

    return unescaped.replace("\x00", "").strip()


def _decode_pdf_hex_string(hex_str: str, cmap: Optional[Dict[int, str]] = None) -> str:
    clean = re.sub(r"[^0-9A-Fa-f]", "", hex_str)
    if not clean:
        return ""
    if len(clean) % 2 != 0:
        clean += "0"

    if cmap:
        if len(clean) % 4 == 0:
            mapped2 = []
            hits2 = 0
            total2 = len(clean) // 4
            for i in range(0, len(clean), 4):
                code = int(clean[i:i + 4], 16)
                if code in cmap:
                    mapped2.append(cmap[code])
                    hits2 += 1
                elif 0x20 <= code <= 0x7E:
                    mapped2.append(chr(code))
            if hits2 > 0 and hits2 >= total2 * 0.4:
                return "".join(mapped2)
        mapped1 = []
        hits1 = 0
        total1 = len(clean) // 2
        for i in range(0, len(clean), 2):
            code = int(clean[i:i + 2], 16)
            if code in cmap:
                mapped1.append(cmap[code])
                hits1 += 1
            elif 0x20 <= code <= 0x7E:
                mapped1.append(chr(code))
        if hits1 > 0 and hits1 >= total1 * 0.4:
            return "".join(mapped1)

    raw = bytes.fromhex(clean)
    if len(raw) >= 2 and raw[0] == 0xFE and raw[1] == 0xFF:
        return raw[2:].decode("utf-16-be", errors="ignore").replace("\x00", "")
    if len(raw) >= 4 and len(raw) % 2 == 0:
        zero_even = sum(1 for i in range(0, len(raw), 2) if raw[i] == 0)
        if zero_even >= (len(raw) // 2) * 0.7:
            return raw.decode("utf-16-be", errors="ignore").replace("\x00", "")
    return raw.decode("utf-8", errors="ignore").replace("\x00", "")


def _parse_tounicode_cmap(stream_text: str, cmap: Dict[int, str]) -> None:
    for block in re.finditer(r"beginbfchar(.*?)endbfchar", stream_text, re.DOTALL):
        for m in re.finditer(r"<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>", block.group(1)):
            try:
                src = int(m.group(1), 16)
                dst_hex = m.group(2)
                dst_bytes = bytes.fromhex(dst_hex.zfill((len(dst_hex) + 1) // 2 * 2))
                dst = dst_bytes.decode("utf-16-be", errors="ignore") if len(dst_bytes) >= 2 else chr(dst_bytes[0])
                if dst:
                    cmap[src] = dst
            except Exception:
                continue
    for block in re.finditer(r"beginbfrange(.*?)endbfrange", stream_text, re.DOTALL):
        for m in re.finditer(r"<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>", block.group(1)):
            try:
                start = int(m.group(1), 16)
                end = int(m.group(2), 16)
                target = int(m.group(3), 16)
                if end >= start and (end - start) <= 512:
                    for c in range(start, end + 1):
                        cmap[c] = chr(target)
                        target += 1
            except Exception:
                continue


def _extract_pdf_native_streams(raw_bytes: bytes) -> str:
    """
    Pure Python PDF text stream parser (supports uncompressed & FlateDecode streams,
    ToUnicode CMap, UTF-16BE hex strings, and Tj / TJ operators).
    """
    extracted_chunks: List[str] = []
    decompressed_streams: List[str] = []
    cmap: Dict[int, str] = {}

    stream_pattern = re.compile(rb"stream[\r\n]+(.*?)[\r\n]+endstream", re.DOTALL)
    for match in stream_pattern.finditer(raw_bytes):
        header_start = max(0, match.start() - 240)
        header_bytes = raw_bytes[header_start:match.start()]
        if re.search(rb"/Subtype\s*/Image|/FontFile[23]?|/Type\s*/XRef|/Type\s*/Metadata", header_bytes, re.IGNORECASE):
            continue

        stream_data = match.group(1)
        decompressed = None
        for wbits in (zlib.MAX_WBITS, -zlib.MAX_WBITS):
            try:
                decompressed = zlib.decompress(stream_data, wbits)
                break
            except Exception:
                continue
        target_bytes = decompressed if decompressed is not None else stream_data
        try:
            text_str = target_bytes.decode("latin-1", errors="ignore")
        except Exception:
            continue

        if "beginbfchar" in text_str or "beginbfrange" in text_str:
            _parse_tounicode_cmap(text_str, cmap)
            continue
        decompressed_streams.append(text_str)

    for text_str in decompressed_streams:
        if "Tj" not in text_str and "TJ" not in text_str and "BT" not in text_str:
            continue
        stream_pieces: List[str] = []

        for tj_match in re.finditer(r"\(([^()\\]*(?:\\.[^()\\]*)*)\)\s*Tj", text_str):
            piece = _decode_pdf_literal_string(tj_match.group(1), cmap)
            if piece:
                stream_pieces.append(piece)

        for hex_tj in re.finditer(r"<([0-9A-Fa-f\s]+)>\s*Tj", text_str):
            piece = _decode_pdf_hex_string(hex_tj.group(1), cmap).strip()
            if piece:
                stream_pieces.append(piece)

        for tj_arr in re.finditer(r"\[(.*?)\]\s*TJ", text_str, re.DOTALL):
            arr_body = tj_arr.group(1)
            parts = []
            for tk in re.finditer(r"\(([^()\\]*(?:\\.[^()\\]*)*)\)|<([0-9A-Fa-f\s]+)>", arr_body):
                if tk.group(1) is not None:
                    p_str = _decode_pdf_literal_string(tk.group(1), cmap)
                else:
                    p_str = _decode_pdf_hex_string(tk.group(2), cmap)
                if p_str:
                    parts.append(p_str)
            joined_part = "".join(parts).strip()
            if joined_part:
                stream_pieces.append(joined_part)

        joined_stream = " ".join(stream_pieces).strip()
        if _is_readable_human_text(joined_stream):
            extracted_chunks.append(joined_stream)

    final_str = " ".join(extracted_chunks).strip()
    return final_str if _is_readable_human_text(final_str) else ""


def _run_tesseract_on_pil_image(img) -> str:
    import pytesseract
    for lang in ("ind+eng", "eng", None):
        try:
            txt = pytesseract.image_to_string(img, lang=lang) if lang else pytesseract.image_to_string(img)
            if txt and _is_readable_human_text(txt):
                return txt.strip()
        except Exception:
            continue
    return ""


def _extract_ocr_from_image_or_pdf(raw_bytes: bytes, file_type: str) -> str:
    """
    Runs OCR using pytesseract on Image files or embedded images inside scanned PDFs.
    """
    try:
        from PIL import Image
        if file_type == "IMAGE":
            img = Image.open(io.BytesIO(raw_bytes))
            return _run_tesseract_on_pil_image(img)
        elif file_type == "PDF":
            ocr_pages: List[str] = []
            # 1. Extract embedded page images via pypdf (works for scanned PDFs without poppler)
            for pkg_name in ("pypdf", "PyPDF2"):
                try:
                    mod = __import__(pkg_name)
                    reader = getattr(mod, "PdfReader")(io.BytesIO(raw_bytes))
                    for page in reader.pages[:15]:
                        for img_file in getattr(page, "images", []):
                            img = Image.open(io.BytesIO(img_file.data))
                            pt = _run_tesseract_on_pil_image(img)
                            if pt:
                                ocr_pages.append(pt)
                    if ocr_pages:
                        return "\n".join(ocr_pages).strip()
                except Exception:
                    continue
    except Exception:
        pass
    return ""


def _extract_printable_strings(raw_bytes: bytes, allow_binary_scan: bool = False) -> str:
    """Extracts human-readable UTF-8 / UTF-16LE text segments from plain text or legacy .doc files."""
    try:
        decoded_utf8 = raw_bytes.decode("utf-8", errors="ignore")
        printable_ratio = sum(1 for c in decoded_utf8 if c.isprintable() or c in "\r\n\t") / max(len(decoded_utf8), 1)
        if printable_ratio > 0.85 and _is_readable_human_text(decoded_utf8):
            return decoded_utf8.strip()
    except Exception:
        pass

    if not allow_binary_scan:
        return ""

    matches = re.findall(rb"[A-Za-z0-9 ,.\-:;/()%'\"?!\n\r]{12,}", raw_bytes)
    cleaned = []
    for m in matches:
        s = m.decode("latin-1", errors="ignore").strip()
        if _is_readable_human_text(s) and not any(tok in s.lower() for tok in ["content_types", "xml", "rels", "endobj", "stream", "xref", "flatedecode"]):
            cleaned.append(s)
    result = " ".join(cleaned).strip()
    return result if _is_readable_human_text(result) else ""


def extract_text_from_document(
    file_content_base64: Optional[str] = None,
    raw_bytes: Optional[bytes] = None,
    file_name: Optional[str] = None,
    fallback_text: Optional[str] = None
) -> Dict[str, Any]:
    """
    Main entrypoint for Document Text Extraction & OCR.
    Supports PDF, Word (.docx/.doc), Excel (.xlsx/.xls/.csv), Image OCR, and Plain Text.
    Returns:
      {
        "text": str,
        "file_type": str,
        "extraction_method": str,
        "char_count": int
      }
    """
    if raw_bytes is None and file_content_base64:
        b64_clean = file_content_base64.strip()
        if "," in b64_clean and b64_clean.startswith("data:"):
            b64_clean = b64_clean.split(",", 1)[1]
        try:
            raw_bytes = base64.b64decode(b64_clean)
        except Exception:
            raw_bytes = b64_clean.encode("utf-8", errors="ignore")

    file_type = detect_file_type(file_name, raw_bytes)
    extracted_text = ""
    extraction_method = "Direct Text Input"

    if raw_bytes:
        if file_type == "PDF":
            extraction_method = "PDF Text & OCR Engine"
            # 1. Try pypdf / PyPDF2 first (without deprecated strict=False)
            for pkg_name in ("pypdf", "PyPDF2"):
                try:
                    mod = __import__(pkg_name)
                    reader = getattr(mod, "PdfReader")(io.BytesIO(raw_bytes))
                    pages_text = []
                    for p in reader.pages:
                        pt = p.extract_text()
                        if pt and pt.strip():
                            pages_text.append(pt.strip())
                    cand = "\n".join(pages_text).strip()
                    if _is_readable_human_text(cand):
                        extracted_text = cand
                        break
                except Exception:
                    continue

            # 2. Try PyMuPDF (fitz) if available
            if not _is_readable_human_text(extracted_text):
                try:
                    import fitz
                    doc = fitz.open(stream=raw_bytes, filetype="pdf")
                    pages_text = []
                    for p in doc:
                        pt = p.get_text()
                        if pt and pt.strip():
                            pages_text.append(pt.strip())
                    cand = "\n".join(pages_text).strip()
                    if _is_readable_human_text(cand):
                        extracted_text = cand
                except Exception:
                    pass

            # 3. Try native PDF stream decompressor (with ToUnicode CMap & UTF-16BE support)
            if not _is_readable_human_text(extracted_text):
                cand = _extract_pdf_native_streams(raw_bytes)
                if _is_readable_human_text(cand):
                    extracted_text = cand

            # 4. Try OCR if scanned PDF (extracts embedded page images and runs Tesseract OCR)
            if not _is_readable_human_text(extracted_text):
                ocr_res = _extract_ocr_from_image_or_pdf(raw_bytes, "PDF")
                if _is_readable_human_text(ocr_res):
                    extracted_text = ocr_res
                    extraction_method = "Optical Character Recognition (OCR)"
                else:
                    extracted_text = ""

        elif file_type == "DOCX":
            extraction_method = "OOXML Word Extractor"
            try:
                import docx
                doc = docx.Document(io.BytesIO(raw_bytes))
                parts = [p.text.strip() for p in doc.paragraphs if p.text.strip()]
                for table in doc.tables:
                    for row in table.rows:
                        row_cells = [c.text.strip() for c in row.cells if c.text.strip()]
                        if row_cells:
                            parts.append(" | ".join(row_cells))
                extracted_text = "\n".join(parts).strip()
            except Exception:
                pass

            if not extracted_text:
                extracted_text = _extract_docx_native(raw_bytes)

        elif file_type == "DOC":
            extraction_method = "Legacy Word Binary Extractor"
            extracted_text = _extract_printable_strings(raw_bytes, allow_binary_scan=True)

        elif file_type in ("XLSX", "XLS"):
            extraction_method = "Spreadsheet Table Extractor"
            if file_type == "XLSX":
                extracted_text = _extract_xlsx_native(raw_bytes)
            if not extracted_text:
                try:
                    import pandas as pd
                    sheets = pd.read_excel(io.BytesIO(raw_bytes), sheet_name=None)
                    sheet_texts = []
                    for sh_name, df in sheets.items():
                        df_clean = df.dropna(how="all").fillna("")
                        for _, r in df_clean.iterrows():
                            vals = [str(v).strip() for v in r.values if str(v).strip()]
                            if vals:
                                sheet_texts.append(" | ".join(vals))
                    extracted_text = "\n".join(sheet_texts).strip()
                except Exception:
                    extracted_text = _extract_printable_strings(raw_bytes, allow_binary_scan=True)

        elif file_type == "CSV":
            extraction_method = "CSV Structured Extractor"
            try:
                csv_str = raw_bytes.decode("utf-8", errors="ignore")
                reader = csv.reader(io.StringIO(csv_str))
                rows = [" | ".join([c.strip() for c in r if c.strip()]) for r in reader]
                extracted_text = "\n".join([r for r in rows if r]).strip()
            except Exception:
                extracted_text = _extract_printable_strings(raw_bytes, allow_binary_scan=False)

        elif file_type == "IMAGE":
            extraction_method = "Optical Character Recognition (OCR)"
            extracted_text = _extract_ocr_from_image_or_pdf(raw_bytes, "IMAGE")

        else:
            extraction_method = "Plain Text Extractor"
            extracted_text = _extract_printable_strings(raw_bytes, allow_binary_scan=False)

    # Validate fallback_text is genuine readable text before accepting it
    if not _is_readable_human_text(extracted_text):
        extracted_text = ""
    if not extracted_text and fallback_text and _is_readable_human_text(fallback_text):
        extracted_text = fallback_text.strip()

    if not extracted_text and file_name:
        clean_title = os.path.splitext(os.path.basename(file_name))[0].replace("_", " ").replace("-", " ")
        extracted_text = f"Dokumen audit {clean_title}."

    # Normalize whitespace
    extracted_text = re.sub(r"\s+", " ", extracted_text).strip()

    return {
        "text": extracted_text,
        "file_type": file_type,
        "extraction_method": extraction_method,
        "char_count": len(extracted_text)
    }

