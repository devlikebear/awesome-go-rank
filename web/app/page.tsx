import Link from "next/link"
import { Package, TrendingUp } from "lucide-react"
import { RepoExplorer } from "@/components/repo-explorer"
import { getRepoData, selectInitialData, slugifySection } from "@/lib/data"

export const dynamic = "force-static"

const INITIAL_REPO_LIMIT = 50

export default async function HomePage() {
  const data = await getRepoData()
  const initialData = selectInitialData(data, INITIAL_REPO_LIMIT)

  return (
    <div className="container mx-auto px-4 py-8">
      <section className="text-center mb-12">
        <h1 className="text-4xl md:text-5xl font-bold mb-4 bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent">
          Discover Amazing Go Projects
        </h1>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Explore {data.totalRepos.toLocaleString()} curated Go repositories from{" "}
          <a
            href={data.metadata.sourceUrl}
            className="underline hover:text-foreground"
          >
            awesome-go
          </a>
        </p>
      </section>

      <section className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-10" aria-label="Ranking statistics">
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
      </section>

      <nav className="mb-10" aria-labelledby="category-heading">
        <h2 id="category-heading" className="text-2xl font-bold mb-4">Browse Go library categories</h2>
        <div className="flex flex-wrap gap-2">
          {data.sections.map((section) => (
            <Link
              key={section.name}
              href={`/sections/${slugifySection(section.name)}`}
              className="rounded-full border bg-card px-3 py-1.5 text-sm hover:border-primary/50 hover:text-primary"
            >
              {section.name} ({section.repoCount})
            </Link>
          ))}
        </div>
      </nav>

      <RepoExplorer initialData={initialData} />
    </div>
  )
}
