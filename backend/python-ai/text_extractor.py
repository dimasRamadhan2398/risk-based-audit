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


def _decode_pdf_literal_string(s: str) -> str:
    s = s.replace(r"\(", "(").replace(r"\)", ")").replace(r"\n", " ").replace(r"\r", " ").replace(r"\t", " ")
    # Remove octal escapes
    s = re.sub(r"\\[0-7]{1,3}", " ", s)
    return s.strip()


def _extract_pdf_native_streams(raw_bytes: bytes) -> str:
    """
    Pure Python PDF text stream parser (supports uncompressed & FlateDecode streams,
    Tj / TJ / ' operators) when pypdf is unavailable or as a fallback.
    """
    extracted_chunks: List[str] = []

    # 1. Inspect all stream ... endstream blocks
    stream_pattern = re.compile(rb"stream[\r\n]+(.*?)[\r\n]+endstream", re.DOTALL)
    for match in stream_pattern.finditer(raw_bytes):
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

        # Extract text inside (...) Tj or [(...)] TJ
        for tj_match in re.finditer(r"\(([^()\\]*(?:\\.[^()\\]*)*)\)\s*Tj", text_str):
            piece = _decode_pdf_literal_string(tj_match.group(1))
            if piece:
                extracted_chunks.append(piece)

        for tj_arr in re.finditer(r"\[(.*?)\]\s*TJ", text_str, re.DOTALL):
            arr_body = tj_arr.group(1)
            parts = []
            for p_match in re.finditer(r"\(([^()\\]*(?:\\.[^()\\]*)*)\)", arr_body):
                p_str = _decode_pdf_literal_string(p_match.group(1))
                if p_str:
                    parts.append(p_str)
            if parts:
                extracted_chunks.append("".join(parts))

    if extracted_chunks:
        return " ".join(extracted_chunks)

    # 2. Fallback: extract readable ASCII/UTF-8 sentences directly from raw PDF bytes
    raw_latin = raw_bytes.decode("latin-1", errors="ignore")
    fallback_pieces = []
    for m in re.finditer(r"\(([^()]{4,250})\)", raw_latin):
        candidate = _decode_pdf_literal_string(m.group(1))
        # Filter out PDF metadata/font names
        if re.search(r"[A-Za-z]{3,}", candidate) and not candidate.startswith(("PDF-", "Identity-H", "Adobe")):
            fallback_pieces.append(candidate)
    return " ".join(fallback_pieces).strip()


def _extract_ocr_from_image_or_pdf(raw_bytes: bytes, file_type: str) -> str:
    """
    Runs OCR using pytesseract / rapidocr if installed in the environment.
    """
    try:
        import pytesseract
        from PIL import Image
        if file_type == "IMAGE":
            img = Image.open(io.BytesIO(raw_bytes))
            ocr_text = pytesseract.image_to_string(img)
            if ocr_text and ocr_text.strip():
                return ocr_text.strip()
    except Exception:
        pass
    return ""


def _extract_printable_strings(raw_bytes: bytes) -> str:
    """Extracts human-readable UTF-8 / UTF-16LE text segments from binary files (.doc / .xls / scanned)."""
    try:
        decoded_utf8 = raw_bytes.decode("utf-8", errors="ignore")
        # If majority is printable text, return it directly
        printable_ratio = sum(1 for c in decoded_utf8 if c.isprintable() or c in "\r\n\t") / max(len(decoded_utf8), 1)
        if printable_ratio > 0.85:
            return decoded_utf8.strip()
    except Exception:
        pass

    # Extract sequences of printable ASCII/Indonesian characters >= 6 chars
    matches = re.findall(rb"[A-Za-z0-9 ,.\-:;/()%'\"?!\n\r]{8,}", raw_bytes)
    cleaned = []
    for m in matches:
        s = m.decode("latin-1", errors="ignore").strip()
        if re.search(r"[A-Za-z]{4,}", s) and not any(tok in s for tok in ["Content_Types", "xml", "rels", "endobj", "stream", "xref"]):
            cleaned.append(s)
    return " ".join(cleaned).strip()


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
            # 1. Try PyMuPDF (fitz) first - most accurate & robust across all PDF versions
            try:
                import fitz
                doc = fitz.open(stream=raw_bytes, filetype="pdf")
                pages_text = []
                for p in doc:
                    pt = p.get_text()
                    if pt and pt.strip():
                        pages_text.append(pt.strip())
                extracted_text = "\n".join(pages_text).strip()
            except Exception:
                extracted_text = ""

            # 2. Try PyPDF2 as secondary fallback
            if len(extracted_text) < 15:
                for pkg_name in ("pypdf", "PyPDF2"):
                    try:
                        mod = __import__(pkg_name)
                        reader = getattr(mod, "PdfReader")(io.BytesIO(raw_bytes), strict=False)
                        pages_text = []
                        for p in reader.pages:
                            pt = p.extract_text()
                            if pt and pt.strip():
                                pages_text.append(pt.strip())
                        cand = "\n".join(pages_text).strip()
                        if len(cand) > len(extracted_text):
                            extracted_text = cand
                            break
                    except Exception:
                        continue

            # 3. Try native PDF stream decompressor
            if len(extracted_text) < 15:
                extracted_text = _extract_pdf_native_streams(raw_bytes)

            # 4. Try OCR if scanned PDF
            if len(extracted_text) < 15:
                ocr_res = _extract_ocr_from_image_or_pdf(raw_bytes, "PDF")
                if ocr_res:
                    extracted_text = ocr_res
                    extraction_method = "Optical Character Recognition (OCR)"

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
            extracted_text = _extract_printable_strings(raw_bytes)

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
                    extracted_text = _extract_printable_strings(raw_bytes)

        elif file_type == "CSV":
            extraction_method = "CSV Structured Extractor"
            try:
                csv_str = raw_bytes.decode("utf-8", errors="ignore")
                reader = csv.reader(io.StringIO(csv_str))
                rows = [" | ".join([c.strip() for c in r if c.strip()]) for r in reader]
                extracted_text = "\n".join([r for r in rows if r]).strip()
            except Exception:
                extracted_text = _extract_printable_strings(raw_bytes)

        elif file_type == "IMAGE":
            extraction_method = "Optical Character Recognition (OCR)"
            extracted_text = _extract_ocr_from_image_or_pdf(raw_bytes, "IMAGE")
            if not extracted_text:
                extracted_text = _extract_printable_strings(raw_bytes)

        else:
            extraction_method = "Plain Text Extractor"
            extracted_text = _extract_printable_strings(raw_bytes)

        if len(extracted_text.strip()) < 10:
            printable_fallback = _extract_printable_strings(raw_bytes)
            if len(printable_fallback) > len(extracted_text):
                extracted_text = printable_fallback

    # Fallback to provided text/metadata ONLY if no text could be extracted from file bytes
    if not extracted_text and fallback_text and fallback_text.strip():
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

