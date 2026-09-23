export enum QAType {
    REGULAR = 'Regular Self Assessment (RSA)',
    SAIV = 'Self Assessment w/ Independent Validation (SAIV)',
    QAR = 'Quality Assurance Review (QAR)',
    IACM = 'BUMN IACM Assessment'
}

export enum QAStatus {
    COMPLETED = 'Completed',
    VERIFIED = 'Verified',
    IN_PROGRESS = 'In Progress',
    PLANNED = 'Planned'
}

export interface QAReport {
    id: string
    type: QAType
    isImported?: boolean
    period: string
    reportName: string
    result: string
    status: QAStatus
    conductedBy?: string
    assessmentTitle: string
    validator?: string
    internalEvaluator?: string
    created_at?: string
    attachment?: {
        name: string
        size: string
        uploadedAt: string
        filePath?: string
    }
}