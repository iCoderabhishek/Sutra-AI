"use client";

import Image from "next/image";
import { useEffect, useRef } from "react";

// Fixed spark layout so server and client markup match (no Math.random in render)
const SPARKS = [
  { left: "8%", top: "78%", t: "7s", d: "0s", dx: "18px" },
  { left: "18%", top: "62%", t: "9s", d: "1.2s", dx: "-12px" },
  { left: "27%", top: "88%", t: "6.5s", d: "2.4s", dx: "26px" },
  { left: "41%", top: "70%", t: "8s", d: "0.6s", dx: "-20px" },
  { left: "56%", top: "84%", t: "7.5s", d: "3.1s", dx: "14px" },
  { left: "67%", top: "58%", t: "10s", d: "1.8s", dx: "-8px" },
  { left: "78%", top: "76%", t: "6s", d: "4s", dx: "22px" },
  { left: "88%", top: "66%", t: "8.5s", d: "2.9s", dx: "-16px" },
  { left: "93%", top: "86%", t: "7s", d: "5s", dx: "10px" },
  { left: "35%", top: "52%", t: "11s", d: "3.6s", dx: "-24px" },
];

export function Portrait() {
  const ref = useRef<HTMLDivElement>(null);

  // Subtle pointer parallax, written straight to CSS vars to skip React renders
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)");
    const fine = window.matchMedia("(pointer: fine)");
    if (reduce.matches || !fine.matches) return;

    let frame = 0;
    const onMove = (e: PointerEvent) => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => {
        const x = e.clientX / window.innerWidth - 0.5;
        const y = e.clientY / window.innerHeight - 0.5;
        el.style.setProperty("--px", `${(x * -14).toFixed(2)}px`);
        el.style.setProperty("--py", `${(y * -10).toFixed(2)}px`);
      });
    };
    window.addEventListener("pointermove", onMove, { passive: true });
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener("pointermove", onMove);
    };
  }, []);

  return (
    <div ref={ref} className="relative h-full w-full">
      {/* Violet aura behind the frame */}
      <div
        aria-hidden
        className="breathe absolute -inset-[12%] -z-10 rounded-full bg-[radial-gradient(closest-side,rgba(139,92,246,0.45),rgba(139,92,246,0.08)_60%,transparent)] blur-2xl"
      />

      <div className="edge cut-lg relative h-full w-full">
        <div className="cut-lg relative h-full w-full overflow-hidden bg-[#0d0817]">
          <div
            className="absolute -inset-4 transition-transform duration-500 ease-out"
            style={{ transform: "translate3d(var(--px,0), var(--py,0), 0) scale(1.04)" }}
          >
            <Image
              src="/sutra.png"
              alt="Sūtra, a woman with glowing violet eyes wrapped in strands of violet light"
              fill
              preload
              sizes="(max-width: 1024px) 92vw, 44vw"
              className="object-cover object-[50%_30%]"
            />
          </div>

          {/* Fade the portrait into the page at the bottom and left edge */}
          <div
            aria-hidden
            className="absolute inset-0 bg-[linear-gradient(to_top,#07050d_0%,rgba(7,5,13,0.35)_28%,transparent_55%),linear-gradient(to_right,rgba(7,5,13,0.55),transparent_30%)]"
          />

          {SPARKS.map((s, i) => (
            <span
              key={i}
              aria-hidden
              className="spark"
              style={
                {
                  left: s.left,
                  top: s.top,
                  "--t": s.t,
                  "--d": s.d,
                  "--dx": s.dx,
                } as React.CSSProperties
              }
            />
          ))}

          {/* HUD corner ticks */}
          <span aria-hidden className="absolute left-5 top-5 h-4 w-4 border-l border-t border-white/40" />
          <span aria-hidden className="absolute bottom-5 right-5 h-4 w-4 border-b border-r border-white/40" />

          <div className="absolute right-6 top-6 font-mono text-[10px] uppercase tracking-[0.3em] text-white/55">
            Agent · 001
          </div>

          <div className="absolute bottom-6 left-6 right-6 flex items-end justify-between gap-4">
            <p className="font-display text-[11px] font-medium uppercase tracking-[0.32em] text-white/85">
              Sūtra
              <span className="mx-2 text-glow">·</span>
              The keeper of threads
            </p>
            <span className="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.25em] text-glow">
              <span className="h-1.5 w-1.5 rounded-full bg-glow shadow-[0_0_8px_2px_rgba(192,132,252,0.8)]" />
              Awake
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
