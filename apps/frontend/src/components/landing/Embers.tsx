"use client";

import { useEffect, useRef } from "react";

type Ember = {
  x: number;
  y: number;
  vx: number;
  vy: number;
  r: number;
  life: number;
  max: number;
  sway: number;
  phase: number;
  copper: boolean;
};

/**
 * Violet embers rising from the portrait side of the hero.
 * Canvas, so 40+ particles cost one layer instead of 40 DOM nodes.
 * Pauses when the hero is offscreen or the tab is hidden.
 * With reduced motion it paints one still frame instead of animating.
 */
export function Embers() {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = ref.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;

    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    let w = 0;
    let h = 0;
    let embers: Ember[] = [];
    let raf = 0;
    let visible = true;
    let last = performance.now();

    const count = () => (window.innerWidth < 768 ? 22 : 48);

    // Spawn mostly on the right (portrait) side, low in the hero
    const spawn = (anywhere = false): Ember => {
      const mobile = window.innerWidth < 1024;
      const minX = mobile ? 0.05 : 0.52;
      const max = 4 + Math.random() * 5;
      return {
        x: w * (minX + Math.random() * (0.98 - minX)),
        y: anywhere ? Math.random() * h : h * (0.55 + Math.random() * 0.45),
        vx: (Math.random() - 0.5) * 8,
        vy: -(14 + Math.random() * 30),
        r: 0.6 + Math.random() * 1.8,
        life: anywhere ? Math.random() * max : 0,
        max,
        sway: 6 + Math.random() * 14,
        phase: Math.random() * Math.PI * 2,
        copper: Math.random() < 0.22,
      };
    };

    const resize = () => {
      const rect = canvas.getBoundingClientRect();
      w = rect.width;
      h = rect.height;
      canvas.width = Math.round(w * dpr);
      canvas.height = Math.round(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      embers = Array.from({ length: count() }, () => spawn(true));
    };

    const draw = () => {
      ctx.clearRect(0, 0, w, h);
      ctx.globalCompositeOperation = "lighter";
      for (const e of embers) {
        const t = e.life / e.max;
        // Fade in fast, fade out slow, and dim near the headline on the left
        const fade = Math.min(1, t * 6) * (1 - t);
        const side = window.innerWidth < 1024 ? 1 : Math.min(1, Math.max(0, (e.x / w - 0.42) * 4));
        const a = fade * side;
        if (a <= 0.01) continue;
        const x = e.x + Math.sin(e.phase + e.life * 1.3) * e.sway;
        const glow = e.r * 5;
        const g = ctx.createRadialGradient(x, e.y, 0, x, e.y, glow);
        const core = e.copper ? "243,192,155" : "233,213,255";
        const halo = e.copper ? "232,168,124" : "167,139,250";
        g.addColorStop(0, `rgba(${core},${a})`);
        g.addColorStop(0.25, `rgba(${halo},${a * 0.55})`);
        g.addColorStop(1, `rgba(${halo},0)`);
        ctx.fillStyle = g;
        ctx.beginPath();
        ctx.arc(x, e.y, glow, 0, Math.PI * 2);
        ctx.fill();
      }
      ctx.globalCompositeOperation = "source-over";
    };

    const tick = (now: number) => {
      const dt = Math.min(0.05, (now - last) / 1000);
      last = now;
      for (let i = 0; i < embers.length; i++) {
        const e = embers[i];
        e.life += dt;
        e.x += e.vx * dt;
        e.y += e.vy * dt;
        if (e.life >= e.max || e.y < -20) embers[i] = spawn();
      }
      draw();
      raf = requestAnimationFrame(tick);
    };

    const start = () => {
      if (reduce || raf || !visible || document.hidden) return;
      last = performance.now();
      raf = requestAnimationFrame(tick);
    };
    const stop = () => {
      cancelAnimationFrame(raf);
      raf = 0;
    };

    resize();
    draw();
    start();

    const ro = new ResizeObserver(() => {
      resize();
      draw();
    });
    ro.observe(canvas);

    const io = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
      if (visible) start();
      else stop();
    });
    io.observe(canvas);

    const onVis = () => (document.hidden ? stop() : start());
    document.addEventListener("visibilitychange", onVis);

    return () => {
      stop();
      ro.disconnect();
      io.disconnect();
      document.removeEventListener("visibilitychange", onVis);
    };
  }, []);

  return <canvas ref={ref} aria-hidden className="pointer-events-none absolute inset-0 h-full w-full" />;
}
