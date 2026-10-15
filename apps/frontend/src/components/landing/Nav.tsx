import { LINKS } from "@/src/constants/landing";

const linkCls =
  "font-mono text-[11px] uppercase tracking-[0.22em] text-white/70 transition-colors hover:text-white";

export function Nav() {
  return (
    <header className="relative z-20 mx-auto flex w-full max-w-[1400px] items-center justify-between px-5 py-6 sm:px-8 lg:px-12">
      <a href="#top" className="flex items-center gap-3" aria-label="Sutra AI by Abhishek, home">
        <ThreadMark />
        <span className="flex flex-col leading-none">
          <span className="font-display text-lg font-bold tracking-[0.28em] text-white">
            SUTRA <span className="text-glow">AI</span>
          </span>
          <span className="mt-1.5 font-mono text-[10px] uppercase tracking-[0.3em] text-muted">
            by Abhishek
          </span>
        </span>
      </a>

      <nav aria-label="Primary" className="hidden items-center gap-10 md:flex">
        <a className={linkCls} href={LINKS.docs} target="_blank" rel="noopener noreferrer">
          Docs
        </a>
        <a className={linkCls} href={LINKS.github} target="_blank" rel="noopener noreferrer">
          GitHub
        </a>
        <a className={linkCls} href="#install">
          Install
        </a>
      </nav>

      <div className="flex items-center gap-5">
        <span aria-hidden className="hidden h-6 w-px bg-line md:block" />
        <a
          href={`mailto:${LINKS.email}`}
          className="font-mono text-[11px] uppercase tracking-[0.22em] text-white underline decoration-glow/70 underline-offset-[6px] transition-colors hover:decoration-glow"
        >
          Contact us
        </a>
      </div>
    </header>
  );
}

/** A single thread folding back on itself, the sūtra glyph. */
function ThreadMark() {
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        d="M4 19c4 0 4-14 8-14s4 14 8 14"
        stroke="url(#tm)"
        strokeWidth="2"
        strokeLinecap="square"
      />
      <circle cx="12" cy="5" r="1.6" fill="#c084fc" />
      <defs>
        <linearGradient id="tm" x1="4" y1="5" x2="20" y2="19">
          <stop stopColor="#ffffff" />
          <stop offset="1" stopColor="#8b5cf6" />
        </linearGradient>
      </defs>
    </svg>
  );
}
