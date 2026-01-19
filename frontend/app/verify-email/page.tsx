"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

type Status = "idle" | "loading" | "success" | "error";

function getErrorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return "Errore durante la verifica email";
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export default function VerifyEmailPage() {
  const searchParams = useSearchParams();
  const token = useMemo(() => searchParams.get("token") || "", [searchParams]);

  const [status, setStatus] = useState<Status>("idle");
  const [message, setMessage] = useState<string>("");

  const tokenMissing = !token;

  useEffect(() => {
    async function run() {
      if (!token) {
        setStatus("error");
        setMessage("Token mancante. Apri il link ricevuto via email.");
        return;
      }

      try {
        setStatus("loading");
        setMessage("");

        const res = await fetch(`${API_BASE}/auth/verify-email`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ token }),
        });

        if (res.status === 204) {
          setStatus("success");
          setMessage("Email verificata correttamente! Ora puoi effettuare il login.");
          return;
        }

        // prova a leggere messaggio errore JSON
        let data: unknown = null;
        const contentType = res.headers.get("content-type") || "";
        const isJson = contentType.includes("application/json");
        if (isJson) {
          data = await res.json().catch(() => null);
        }

        let backendMsg = `HTTP ${res.status}`;
        if (isObject(data) && typeof data["error"] === "string") {
          backendMsg = data["error"] as string;
        }

        setStatus("error");
        setMessage(backendMsg);
      } catch (err) {
        setStatus("error");
        setMessage(getErrorMessage(err));
      }
    }

    run();
  }, [token]);

  return (
    <main className="container py-4 py-md-5">
      <div className="row justify-content-center">
        <div className="col-12 col-sm-10 col-md-8 col-lg-6">
          <h1 className="h3 mb-3 mb-md-4 text-center">Verifica Email</h1>

          {tokenMissing && (
            <div className="alert alert-danger" role="alert">
              Token mancante. Apri il link ricevuto via email.
            </div>
          )}

          {status === "loading" && (
            <div className="alert alert-info" role="alert">
              Verifica in corso...
            </div>
          )}

          {status === "success" && (
            <div className="alert alert-success" role="alert">
              {message}
            </div>
          )}

          {status === "error" && (
            <div className="alert alert-danger" role="alert">
              {message || "Verifica fallita. Il token potrebbe essere scaduto o non valido."}
            </div>
          )}

          <div className="card">
            <div className="card-body">
              <div className="d-grid d-sm-flex gap-2">
                <Link className="btn btn-primary" href="/login">
                  Vai al login
                </Link>
                <Link className="btn btn-outline-primary" href="/">
                  Home
                </Link>
              </div>

              <div className="mt-3">
                <small className="text-muted">
                  Se il link è scaduto, richiedi una nuova email di verifica (funzione che aggiungeremo nello step
                  “Recupera password / gestione email”).
                </small>
              </div>
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
