"use client";

import { useEffect, useRef, useState } from "react";

type Burst = { id: number; sparks: { a: number; r: number; s: number; c: string }[] };

const COUNT = 12;
const LIFE_MS = 750;

/**
 * Metallic cut-corner CTA that throws a short spark burst on hover/focus.
 * The sparks live on a wrapper outside the clipped button so clip-path
 * doesn't cut them off, and they ignore the pointer so clicks are untouched.
 */
export function InvokeButton({ href, children }: { href: string; children: React.ReactNode }) {
  const [bursts, setBursts] = useState<Burst[]>([]);
  const nextId = useRef(0);
  const last = useRef(0);
  const timers = useRef<ReturnType<typeof setTimeout>[]>([]);

  useEffect(() => {
    const pending = timers.current;
    return () => pending.forEach(clearTimeout);
  }, []);

  const fire = () => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    // Throttle so jittery hovering doesn't stack bursts
    const now = performance.now();
    if (now - last.current < 600) return;
    last.current = now;

    const id = nextId.current++;
    const sparks = Array.from({ length: COUNT }, (_, i) => ({
      a: (360 / COUNT) * i + Math.random() * 24 - 12,
      r: 34 + Math.random() * 38,
      s: 1.5 + Math.random() * 1.5,
      c: Math.random() < 0.25 ? "#f3c09b" : "#e9d5ff",
    }));
    setBursts((b) => [...b, { id, sparks }]);
    timers.current.push(
      setTimeout(() => setBursts((b) => b.filter((x) => x.id !== id)), LIFE_MS),
    );
  };

  return (
    // Glow lives on the wrapper: a filter on the clipped button itself would be cut off
    <span className="cta relative inline-flex">
      <a
        href={href}
        onPointerEnter={fire}
        onFocus={fire}
        className="btn-metal cut relative inline-flex h-14 items-center px-10 font-display text-[13px] font-bold uppercase tracking-[0.2em]"
      >
        {children}
      </a>

      <span aria-hidden className="pointer-events-none absolute left-1/2 top-1/2">
        {bursts.map((b) =>
          b.sparks.map((s, i) => (
            <span
              key={`${b.id}-${i}`}
              className="burst"
              style={
                {
                  "--a": `${s.a}deg`,
                  "--r": `${s.r}px`,
                  width: `${s.s}px`,
                  height: `${s.s}px`,
                  background: s.c,
                } as React.CSSProperties
              }
            />
          )),
        )}
      </span>
    </span>
  );
}
