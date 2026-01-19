"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, MeResponse } from "@/lib/api";

function getErrMsg(e: unknown) {
  return e instanceof Error ? e.message : "Errore nel recupero utente";
}

export default function DashboardPage() {
  const [me, setMe] = useState<MeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      if (!cancelled) {
        setLoading(true);
        setError(null);
      }

      try {
        const data = await api.me();
        if (!cancelled) setMe(data);
      } catch (e: unknown) {
        if (!cancelled) setError(getErrMsg(e));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    run();

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="container py-4 py-md-5">
      <div className="row justify-content-center">
        <div className="col-12 col-md-10 col-lg-8">
          <div className="card">
            <div className="card-body">
              <div className="d-flex flex-column flex-sm-row justify-content-between align-items-sm-center gap-2 mb-3">
                <h1 className="h3 mb-0">Dashboard</h1>

                {/* CTA dashboard (responsive) */}
                <div className="d-grid d-sm-flex gap-2">
                  <Link className="btn btn-outline-primary" href="/dashboard/users">
                    Gestione utenti
                  </Link>
                  <Link className="btn btn-primary" href="/dashboard/inventory">
                    Gestione inventario
                  </Link>
                </div>
              </div>

              {error && (
                <div className="alert alert-danger mb-0" role="alert">
                  {error}
                </div>
              )}

              {!error && loading && (
                <div className="alert alert-info mb-0" role="alert">
                  Caricamento…
                </div>
              )}

              {!error && !loading && me && (
                <div className="alert alert-success mb-0" role="alert">
                  Benvenuto, <strong>{me.username}</strong> (id: {me.user_id})
                </div>
              )}

              {!error && !loading && !me && (
                <div className="alert alert-warning mb-0" role="alert">
                  Nessun utente disponibile.
                </div>
              )}

              <div className="mt-3">
                <small className="text-muted">
                  Se non sei autenticato, verrai reindirizzato al login quando provi ad accedere alle pagine protette.
                </small>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
