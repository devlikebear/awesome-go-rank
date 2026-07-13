import type { Metadata } from "next"
import Link from "next/link"
import { notFound } from "next/navigation"
import { RepoCard } from "@/components/repo-card"
import { getRepoData, getSectionBySlug, slugifySection } from "@/lib/data"

interface SectionPageProps {
  params: { name: string }
}

export async function generateStaticParams() {
  const data = await getRepoData()
  return data.sections.map((section) => ({ name: slugifySection(section.name) }))
}

export async function generateMetadata({ params }: SectionPageProps): Promise<Metadata> {
  const data = await getRepoData()
  const section = getSectionBySlug(data, params.name)
  if (!section) return { title: "Go library category not found" }
  return {
    title: `${section.name} Go Libraries Ranking`,
    description: `${section.description || `Compare ${section.name} libraries for Go.`} Ranked by stars, forks, activity, and growth.`,
  }
}

export default async function SectionPage({ params }: SectionPageProps) {
  const data = await getRepoData()
  const section = getSectionBySlug(data, params.name)
  if (!section) notFound()
  const repos = [...section.repos].sort((a, b) => b.stars - a.stars)

  return (
    <div className="container mx-auto px-4 py-8">
      <Link href="/" className="text-sm text-primary hover:underline">← All Go categories</Link>
      <header className="my-8">
        <h1 className="text-4xl font-bold mb-3">{section.name} Go Libraries Ranking</h1>
        <p className="text-lg text-muted-foreground max-w-3xl">
          {section.description || `Explore and compare ${section.name} libraries from awesome-go.`}
        </p>
        <p className="mt-3 text-sm text-muted-foreground">{section.repoCount.toLocaleString()} repositories, updated daily.</p>
      </header>
      <div className="space-y-4">
        {repos.map((repo, index) => <RepoCard key={repo.url} repo={repo} rank={index + 1} />)}
      </div>
    </div>
  )
}
