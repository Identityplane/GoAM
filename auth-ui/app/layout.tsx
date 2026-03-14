import type React from "react"
import type { Metadata } from "next"
import { GeistSans } from "geist/font/sans"
import { GeistMono } from "geist/font/mono"
import { DM_Sans, Inter } from "next/font/google"
import "./globals.css"

const dmSans = DM_Sans({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-dm-sans",
})

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-inter",
})

export const metadata: Metadata = {
  title: "IdentityPlane",
  description: "Secure Identity Management",
  generator: "v0.app",
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <head>
        <style>{`
html {
  font-family: ${inter.style.fontFamily}, ${GeistSans.style.fontFamily};
  --font-sans: ${inter.variable};
  --font-mono: ${GeistMono.variable};
  --font-dm-sans: ${dmSans.variable};
  --font-inter: ${inter.variable};
}
        `}</style>
      </head>
      <body className={`${inter.variable} ${dmSans.variable}`}>{children}</body>
    </html>
  )
}
