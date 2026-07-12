"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { Star } from "lucide-react"
import { RepoCard } from "@/components/repo-card"
import { SearchBar } from "@/components/search-bar"
import { FilterPanel } from "@/components/filter-panel"
import { SortControls, type SortOption, type SortOrder } from "@/components/sort-controls"
import type { Repo, RepoData } from "@/lib/data"

const TRENDING_MIN_STARS = 100
const PAGE_SIZE = 100

interface RepoExplorerProps {
  initialData: RepoData
}

export function RepoExplorer({ initialData }: RepoExplorerProps) {
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedSection, setSelectedSection] = useState("")
  const [minStars, setMinStars] = useState(0)
  const [sortBy, setSortBy] = useState<SortOption>("stars")
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc")
  const [data, setData] = useState(initialData)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [visibleCount, setVisibleCount] = useState(PAGE_SIZE)

  useEffect(() => {
    const controller = new AbortController()
    fetch("/data/repos.json", { signal: controller.signal })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`Repository data request failed (${response.status})`)
        }
        return response.json() as Promise<RepoData>
      })
      .then((repoData) => {
        if (!Array.isArray(repoData.sections)) {
          throw new Error("Repository data has an invalid sections field")
        }
        setData(repoData)
        setLoadError(null)
      })
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === "AbortError") return
        setLoadError(error instanceof Error ? error.message : "Repository data could not be loaded")
      })
    return () => controller.abort()
  }, [])

  const filteredRepos = useMemo(() => {
    let repos: Repo[]
    if (selectedSection) {
      repos = [...(data.sections.find((section) => section.name === selectedSection)?.repos ?? [])]
    } else {
      const seen = new Set<string>()
      repos = data.sections.flatMap((section) => section.repos).filter((repo) => {
        if (seen.has(repo.url)) return false
        seen.add(repo.url)
        return true
      })
    }

    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      repos = repos.filter((repo) =>
        repo.name.toLowerCase().includes(query) || repo.description?.toLowerCase().includes(query),
      )
    }
    repos = repos.filter((repo) => repo.stars >= minStars)
    if (sortBy === "trending") {
      repos = repos.filter((repo) => repo.stars >= TRENDING_MIN_STARS)
    }

    const direction = sortOrder === "desc" ? -1 : 1
    const numericSort = (a: number, b: number) => direction * (a - b)
    switch (sortBy) {
      case "stars":
        repos.sort((a, b) => numericSort(a.stars, b.stars))
        break
      case "forks":
        repos.sort((a, b) => numericSort(a.forks, b.forks))
        break
      case "updated":
        repos.sort((a, b) => numericSort(new Date(a.lastUpdated).getTime(), new Date(b.lastUpdated).getTime()))
        break
      case "trending":
        repos.sort((a, b) => compareTrending(a, b, direction))
        break
    }
    return repos
  }, [data, selectedSection, searchQuery, minStars, sortBy, sortOrder])

  useEffect(() => {
    setVisibleCount(PAGE_SIZE)
  }, [selectedSection, searchQuery, minStars, sortBy, sortOrder])

  const visibleRepos = filteredRepos.slice(0, visibleCount)

  const handleSearch = useCallback((query: string) => setSearchQuery(query), [])

  return (
    <section aria-labelledby="repository-heading">
      {loadError && (
        <div role="alert" className="mb-6 rounded-lg border border-red-500/40 bg-red-500/10 p-4 text-sm">
          Full repository data failed to load: {loadError}. Showing the statically generated top repositories only.
        </div>
      )}
      <div className="grid grid-cols-1 md:grid-cols-[1fr_auto] gap-4 items-center mb-8">
        <SearchBar onSearch={handleSearch} />
        <div className="flex items-center gap-2 rounded-lg border bg-card px-4 py-3">
          <Star className="h-5 w-5 text-yellow-500" />
          <span className="text-sm text-muted-foreground">Results</span>
          <strong>{filteredRepos.length.toLocaleString()}</strong>
        </div>
      </div>

      <div className="flex flex-col lg:flex-row gap-8">
        <FilterPanel
          sections={data.sections}
          selectedSection={selectedSection}
          onSectionChange={setSelectedSection}
          minStars={minStars}
          onMinStarsChange={setMinStars}
        />
        <div className="flex-1 min-w-0">
          <div className="flex flex-col xl:flex-row xl:items-center justify-between gap-4 mb-6">
            <h2 id="repository-heading" className="text-2xl font-bold">
              {selectedSection || "Top repositories"}
            </h2>
            <SortControls
              sortBy={sortBy}
              sortOrder={sortOrder}
              onSortChange={setSortBy}
              onSortOrderChange={setSortOrder}
            />
          </div>
          {sortBy === "trending" && (
            <p className="mb-4 text-sm text-muted-foreground">
              Trending uses 30-day star growth and includes repositories with at least {TRENDING_MIN_STARS} stars.
            </p>
          )}
          {filteredRepos.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-muted-foreground text-lg">No repositories found matching your criteria.</p>
              <p className="text-sm text-muted-foreground mt-2">Try adjusting your filters or search query.</p>
            </div>
          ) : (
            <div className="space-y-4">
              {visibleRepos.map((repo, index) => (
                <RepoCard key={repo.url} repo={repo} rank={index + 1} />
              ))}
            </div>
          )}
          {visibleCount < filteredRepos.length && (
            <div className="mt-8 text-center">
              <button
                type="button"
                onClick={() => setVisibleCount((count) => count + PAGE_SIZE)}
                className="rounded-md bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
              >
                Load more repositories
              </button>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function compareTrending(a: Repo, b: Repo, direction: number) {
  if (a.starsDelta30d == null && b.starsDelta30d == null) return b.stars - a.stars
  if (a.starsDelta30d == null) return 1
  if (b.starsDelta30d == null) return -1
  const trendOrder = direction * (a.starsDelta30d - b.starsDelta30d)
  return trendOrder || b.stars - a.stars
}
