"use client"

import { ArrowUpDown } from "lucide-react"

export type SortOption = "stars" | "forks" | "updated"

interface SortControlsProps {
  sortBy: SortOption
  onSortChange: (sort: SortOption) => void
}

export function SortControls({ sortBy, onSortChange }: SortControlsProps) {
  const options: { value: SortOption; label: string }[] = [
    { value: "stars", label: "Stars" },
    { value: "forks", label: "Forks" },
    { value: "updated", label: "Recently Updated" },
  ]

  return (
    <div className="flex items-center gap-2">
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
    </div>
  )
}
