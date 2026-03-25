export type LocaleCode = "ru" | "en";

export type UiFormatter = (...args: any[]) => string;

// Temporary compatibility type for the migration period.
// Once all legacy dictionary-style access is removed, this can be narrowed.
export interface UiDictionary {
  [key: string]: any;
}
