import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { Header } from "@/components/header";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Awesome Go Ranking",
  description: "Discover and explore the most popular Go repositories from awesome-go, ranked by stars, forks, and recent activity.",
  keywords: ["go", "golang", "awesome-go", "repositories", "ranking", "stars", "open-source"],
  authors: [{ name: "awesome-go-rank" }],
  openGraph: {
    title: "Awesome Go Ranking",
    description: "Discover the most popular Go repositories",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <div className="min-h-screen flex flex-col">
            <Header />
            <main className="flex-1">
              {children}
            </main>
            <footer className="border-t py-6 mt-12">
              <div className="container mx-auto px-4 text-center text-sm text-muted-foreground">
                <p>
                  Data from{" "}
                  <a
                    href="https://github.com/avelino/awesome-go"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline hover:text-foreground"
                  >
                    awesome-go
                  </a>
                  {" "}• Updated daily
                </p>
              </div>
            </footer>
          </div>
        </ThemeProvider>
      </body>
    </html>
  );
}
