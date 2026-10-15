import { ImageResponse } from "next/og";

export const size = { width: 180, height: 180 };
export const contentType = "image/png";

// Home-screen icon: the same thread mark as icon.svg, on the page's violet-black
export default function AppleIcon() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "radial-gradient(circle at 50% 30%, #2a1748 0%, #07050d 70%)",
        }}
      >
        <svg width="120" height="120" viewBox="0 0 32 32">
          <defs>
            <linearGradient id="g" x1="6" y1="8" x2="26" y2="26" gradientUnits="userSpaceOnUse">
              <stop stopColor="#ffffff" />
              <stop offset="1" stopColor="#a78bfa" />
            </linearGradient>
          </defs>
          <path
            d="M6 25c5.3 0 5.3-17.5 10-17.5S20.7 25 26 25"
            fill="none"
            stroke="url(#g)"
            strokeWidth="2.8"
            strokeLinecap="square"
          />
          <circle cx="16" cy="7.5" r="2.3" fill="#c084fc" />
        </svg>
      </div>
    ),
    size,
  );
}
