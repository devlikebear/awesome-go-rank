import type { MetadataRoute } from "next"
import { getRepoData, slugifySection } from "@/lib/data"

const siteURL = "https://awesome-go-rank.vercel.app"

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const data = await getRepoData()
  const lastModified = new Date(data.updatedAt)
  return [
    { url: siteURL, lastModified, changeFrequency: "daily", priority: 1 },
    ...data.sections.map((section) => ({
      url: `${siteURL}/sections/${slugifySection(section.name)}`,
      lastModified,
      changeFrequency: "daily" as const,
      priority: 0.8,
    })),
  ]
}
