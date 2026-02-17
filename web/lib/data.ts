import { readFileSync } from 'fs'
import { join } from 'path'

export interface Repo {
  name: string
  url: string
  stars: number
  forks: number
  lastUpdated: string
  description: string
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
  try {
    // In production, this will be in public/data/repos.json
    const filePath = join(process.cwd(), '..', 'public', 'data', 'repos.json')
    const fileContents = readFileSync(filePath, 'utf8')
    return JSON.parse(fileContents)
  } catch (error) {
    // Return mock data for development
    console.warn('Could not load repos.json, using mock data')
    return getMockData()
  }
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
      version: "1.0.0"
    }
  }
}

export function getSectionByName(data: RepoData, name: string): Section | undefined {
  return data.sections.find(s => s.name === name)
}

export function searchRepos(data: RepoData, query: string): Repo[] {
  if (!query) return []

  const lowerQuery = query.toLowerCase()
  const results: Repo[] = []

  for (const section of data.sections) {
    for (const repo of section.repos) {
      if (
        repo.name.toLowerCase().includes(lowerQuery) ||
        repo.description.toLowerCase().includes(lowerQuery)
      ) {
        results.push(repo)
      }
    }
  }

  return results
}
