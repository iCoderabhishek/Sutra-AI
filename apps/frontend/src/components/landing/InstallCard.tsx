"use client";

import {
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
  type KeyboardEvent,
} from "react";
import { Check, Copy } from "lucide-react";
import { INSTALL } from "@/src/constants/landing";

const noopSubscribe = () => () => {};

async function copyText(text: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  // Fallback for non-secure contexts
  const ta = document.createElement("textarea");
  ta.value = text;
  ta.setAttribute("readonly", "");
  ta.style.position = "absolute";
  ta.style.left = "-9999px";
  document.body.appendChild(ta);
  ta.select();
  document.execCommand("copy");
  document.body.removeChild(ta);
}

export function InstallCard() {
  // Default to the Windows tab for Windows visitors, without a hydration mismatch
  const detected = useSyncExternalStore(
    noopSubscribe,
    () => (/Win/i.test(navigator.userAgent) ? 1 : 0),
    () => 0,
  );
  const [chosen, setActive] = useState<number | null>(null);
  const active = chosen ?? detected;
  const [copied, setCopied] = useState(false);
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => () => {
    if (timer.current) clearTimeout(timer.current);
  }, []);

  const target = INSTALL[active];

  const onCopy = async () => {
    try {
      await copyText(target.command);
      setCopied(true);
      if (timer.current) clearTimeout(timer.current);
      timer.current = setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
    }
  };

  const onTabKey = (e: KeyboardEvent<HTMLButtonElement>) => {
    if (e.key !== "ArrowRight" && e.key !== "ArrowLeft") return;
    e.preventDefault();
    const next = (active + (e.key === "ArrowRight" ? 1 : -1) + INSTALL.length) % INSTALL.length;
    setActive(next);
    setCopied(false);
    tabRefs.current[next]?.focus();
  };

  return (
    <div id="install" className="edge cut h-full scroll-mt-24">
      <div className="cut flex h-full flex-col bg-[#0b0814]/95 p-5 sm:p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="font-display text-sm font-bold uppercase tracking-[0.18em] text-white">
            Invoke her
          </h2>

          <div role="tablist" aria-label="Operating system" className="flex border border-line">
            {INSTALL.map((t, i) => (
              <button
                key={t.id}
                ref={(el) => {
                  tabRefs.current[i] = el;
                }}
                role="tab"
                id={`tab-${t.id}`}
                aria-selected={i === active}
                aria-controls="install-panel"
                tabIndex={i === active ? 0 : -1}
                onClick={() => {
                  setActive(i);
                  setCopied(false);
                }}
                onKeyDown={onTabKey}
                className={`px-3 py-1.5 font-mono text-[11px] uppercase tracking-[0.15em] transition-colors ${
                  i === active ? "bg-white/10 text-white" : "text-muted hover:text-white"
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>
        </div>

        <div
          id="install-panel"
          role="tabpanel"
          aria-labelledby={`tab-${target.id}`}
          className="mt-4 flex items-stretch gap-0 border border-line bg-black/50"
        >
          <code className="flex min-w-0 flex-1 items-center gap-3 overflow-x-auto whitespace-nowrap px-4 py-3 font-mono text-[13px] text-[#e9e4f5] [scrollbar-width:thin]">
            <span className="select-none text-glow">{target.prompt}</span>
            <span>{target.command}</span>
          </code>
          <button
            type="button"
            onClick={onCopy}
            aria-label={copied ? "Copied install command" : "Copy install command"}
            className="flex shrink-0 items-center gap-2 border-l border-line px-4 font-mono text-[11px] uppercase tracking-[0.15em] text-white transition-colors hover:bg-violet/20"
          >
            {copied ? <Check size={14} aria-hidden /> : <Copy size={14} aria-hidden />}
            {copied ? "Copied" : "Copy"}
          </button>
        </div>

        <p className="mt-3 font-mono text-xs text-muted">
          Then whisper <span className="text-white">sutra</span>
          <span className="cursor ml-0.5 text-glow" aria-hidden>▍</span>
        </p>

        <span className="sr-only" aria-live="polite">
          {copied ? "Install command copied to clipboard" : ""}
        </span>
      </div>
    </div>
  );
}
