import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Script from "next/script";
import "./globals.css";

/* eslint-disable @next/next/no-css-tags */

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Inventory Management System",
  description: "Erasmus project – Next.js + Bootstrap Italia",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="it">
      <head>
        <link rel="stylesheet" href="/vendor/bootstrap-italia/bootstrap-italia.min.css" />
      </head>

      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        {children}

        {/* JS Bootstrap Italia: necessario per navbar mobile, dropdown, modal, ecc. */}
        <Script
          src="/vendor/bootstrap-italia/bootstrap-italia.bundle.min.js"
          strategy="afterInteractive"
        />
      </body>
    </html>
  );
}
