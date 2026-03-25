import classicScenarioRu from "../../../locales/scenario/classic/ru.json";
import classicScenarioEn from "../../../locales/scenario/classic/en.json";
import devTestScenarioRu from "../../../locales/scenario/dev_test/ru.json";
import devTestScenarioEn from "../../../locales/scenario/dev_test/en.json";
import type { LocaleCode } from "./localeTypes";

type ScenarioDictionary = Record<string, string>;

const SCENARIO_TEXT: Record<LocaleCode, Record<string, ScenarioDictionary>> = {
  ru: {
    classic: classicScenarioRu as ScenarioDictionary,
    dev_test: devTestScenarioRu as ScenarioDictionary,
  },
  en: {
    classic: classicScenarioEn as ScenarioDictionary,
    dev_test: devTestScenarioEn as ScenarioDictionary,
  },
};

const formatVars = (template: string, vars?: Record<string, unknown>): string => {
  if (!vars) return template;
  return template.replace(/\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}/g, (_full, key: string) => {
    const value = vars[key];
    if (typeof value === "undefined") return `{{${key}}}`;
    return String(value);
  });
};

export const resolveScenarioText = (
  locale: LocaleCode,
  scenarioId: string | undefined,
  key: string | undefined,
  vars?: Record<string, unknown>,
  fallback = ""
): string => {
  const normalizedScenarioId = scenarioId === "classic" ? "classic" : "dev_test";
  const normalizedKey = String(key ?? "").trim();
  if (!normalizedKey) return fallback;
  const dictionary = SCENARIO_TEXT[locale]?.[normalizedScenarioId] ?? SCENARIO_TEXT.ru.classic;
  const raw = dictionary[normalizedKey];
  if (typeof raw !== "string" || !raw.trim()) return fallback;
  return formatVars(raw, vars);
};
