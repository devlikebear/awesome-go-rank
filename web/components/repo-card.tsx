import { Star, GitFork, Clock } from "lucide-react"
import { formatNumber, formatDate } from "@/lib/utils"
import { Repo } from "@/lib/data"

interface RepoCardProps {
  repo: Repo
  rank?: number
}

export function RepoCard({ repo, rank }: RepoCardProps) {
  return (
    <div className="group relative rounded-lg border bg-card p-6 shadow-sm transition-all hover:shadow-md hover:border-primary/50">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-2">
            {rank && (
              <span className="flex items-center justify-center w-8 h-8 rounded-full bg-primary/10 text-primary font-bold text-sm">
                #{rank}
              </span>
            )}
            <h3 className="font-semibold text-lg truncate group-hover:text-primary transition-colors">
              <a
                href={repo.url}
                target="_blank"
                rel="noopener noreferrer"
                className="hover:underline"
              >
                {repo.name}
              </a>
            </h3>
          </div>

          <p className="text-sm text-muted-foreground line-clamp-2 mb-4">
            {repo.description || "No description available"}
          </p>

          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-1" title="Stars">
              <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
              <span className="font-medium">{formatNumber(repo.stars)}</span>
            </div>

            <div className="flex items-center gap-1" title="Forks">
              <GitFork className="h-4 w-4" />
              <span>{formatNumber(repo.forks)}</span>
            </div>

            <div className="flex items-center gap-1" title="Last Updated">
              <Clock className="h-4 w-4" />
              <span className="text-xs">{formatDate(repo.lastUpdated)}</span>
            </div>
          </div>
        </div>

        <a
          href={repo.url}
          target="_blank"
          rel="noopener noreferrer"
          className="shrink-0 px-4 py-2 text-sm font-medium rounded-md bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
        >
          View
        </a>
      </div>
    </div>
  )
}
