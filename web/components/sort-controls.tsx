"use client"

import { ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react"

export type SortOption = "stars" | "forks" | "updated"
export type SortOrder = "asc" | "desc"

interface SortControlsProps {
  sortBy: SortOption
  sortOrder: SortOrder
  onSortChange: (sort: SortOption) => void
  onSortOrderChange: (order: SortOrder) => void
}

export function SortControls({ sortBy, sortOrder, onSortChange, onSortOrderChange }: SortControlsProps) {
  const options: { value: SortOption; label: string }[] = [
    { value: "stars", label: "Stars" },
    { value: "forks", label: "Forks" },
    { value: "updated", label: "Recently Updated" },
  ]

  return (
    <div className="flex items-center gap-3">
      <ArrowUpDown className="h-4 w-4 text-muted-foreground" />
      <span className="text-sm text-muted-foreground">Sort by:</span>
      <div className="flex gap-2">
        {options.map((option) => (
          <button
            key={option.value}
            onClick={() => onSortChange(option.value)}
            className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              sortBy === option.value
                ? "bg-primary text-primary-foreground"
                : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
            }`}
          >
            {option.label}
          </button>
        ))}
      </div>
      <button
        onClick={() => onSortOrderChange(sortOrder === "desc" ? "asc" : "desc")}
        className="p-2 rounded-md bg-secondary hover:bg-secondary/80 transition-colors"
        title={sortOrder === "desc" ? "Descending (highest first)" : "Ascending (lowest first)"}
      >
        {sortOrder === "desc" ? (
          <ArrowDown className="h-4 w-4" />
        ) : (
          <ArrowUp className="h-4 w-4" />
        )}
      </button>
    </div>
  )
}
