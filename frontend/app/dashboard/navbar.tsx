"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

export default function DashboardNavbar() {
  const router = useRouter();

  async function onLogout() {
    try {
      await api.logout();
    } finally {
      router.push("/login");
      router.refresh();
    }
  }

  return (
    <header className="it-header-wrapper">
      <div className="it-nav-wrapper">
        <div className="it-header-navbar-wrapper">
          <div className="container">
            <div className="d-flex align-items-center justify-content-between py-2">
              <Link className="navbar-brand mb-0" href="/dashboard">
                Inventory
              </Link>

              <div className="d-flex align-items-center gap-3">
                <Link className="nav-link" href="/dashboard/users">
                  Utenti
                </Link>
                <Link className="nav-link" href="/dashboard/inventory">
                  Inventario
                </Link>

                <button className="btn btn-outline-light" onClick={onLogout} type="button">
                  Logout
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
