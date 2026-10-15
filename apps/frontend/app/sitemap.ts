import type { MetadataRoute } from "next";
import { SITE } from "@/src/constants/landing";

export default function sitemap(): MetadataRoute.Sitemap {
  return [
    {
      url: SITE.url,
      changeFrequency: "monthly",
      priority: 1,
      images: [`${SITE.url}/opengraph-image.png`],
    },
  ];
}
