"use client";

import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function DashboardNavbar() {
  const router = useRouter();
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  // Chiude il menu quando cambi route (utile su mobile)
  useEffect(() => {
    setOpen(false);
  }, [pathname]);

  async function onLogout() {
    try {
      await api.logout();
    } finally {
      router.push("/login");
      router.refresh();
    }
  }

  const isUsers = pathname?.startsWith("/dashboard/users");
  const isInventory = pathname?.startsWith("/dashboard/inventory");

  return (
    <header className="it-header-wrapper">
      <div className="it-nav-wrapper">
        <div className="it-header-navbar-wrapper theme-dark">
          <div className="container-xxl">
            {/* bg-primary => contrasto garantito, link non “spariscono” */}
            <nav
              className="navbar navbar-expand-lg navbar-dark bg-primary py-2"
              aria-label="Navigazione dashboard"
            >
              <Link className="navbar-brand text-white" href="/dashboard">
                Inventory
              </Link>

              <button
                className="navbar-toggler border border-white ms-auto"
                type="button"
                aria-controls="dashboardNav"
                aria-expanded={open}
                aria-label="Mostra/Nascondi menu"
                onClick={() => setOpen((v) => !v)}
              >
                <span className="navbar-toggler-icon" />
              </button>

              {/* ✅ Toggle “hard” su mobile: evita link invisibili ma cliccabili */}
              <div
                id="dashboardNav"
                className={[
                  "navbar-collapse",
                  open ? "d-block" : "d-none",
                  "d-lg-flex",
                  "align-items-lg-center",
                  "ms-lg-auto",
                  "mt-3 mt-lg-0",
                ].join(" ")}
              >
                <ul className="navbar-nav ms-lg-auto mb-3 mb-lg-0">
                  <li className="nav-item">
                    <Link
                      className={`nav-link text-white ${
                        isUsers ? "active fw-semibold" : ""
                      }`}
                      href="/dashboard/users"
                      onClick={() => setOpen(false)}
                      aria-current={isUsers ? "page" : undefined}
                    >
                      Utenti
                    </Link>
                  </li>

                  <li className="nav-item">
                    <Link
                      className={`nav-link text-white ${
                        isInventory ? "active fw-semibold" : ""
                      }`}
                      href="/dashboard/inventory"
                      onClick={() => setOpen(false)}
                      aria-current={isInventory ? "page" : undefined}
                    >
                      Inventario
                    </Link>
                  </li>
                </ul>

                <div className="d-grid d-lg-block ms-lg-3">
                  <button
                    className="btn btn-outline-light"
                    onClick={() => {
                      setOpen(false);
                      void onLogout();
                    }}
                    type="button"
                  >
                    Logout
                  </button>
                </div>
              </div>
            </nav>
          </div>
        </div>
      </div>
    </header>
  );
}
