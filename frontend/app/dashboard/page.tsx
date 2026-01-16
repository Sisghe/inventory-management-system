"use client";

import { useEffect, useState } from "react";
import { api, MeResponse } from "@/lib/api";

export default function DashboardHome() {
  const [me, setMe] = useState<MeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.me()
      .then(setMe)
      .catch((e: unknown) => {
        const msg = e instanceof Error ? e.message : "Errore nel recupero utente";
        setError(msg);
      });
  }, []);

  return (
    <div className="card">
      <div className="card-body">
        <h1 className="h4 mb-3">Dashboard</h1>

        {error && (
          <div className="alert alert-danger" role="alert">
            {error}
          </div>
        )}

        {!error && !me && <p>Caricamento…</p>}

        {me && (
          <p className="mb-0">
            Benvenuto, <strong>{me.username}</strong> (id: {me.user_id})
          </p>
        )}
      </div>
    </div>
  );
}
