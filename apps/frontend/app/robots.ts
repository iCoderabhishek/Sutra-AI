import type { MetadataRoute } from "next";
import { SITE } from "@/src/constants/landing";

// Open to search engines and AI crawlers alike; the dashboard is private.
export default function robots(): MetadataRoute.Robots {
  return {
    rules: [{ userAgent: "*", allow: "/", disallow: ["/dashboard"] }],
    sitemap: `${SITE.url}/sitemap.xml`,
    host: SITE.url,
  };
}
