import { ArrowUpRight } from "lucide-react";
import { Nav } from "@/src/components/landing/Nav";
import { Portrait } from "@/src/components/landing/Portrait";
import { InstallCard } from "@/src/components/landing/InstallCard";
import { InvokeButton } from "@/src/components/landing/InvokeButton";
import { Embers } from "@/src/components/landing/Embers";
import { Powers } from "@/src/components/landing/Powers";
import { LINKS, SITE } from "@/src/constants/landing";

const HEADLINE = ["WRITTEN", "ONCE.", "OBEYED", "ON TIME."];

// Structured data so search engines and AI agents can read what Sutra is
const JSON_LD = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: SITE.name,
  description: SITE.description,
  url: SITE.url,
  image: `${SITE.url}/opengraph-image.png`,
  applicationCategory: "DeveloperApplication",
  operatingSystem: "macOS, Linux, Windows",
  license: "https://opensource.org/licenses/MIT",
  codeRepository: LINKS.github,
  downloadUrl: LINKS.github,
  installUrl: `${SITE.url}/#install`,
  author: { "@type": "Person", name: SITE.author, url: LINKS.author, email: LINKS.email },
};

export default function Home() {
  return (
    <main id="top" className="relative flex-1 overflow-x-clip">
      <script
        type="application/ld+json"
        // Static constant, but escape "<" anyway so it can never close the script tag
        dangerouslySetInnerHTML={{ __html: JSON.stringify(JSON_LD).replace(/</g, "\\u003c") }}
      />
      {/* ---------- Section 1: hero ---------- */}
      <section aria-labelledby="hero-title" className="relative isolate min-h-svh">
        {/* Backdrop: violet bloom on the portrait side + grid fading in from the right */}
        <div
          aria-hidden
          className="absolute inset-0 -z-10 bg-[radial-gradient(60%_70%_at_78%_38%,rgba(139,92,246,0.22),transparent_70%),radial-gradient(40%_40%_at_10%_100%,rgba(232,168,124,0.06),transparent_70%)]"
        />
        <div aria-hidden className="grid-fade absolute inset-0 -z-10" />
        {/* Drifting violet smoke on the portrait side, kept away from the headline */}
        <div aria-hidden className="pointer-events-none absolute inset-0 -z-10 overflow-hidden">
          <div className="haze haze-a right-[-8%] top-[4%] h-[60%] w-[46%] bg-[radial-gradient(closest-side,rgba(139,92,246,0.55),transparent)]" />
          <div className="haze haze-b right-[18%] top-[36%] h-[50%] w-[34%] bg-[radial-gradient(closest-side,rgba(192,132,252,0.35),rgba(232,168,124,0.1)_60%,transparent)]" />
        </div>

        {/* Embers float over the portrait but stay under the headline and install strip (z-10) */}
        <div aria-hidden className="pointer-events-none absolute inset-0 z-[5]">
          <Embers />
        </div>

        <Nav />

        <div className="mx-auto grid w-full max-w-[1400px] grid-cols-1 gap-10 px-5 pb-10 pt-6 sm:px-8 lg:grid-cols-12 lg:gap-6 lg:px-12 lg:pt-10">
          {/* Side blurb */}
          <aside className="order-2 lg:order-none lg:col-span-2 lg:pt-6">
            <Sparkle />
            <p className="mt-5 max-w-[22ch] text-sm leading-relaxed text-white/75">
              You speak once. She never forgets the words.
            </p>
          </aside>

          {/* Headline + CTAs */}
          <div className="order-3 lg:order-none lg:col-span-6 lg:-mr-24 lg:pt-2 relative z-10">
            <p className="inline-flex items-center gap-3 border border-line bg-violet/10 px-3 py-1.5 font-mono text-[11px] uppercase tracking-[0.22em] text-white">
              <span className="h-1.5 w-1.5 bg-glow shadow-[0_0_8px_2px_rgba(192,132,252,0.8)]" aria-hidden />
              AI agent CLI
            </p>

            <h1
              id="hero-title"
              className="mt-5 font-display text-[clamp(2.6rem,6.4vw,5.6rem)] font-extrabold leading-[0.95] tracking-[0.01em]"
            >
              {HEADLINE.map((word, i) => (
                <span
                  key={word}
                  className={`rise block ${i === 2 ? "mt-[0.12em]" : ""}`}
                  style={{ animationDelay: `${120 + i * 110}ms` }}
                >
                  {/* Inner span carries the light sweep, cascading down the lines */}
                  <span className="sheen" style={{ animationDelay: `${1600 + i * 140}ms` }}>
                    {word}
                  </span>
                </span>
              ))}
            </h1>

            <p
              className="rise mt-7 max-w-[44ch] text-base leading-relaxed text-white/70 sm:text-lg"
              style={{ animationDelay: "650ms" }}
            >
              Write one prompt. Your agent carries it out on your schedule.
            </p>

            <div
              className="rise mt-9 flex flex-wrap items-center gap-6"
              style={{ animationDelay: "780ms" }}
            >
              <InvokeButton href="#install">Invoke her</InvokeButton>
              <a
                href={LINKS.github}
                target="_blank"
                rel="noopener noreferrer"
                className="group inline-flex items-center gap-2 font-mono text-[11px] uppercase tracking-[0.22em] text-white/80 transition-colors hover:text-white"
              >
                View on GitHub
                <ArrowUpRight
                  size={15}
                  aria-hidden
                  className="text-glow transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
                />
              </a>
            </div>
          </div>

          {/* Portrait */}
          <div className="order-1 h-[62svh] min-h-[420px] lg:order-none lg:col-span-4 lg:h-[min(74svh,720px)]">
            <Portrait />
          </div>
        </div>

        {/* Bottom bento strip, overlapping the hero like the reference's tiles */}
        <div className="mx-auto grid w-full max-w-[1400px] grid-cols-1 gap-3 px-5 pb-12 sm:px-8 lg:-mt-24 lg:grid-cols-12 lg:px-12">
          <div className="relative z-10 lg:col-span-7">
            <InstallCard />
          </div>

          <div className="relative z-10 lg:col-span-5">
            <div className="edge cut h-full">
              <div className="cut relative flex h-full flex-col justify-between overflow-hidden bg-[linear-gradient(140deg,#2a1748_0%,#150c26_55%,#0b0814_100%)] p-6">
                <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-white/60">
                  The oath
                </p>
                <div className="mt-6 flex items-end justify-between gap-4">
                  <p className="font-display text-[clamp(2.4rem,4.6vw,3.6rem)] font-medium leading-none text-transparent [-webkit-text-stroke:1.5px_#e8a87c]">
                    1<span className="mx-3 text-copper/70 [-webkit-text-stroke:0]">→</span>N
                  </p>
                  <p className="max-w-[16ch] text-right font-display text-xs font-bold uppercase leading-relaxed tracking-[0.16em] text-white">
                    One prompt.
                    <br />
                    Every run after.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ---------- Section 2: powers + footer ---------- */}
      <Powers />
    </main>
  );
}

/** Four-point twin sparkle, as in the reference's left column. */
function Sparkle() {
  return (
    <svg width="40" height="40" viewBox="0 0 40 40" fill="none" aria-hidden>
      <path
        d="M16 6c.8 7.2 2.8 9.2 10 10-7.2.8-9.2 2.8-10 10-.8-7.2-2.8-9.2-10-10 7.2-.8 9.2-2.8 10-10Z"
        fill="url(#sp)"
      />
      <path
        d="M31 2c.4 3.3 1.3 4.2 4.6 4.6-3.3.4-4.2 1.3-4.6 4.6-.4-3.3-1.3-4.2-4.6-4.6 3.3-.4 4.2-1.3 4.6-4.6Z"
        fill="#e9d5ff"
      />
      <defs>
        <linearGradient id="sp" x1="6" y1="6" x2="26" y2="26">
          <stop stopColor="#ffffff" />
          <stop offset="1" stopColor="#c084fc" />
        </linearGradient>
      </defs>
    </svg>
  );
}
