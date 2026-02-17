"use client"

import { useState, useMemo, useCallback } from "react"
import { RepoCard } from "@/components/repo-card"
import { SearchBar } from "@/components/search-bar"
import { FilterPanel } from "@/components/filter-panel"
import { SortControls, SortOption } from "@/components/sort-controls"
import { Repo, Section } from "@/lib/data"
import { TrendingUp, Package, Star } from "lucide-react"

// This would normally come from getRepoData() in a real implementation
// For now, using mock data that will be replaced when the JSON file is available
const mockData = {
  updatedAt: new Date().toISOString(),
  totalRepos: 0,
  totalSections: 0,
  sections: [] as Section[],
}

export default function HomePage() {
  const [searchQuery, setSearchQuery] = useState("")
  const [selectedSection, setSelectedSection] = useState("")
  const [minStars, setMinStars] = useState(0)
  const [sortBy, setSortBy] = useState<SortOption>("stars")

  // In production, this would use: const data = await getRepoData()
  const data = mockData

  // Filter and sort repositories
  const filteredRepos = useMemo(() => {
    let repos: Repo[] = []

    // Get repos from selected section or all sections
    if (selectedSection) {
      const section = data.sections.find((s) => s.name === selectedSection)
      if (section) {
        repos = [...section.repos]
      }
    } else {
      repos = data.sections.flatMap((s) => s.repos)
    }

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      repos = repos.filter(
        (repo) =>
          repo.name.toLowerCase().includes(query) ||
          repo.description.toLowerCase().includes(query)
      )
    }

    // Apply stars filter
    repos = repos.filter((repo) => repo.stars >= minStars)

    // Apply sorting
    switch (sortBy) {
      case "stars":
        repos.sort((a, b) => b.stars - a.stars)
        break
      case "forks":
        repos.sort((a, b) => b.forks - a.forks)
        break
      case "updated":
        repos.sort(
          (a, b) =>
            new Date(b.lastUpdated).getTime() -
            new Date(a.lastUpdated).getTime()
        )
        break
    }

    return repos
  }, [data, selectedSection, searchQuery, minStars, sortBy])

  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query)
  }, [])

  const handleSectionChange = useCallback((section: string) => {
    setSelectedSection(section)
  }, [])

  const handleMinStarsChange = useCallback((stars: number) => {
    setMinStars(stars)
  }, [])

  const handleSortChange = useCallback((sort: SortOption) => {
    setSortBy(sort)
  }, [])

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Hero Section */}
      <div className="text-center mb-12">
        <h1 className="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent">
          Discover Amazing Go Projects
        </h1>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Explore {data.totalRepos.toLocaleString()} curated Go repositories
          from{" "}
          <a
            href="https://github.com/avelino/awesome-go"
            target="_blank"
            rel="noopener noreferrer"
            className="underline hover:text-foreground"
          >
            awesome-go
          </a>
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="rounded-lg border bg-card p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-full bg-blue-500/10">
              <Package className="h-6 w-6 text-blue-500" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Total Repositories</p>
              <p className="text-2xl font-bold">{data.totalRepos.toLocaleString()}</p>
            </div>
          </div>
        </div>

        <div className="rounded-lg border bg-card p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-full bg-purple-500/10">
              <TrendingUp className="h-6 w-6 text-purple-500" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Categories</p>
              <p className="text-2xl font-bold">{data.totalSections}</p>
            </div>
          </div>
        </div>

        <div className="rounded-lg border bg-card p-6">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-full bg-yellow-500/10">
              <Star className="h-6 w-6 text-yellow-500" />
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Filtered Results</p>
              <p className="text-2xl font-bold">{filteredRepos.length.toLocaleString()}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Search Bar */}
      <div className="flex justify-center mb-8">
        <SearchBar onSearch={handleSearch} />
      </div>

      {/* Main Content */}
      <div className="flex flex-col lg:flex-row gap-8">
        {/* Filters Sidebar */}
        <FilterPanel
          sections={data.sections}
          selectedSection={selectedSection}
          onSectionChange={handleSectionChange}
          minStars={minStars}
          onMinStarsChange={handleMinStarsChange}
        />

        {/* Repository List */}
        <div className="flex-1">
          {/* Sort Controls */}
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold">
              {selectedSection || "All Repositories"}
            </h2>
            <SortControls sortBy={sortBy} onSortChange={handleSortChange} />
          </div>

          {/* Results */}
          {filteredRepos.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-muted-foreground text-lg">
                No repositories found matching your criteria.
              </p>
              <p className="text-sm text-muted-foreground mt-2">
                Try adjusting your filters or search query.
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {filteredRepos.map((repo, index) => (
                <RepoCard key={repo.url} repo={repo} rank={index + 1} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
