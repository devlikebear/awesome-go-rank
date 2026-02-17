"use client"

import { Section } from "@/lib/data"
import { ChevronDown } from "lucide-react"
import { useState } from "react"

interface FilterPanelProps {
  sections: Section[]
  selectedSection: string
  onSectionChange: (section: string) => void
  minStars: number
  onMinStarsChange: (stars: number) => void
}

export function FilterPanel({
  sections,
  selectedSection,
  onSectionChange,
  minStars,
  onMinStarsChange,
}: FilterPanelProps) {
  const [isOpen, setIsOpen] = useState(false)

  const starsOptions = [
    { label: "All", value: 0 },
    { label: "1K+ stars", value: 1000 },
    { label: "5K+ stars", value: 5000 },
    { label: "10K+ stars", value: 10000 },
  ]

  return (
    <div className="w-full lg:w-64 space-y-6">
      {/* Mobile Toggle */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="lg:hidden w-full flex items-center justify-between p-4 rounded-lg border bg-card"
      >
        <span className="font-semibold">Filters</span>
        <ChevronDown className={`h-5 w-5 transition-transform ${isOpen ? "rotate-180" : ""}`} />
      </button>

      {/* Filters */}
      <div className={`space-y-6 ${isOpen ? "block" : "hidden lg:block"}`}>
        {/* Stars Filter */}
        <div className="space-y-2">
          <h3 className="font-semibold text-sm">Minimum Stars</h3>
          <div className="space-y-1">
            {starsOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => onMinStarsChange(option.value)}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                  minStars === option.value
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-accent"
                }`}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        {/* Category Filter */}
        <div className="space-y-2">
          <h3 className="font-semibold text-sm">Category</h3>
          <div className="space-y-1 max-h-96 overflow-y-auto">
            <button
              onClick={() => onSectionChange("")}
              className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                selectedSection === ""
                  ? "bg-primary text-primary-foreground"
                  : "hover:bg-accent"
              }`}
            >
              All Categories ({sections.reduce((sum, s) => sum + s.repoCount, 0)})
            </button>
            {sections.map((section) => (
              <button
                key={section.name}
                onClick={() => onSectionChange(section.name)}
                className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                  selectedSection === section.name
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-accent"
                }`}
              >
                {section.name} ({section.repoCount})
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
