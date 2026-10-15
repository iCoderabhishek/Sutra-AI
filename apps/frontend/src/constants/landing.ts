export const SITE = {
  url: "https://sutraai.0bhishek.com",
  name: "Sutra AI",
  title: "Sutra AI · The AI Agent CLI",
  description:
    "Sutra is an open-source AI agent CLI. Write one prompt in your terminal and it runs on a cron schedule, streams every step live, remembers the last run, and alerts you on Telegram when something changes.",
  author: "Abhishek Jha",
} as const;

export const LINKS = {
  github: "https://github.com/iCoderabhishek/Sutra-AI",
  docs: "https://github.com/iCoderabhishek/Sutra-AI#readme",
  email: "connect@0bhishek.com",
  author: "https://0bhishek.com",
} as const;

export type InstallTarget = {
  id: "unix" | "windows";
  label: string;
  prompt: string;
  command: string;
};

export const INSTALL: InstallTarget[] = [
  {
    id: "unix",
    label: "macOS / Linux",
    prompt: "$",
    command:
      "curl -fsSL https://raw.githubusercontent.com/iCoderabhishek/Sutra-AI/HEAD/scripts/install.sh | sh",
  },
  {
    id: "windows",
    label: "Windows",
    prompt: "PS>",
    command:
      "irm https://raw.githubusercontent.com/iCoderabhishek/Sutra-AI/HEAD/scripts/install.ps1 | iex",
  },
];

export type Power = {
  name: string;
  devanagari: string;
  meaning: string;
  /** What the power actually is, in plain agent terms */
  ability: string;
  line: string;
};

export const POWERS: Power[] = [
  {
    name: "KAAL",
    devanagari: "काल",
    meaning: "Time",
    ability: "Scheduled runs",
    line: "She wakes on your hour.",
  },
  {
    name: "DRISHTI",
    devanagari: "दृष्टि",
    meaning: "Sight",
    ability: "Live steps",
    line: "Every step she takes, you see.",
  },
  {
    name: "SMRITI",
    devanagari: "स्मृति",
    meaning: "Memory",
    ability: "Run memory",
    line: "She speaks only of what changed.",
  },
  {
    name: "SANKET",
    devanagari: "संकेत",
    meaning: "Signal",
    ability: "Telegram alerts",
    line: "Word reaches you when it shifts.",
  },
];
