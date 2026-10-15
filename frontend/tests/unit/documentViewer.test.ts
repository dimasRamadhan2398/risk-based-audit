import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  escapeHtml,
  detectMimeType,
  openDocumentTab,
  renderDocumentInTab,
  viewDocumentInNewTab
} from '~/utils/documentViewer';

describe('documentViewer utility', () => {
  describe('escapeHtml', () => {
    it('escapes special HTML characters properly', () => {
      const input = '<script>alert("XSS" & \'test\')</script>';
      const expected = '&lt;script&gt;alert(&quot;XSS&quot; &amp; &#039;test&#039;)&lt;/script&gt;';
      expect(escapeHtml(input)).toBe(expected);
    });

    it('handles empty or null-like inputs gracefully', () => {
      expect(escapeHtml('')).toBe('');
      expect(escapeHtml(null as any)).toBe('');
      expect(escapeHtml(undefined as any)).toBe('');
    });
  });

  describe('detectMimeType', () => {
    it('detects PDF MIME type correctly', () => {
      expect(detectMimeType('document.pdf')).toBe('application/pdf');
      expect(detectMimeType('DOCUMENT.PDF')).toBe('application/pdf');
    });

    it('detects image MIME types correctly', () => {
      expect(detectMimeType('image.png')).toBe('image/png');
      expect(detectMimeType('photo.jpg')).toBe('image/jpeg');
      expect(detectMimeType('photo.jpeg')).toBe('image/jpeg');
    });

    it('detects text/plain correctly', () => {
      expect(detectMimeType('notes.txt')).toBe('text/plain');
    });

    it('respects explicitly provided responseType when not generic octet-stream', () => {
      expect(detectMimeType('unknown.custom', 'application/json')).toBe('application/json');
    });
  });

  describe('openDocumentTab and renderDocumentInTab', () => {
    let mockTab: any;
    let originalOpen: typeof window.open;

    beforeEach(() => {
      originalOpen = window.open;
      mockTab = {
        document: {
          title: '',
          body: {
            style: {},
            innerHTML: ''
          },
          open: vi.fn(),
          write: vi.fn(),
          close: vi.fn()
        },
        addEventListener: vi.fn(),
        close: vi.fn()
      };
      window.open = vi.fn().mockReturnValue(mockTab);
      window.URL.createObjectURL = vi.fn().mockReturnValue('blob:http://localhost:3000/mock-uuid');
      window.URL.revokeObjectURL = vi.fn();
    });

    afterEach(() => {
      window.open = originalOpen;
      vi.restoreAllMocks();
    });

    it('synchronously opens tab and sets loading title', () => {
      const tab = openDocumentTab('Annual Plan 2026.pdf');
      expect(window.open).toHaveBeenCalledWith('', '_blank');
      expect(tab?.document.title).toBe('Memuat Annual Plan 2026.pdf...');
    });

    it('renders document iframe and sets final tab title', () => {
      const blob = new Blob(['mock content'], { type: 'application/pdf' });
      renderDocumentInTab(mockTab, blob, 'Strategic Plan 2026.pdf', 'strategic_plan.pdf');

      expect(mockTab.document.open).toHaveBeenCalled();
      expect(mockTab.document.write).toHaveBeenCalledWith(expect.stringContaining('<title>Strategic Plan 2026.pdf</title>'));
      expect(mockTab.document.write).toHaveBeenCalledWith(expect.stringContaining('<iframe src="blob:http://localhost:3000/mock-uuid"></iframe>'));
      expect(mockTab.document.close).toHaveBeenCalled();
      expect(mockTab.document.title).toBe('Strategic Plan 2026.pdf');
    });

    it('viewDocumentInNewTab coordinates tab creation and rendering', async () => {
      const fetchFn = vi.fn().mockResolvedValue(new Blob(['data'], { type: 'application/pdf' }));
      await viewDocumentInNewTab(fetchFn, 'Working Paper 01.pdf');

      expect(window.open).toHaveBeenCalledWith('', '_blank');
      expect(fetchFn).toHaveBeenCalled();
      expect(mockTab.document.title).toBe('Working Paper 01.pdf');
    });

    it('closes tab if fetchFn fails', async () => {
      const fetchFn = vi.fn().mockRejectedValue(new Error('Network error'));
      await expect(viewDocumentInNewTab(fetchFn, 'ErrorDoc.pdf')).rejects.toThrow('Network error');

      expect(mockTab.close).toHaveBeenCalled();
    });
  });
});
