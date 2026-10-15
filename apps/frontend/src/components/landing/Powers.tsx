import { ArrowUpRight } from "lucide-react";
import { LINKS, POWERS } from "@/src/constants/landing";

export function Powers() {
  return (
    <section
      aria-labelledby="powers-title"
      className="relative mx-auto w-full max-w-[1400px] px-5 pb-10 pt-20 sm:px-8 lg:px-12"
    >
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-line pb-5">
        <h2
          id="powers-title"
          className="font-display text-xl font-bold uppercase tracking-[0.14em] text-white sm:text-2xl"
        >
          Her four powers
        </h2>
      </div>

      <ul className="grid grid-cols-1 border-l border-line sm:grid-cols-2 lg:grid-cols-4">
        {POWERS.map((p, i) => (
          <li
            key={p.name}
            className="group relative overflow-hidden border-b border-r border-line p-6 transition-colors hover:bg-violet/[0.06] sm:p-7"
          >
            {/* Devanagari watermark */}
            <span
              aria-hidden
              className="pointer-events-none absolute -right-2 -top-3 select-none text-7xl font-bold text-white/[0.04] transition-colors group-hover:text-violet/15"
            >
              {p.devanagari}
            </span>

            <span className="font-mono text-[11px] tracking-[0.2em] text-glow">
              0{i + 1}
            </span>
            <h3 className="mt-8 font-display text-lg font-bold tracking-[0.16em] text-white">
              {p.name}
            </h3>
            <p className="mt-1 font-mono text-[11px] uppercase tracking-[0.22em] text-muted">
              {p.ability}
            </p>
            <p className="mt-5 max-w-[26ch] text-[15px] leading-relaxed text-white/75">
              {p.line}
            </p>

            {/* Bottom accent that charges up on hover */}
            <span
              aria-hidden
              className="absolute bottom-0 left-0 h-px w-0 bg-gradient-to-r from-glow to-transparent transition-[width] duration-500 group-hover:w-full"
            />
          </li>
        ))}
      </ul>

      <footer className="mt-16 grid gap-8 border-t border-line pt-8 sm:grid-cols-3 sm:items-end">
        <a
          href={LINKS.docs}
          target="_blank"
          rel="noopener noreferrer"
          className="group block"
        >
          <span className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">
            The scriptures
          </span>
          <span className="mt-2 flex items-center gap-2 font-display text-sm font-bold tracking-[0.12em] text-white">
            READ THE DOCS
            <ArrowUpRight
              size={16}
              aria-hidden
              className="text-glow transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
            />
          </span>
        </a>

        <a href={`mailto:${LINKS.email}`} className="group block">
          <span className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">
            Send word
          </span>
          <span className="mt-2 block font-display text-sm font-bold tracking-[0.08em] text-white underline decoration-glow/0 underline-offset-[6px] transition-colors group-hover:decoration-glow">
            {LINKS.email}
          </span>
        </a>

        <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted sm:text-right">
          <span className="block font-display text-sm font-bold tracking-[0.24em] text-white">
            SUTRA <span className="text-glow">AI</span>
            <a
              href={LINKS.author}
              target="_blank"
              rel="noopener noreferrer"
              className="ml-2 font-mono text-[10px] font-normal tracking-[0.3em] text-muted underline decoration-glow/0 underline-offset-4 transition-colors hover:text-white hover:decoration-glow"
            >
              by Abhishek
              <span className="sr-only"> (opens 0bhishek.com in a new tab)</span>
            </a>
          </span>
          <span className="mt-2 block">Sūtra · one thread, every run.</span>
        </p>
      </footer>
    </section>
  );
}
