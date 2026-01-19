"use client";

import { useState } from "react";
import Link from "next/link";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

function isValidEmailLike(value: string): boolean {
  // validazione "verosimile" lato client (backend valida comunque)
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim());
}

function getErrorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return "Errore durante la richiesta di recupero password";
}

export default function ForgotPasswordPage() {
  const [username, setUsername] = useState("");
  const [loading, setLoading] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    const u = username.trim();
    if (!u) {
      setError("Inserisci la tua email.");
      return;
    }
    if (!isValidEmailLike(u)) {
      setError("Inserisci un indirizzo email valido (es. nome@dominio.it).");
      return;
    }

    try {
      setLoading(true);

      const res = await fetch(`${API_BASE}/auth/forgot-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ username: u }),
      });

      // backend: 204 se ok (anche se utente non esiste)
      if (!res.ok && res.status !== 204) {
        const contentType = res.headers.get("content-type") || "";
        if (contentType.includes("application/json")) {
          const data: unknown = await res.json().catch(() => null);
          if (
            data &&
            typeof data === "object" &&
            "error" in data &&
            typeof (data as Record<string, unknown>).error === "string"
          ) {
            throw new Error((data as Record<string, unknown>).error as string);
          }
        }
        throw new Error(`HTTP ${res.status}`);
      }

      setSuccess("Se l’email è registrata, riceverai a breve un messaggio con il link per reimpostare la password.");
      setUsername("");
    } catch (err: unknown) {
      setError(getErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="container py-4 py-md-5">
      <div className="row justify-content-center">
        <div className="col-12 col-sm-10 col-md-8 col-lg-6">
          <h1 className="h3 mb-3 mb-md-4 text-center">Recupera password</h1>

          {error && (
            <div className="alert alert-danger" role="alert">
              {error}
            </div>
          )}

          {success && (
            <div className="alert alert-success" role="alert">
              {success}
            </div>
          )}

          <div className="card">
            <div className="card-body">
              <p className="text-muted mb-3">
                Inserisci l’email con cui ti sei registrato. Se risulta presente a sistema, ti invieremo un link per
                reimpostare la password.
              </p>

              <form onSubmit={onSubmit}>
                <div className="mb-3">
                  <label className="form-label">Email (username)</label>
                  <input
                    className="form-control"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    autoComplete="email"
                    inputMode="email"
                    placeholder="nome@dominio.it"
                  />
                </div>

                <div className="d-grid d-sm-flex gap-2">
                  <button className="btn btn-primary" type="submit" disabled={loading}>
                    {loading ? "Invio..." : "Invia link di recupero"}
                  </button>

                  <Link className="btn btn-outline-primary" href="/login">
                    Torna al login
                  </Link>
                </div>
              </form>
            </div>
          </div>

          <div className="text-center mt-3">
            <Link className="link-primary" href="/">
              Home
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
