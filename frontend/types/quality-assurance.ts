export enum QAType {
    REGULAR = 'Regular Self Assessment (RSA)',
    SAIV = 'Self Assessment w/ Independent Validation (SAIV)',
    QAR = 'Quality Assurance Review (QAR)'
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
    period: string
    reportName: string
    result: string
    status: QAStatus
    assessmentTitle: string
    validator?: string
    internalEvaluator?: string
    attachment?: {
        name: string
        size: string
        uploadedAt: string
    }
}