import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useUploadStrategicPlanDocumentStore } from '~/stores/upload-strategic-plan-document';
import { useUploadPlanDocumentStore } from '~/stores/upload-plan-document';

describe('Strategic Plan and Activity Plan Store Isolation', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.restoreAllMocks();
  });

  it('Strategic Plan store fetches from /uploaded-strategic-plan-documents', async () => {
    const mockStrategicDocs = [
      { id: 'strat-1', title: 'Strategic Plan 2026', fileName: 'strategic.pdf' }
    ];

    (globalThis as any).$fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/uploaded-strategic-plan-documents')) {
        return Promise.resolve(mockStrategicDocs);
      }
      return Promise.reject(new Error('Unknown url: ' + url));
    });

    const strategicStore = useUploadStrategicPlanDocumentStore();
    await strategicStore.fetchUploadedDocuments();

    expect((globalThis as any).$fetch).toHaveBeenCalledWith(
      expect.stringContaining('/uploaded-strategic-plan-documents'),
      expect.objectContaining({ method: 'GET' })
    );
    expect(strategicStore.uploadedDocuments).toEqual(mockStrategicDocs);
  });

  it('Activity Plan store fetches from /uploaded-plan-documents without mixing with Strategic Plan', async () => {
    const mockActivityDocs = [
      { id: 'act-1', title: 'Activity Plan 2026', fileName: 'activity.pdf' }
    ];

    (globalThis as any).$fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/uploaded-plan-documents')) {
        return Promise.resolve(mockActivityDocs);
      }
      return Promise.reject(new Error('Unknown url: ' + url));
    });

    const activityStore = useUploadPlanDocumentStore();
    await activityStore.fetchUploadedDocuments();

    expect((globalThis as any).$fetch).toHaveBeenCalledWith(
      expect.stringContaining('/uploaded-plan-documents'),
      expect.objectContaining({ method: 'GET' })
    );
    expect(activityStore.uploadedDocuments).toEqual(mockActivityDocs);
  });
});
