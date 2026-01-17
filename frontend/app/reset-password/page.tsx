"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useSearchParams, useRouter } from "next/navigation";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

type Status = "idle" | "loading" | "success" | "error";

function validatePasswordAgID(pw: string): string | null {
  if (pw.length < 8) return "La password deve essere lunga almeno 8 caratteri.";
  if (!/[A-Z]/.test(pw)) return "La password deve contenere almeno una lettera maiuscola.";
  if (!/[!@#$%^&*()_\-+=\[\]{};:'\",.<>/?\\|`~]/.test(pw))
    return "La password deve contenere almeno un carattere speciale.";
  return null;
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export default function ResetPasswordPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = useMemo(() => searchParams.get("token") || "", [searchParams]);

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");

  const [status, setStatus] = useState<Status>("idle");
  const [message, setMessage] = useState<string>("");

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setMessage("");

    const t = token.trim();
    if (!t) {
      setStatus("error");
      setMessage("Token mancante. Apri il link ricevuto via email.");
      return;
    }

    if (!password || !confirm) {
      setStatus("error");
      setMessage("Compila tutti i campi.");
      return;
    }

    if (password !== confirm) {
      setStatus("error");
      setMessage("Le password non coincidono.");
      return;
    }

    const pwErr = validatePasswordAgID(password);
    if (pwErr) {
      setStatus("error");
      setMessage(pwErr);
      return;
    }

    try {
      setStatus("loading");

      const res = await fetch(`${API_BASE}/auth/reset-password`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ token: t, password }),
      });

      if (res.status === 204) {
        setStatus("success");
        setMessage("Password aggiornata correttamente! Ora puoi effettuare il login.");
        // opzionale: dopo 1 secondo vai al login
        setTimeout(() => {
          router.push("/login");
          router.refresh();
        }, 1000);
        return;
      }

      // errore: prova a leggere JSON
      let data: unknown = null;
      const isJson = res.headers.get("content-type")?.includes("application/json");
      if (isJson) {
        data = await res.json().catch(() => null);
      }

      let backendMsg = `HTTP ${res.status}`;
      if (isObject(data) && typeof data["error"] === "string") {
        backendMsg = data["error"] as string;
      }

      setStatus("error");
      setMessage(backendMsg);
    } catch {
      setStatus("error");
      setMessage("Errore durante il reset password.");
    }
  }

  const tokenMissing = !token;

  return (
    <main className="container py-5" style={{ maxWidth: 620 }}>
      <h1 className="mb-4">Reimposta password</h1>

      {tokenMissing && (
        <div className="alert alert-danger" role="alert">
          Token mancante. Apri il link ricevuto via email.
        </div>
      )}

      {status === "loading" && (
        <div className="alert alert-info" role="alert">
          Salvataggio in corso...
        </div>
      )}

      {status === "success" && (
        <div className="alert alert-success" role="alert">
          {message}
        </div>
      )}

      {status === "error" && message && (
        <div className="alert alert-danger" role="alert">
          {message}
        </div>
      )}

      <form onSubmit={onSubmit}>
        <div className="mb-3">
          <label className="form-label">Nuova password</label>
          <input
            className="form-control"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="new-password"
            disabled={tokenMissing || status === "loading"}
          />
          <div className="form-text">
            Min 8 caratteri, 1 maiuscola, 1 carattere speciale. Deve essere diversa dalla precedente.
          </div>
        </div>

        <div className="mb-3">
          <label className="form-label">Conferma password</label>
          <input
            className="form-control"
            type="password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            autoComplete="new-password"
            disabled={tokenMissing || status === "loading"}
          />
        </div>

        <button
          className="btn btn-primary"
          type="submit"
          disabled={tokenMissing || status === "loading"}
        >
          {status === "loading" ? "Salvataggio..." : "Reimposta password"}
        </button>

        <div className="mt-3">
          <Link className="link-primary" href="/login">
            Torna al login
          </Link>
        </div>
      </form>
    </main>
  );
}
