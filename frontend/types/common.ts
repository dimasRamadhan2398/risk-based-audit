// Common utility types

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  message?: string;
  errors?: Record<string, string[]>;
}

export interface SelectOption {
  label: string;
  value: string | number;
}

export interface DateRange {
  start: string;
  end: string;
}

export interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;
  width?: string;
}

export interface TablePagination {
  pageSize: number;
  currentIndex: number;
  totalPages: number;
}

export type Quarter = 1 | 2 | 3 | 4;
export type ImpactKey = `impact_q${Quarter}`;
export type PossibilityKey = `possibility_q${Quarter}`;
export type RiskLevelKey = `risk_level_q${Quarter}`;

export type ResidualFields = {
  [K in ImpactKey]: number;
} & {
  [K in PossibilityKey]: number;
} & {
  [K in RiskLevelKey]: string;
};

export const riskColor = {
  Low: "successLight" as const,
  "Low to Moderate": "successDark" as const,
  Moderate: "warningLight" as const,
  "Moderate to High": "warningDark" as const,
  High: "error" as const,
};

export const riskColorStyling = {
  Low: "bg-success-500/10 text-success-400 ring ring-inset ring-success-400/25 text-inverted" as const,
  "Low to Moderate":
    "bg-success-600 ring-success-800/50 text-inverted" as const,
  Moderate:
    "bg-warning-600 ring ring-inset ring-warning-400/50 text-inverted" as const,
  "Moderate to High":
    "bg-primary-400 ring ring-inset ring-primary-200/50 text-inverted" as const,
  High: "bg-error-500 text-inverted" as const,
};
