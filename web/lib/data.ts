import { readFileSync } from "node:fs"
import { join } from "node:path"

export interface Repo {
  name: string
  url: string
  stars: number
  forks: number
  lastUpdated: string
  description: string
  starsDelta7d?: number | null
  starsDelta30d?: number | null
  archived?: boolean
  isNew?: boolean
}

export interface Section {
  name: string
  description: string
  repoCount: number
  repos: Repo[]
}

export interface RepoData {
  updatedAt: string
  totalRepos: number
  totalSections: number
  sections: Section[]
  metadata: {
    sourceOwner: string
    sourceRepo: string
    sourceUrl: string
    generatedBy: string
    version: string
  }
}

export async function getRepoData(): Promise<RepoData> {
  const filePath = join(process.cwd(), "public", "data", "repos.json")
  try {
    const parsed = JSON.parse(readFileSync(filePath, "utf8")) as RepoData
    if (!Array.isArray(parsed.sections) || typeof parsed.totalRepos !== "number") {
      throw new Error("repos.json does not match the expected schema")
    }
    return parsed
  } catch (error) {
    if (process.env.NODE_ENV !== "production") {
      console.warn(`Could not load ${filePath}; using development-only mock data`, error)
      return getMockData()
    }
    const message = error instanceof Error ? error.message : String(error)
    throw new Error(`Failed to load production repository data at ${filePath}: ${message}`)
  }
}

export function selectInitialData(data: RepoData, limit: number): RepoData {
  const topURLs = new Set(
    uniqueRepos(data)
      .sort((a, b) => b.stars - a.stars)
      .slice(0, limit)
      .map((repo) => repo.url),
  )
  return {
    ...data,
    sections: data.sections.map((section) => ({
      ...section,
      repos: section.repos.filter((repo) => topURLs.has(repo.url)),
    })),
  }
}

export function uniqueRepos(data: RepoData): Repo[] {
  const seen = new Set<string>()
  return data.sections.flatMap((section) => section.repos).filter((repo) => {
    if (seen.has(repo.url)) return false
    seen.add(repo.url)
    return true
  })
}

export function slugifySection(name: string): string {
  return name
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/(^-|-$)/g, "")
}

export function getSectionBySlug(data: RepoData, slug: string): Section | undefined {
  return data.sections.find((section) => slugifySection(section.name) === slug)
}

export function getSectionByName(data: RepoData, name: string): Section | undefined {
  return data.sections.find((section) => section.name === name)
}

export function searchRepos(data: RepoData, query: string): Repo[] {
  if (!query) return []
  const lowerQuery = query.toLowerCase()
  return uniqueRepos(data).filter((repo) =>
    repo.name.toLowerCase().includes(lowerQuery) || repo.description.toLowerCase().includes(lowerQuery),
  )
}

function getMockData(): RepoData {
  return {
    updatedAt: new Date().toISOString(),
    totalRepos: 0,
    totalSections: 0,
    sections: [],
    metadata: {
      sourceOwner: "avelino",
      sourceRepo: "awesome-go",
      sourceUrl: "https://github.com/avelino/awesome-go",
      generatedBy: "awesome-go-rank",
      version: "1.0.0",
    },
  }
}
