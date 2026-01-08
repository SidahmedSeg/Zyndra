import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Click to Deploy',
  description: 'No-code deployment platform',
  icons: {
    icon: '/favicon.svg',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}

