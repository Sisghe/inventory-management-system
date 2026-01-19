"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";

function validatePasswordAgID(pw: string): string | null {
  if (pw.length < 8) return "La password deve essere lunga almeno 8 caratteri.";
  if (!/[A-Z]/.test(pw)) return "La password deve contenere almeno una lettera maiuscola.";
  if (!/[!@#$%^&*()_\-+=\[\]{};:'\",.<>/?\\|`~]/.test(pw))
    return "La password deve contenere almeno un carattere speciale.";
  return null;
}

function getErrorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return "Errore di login";
}

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    const u = username.trim();
    const p = password.trim();

    if (!u || !p) {
      setError("Username e password sono obbligatori.");
      return;
    }

    const pwErr = validatePasswordAgID(p);
    if (pwErr) {
      setError(pwErr);
      return;
    }

    try {
      setLoading(true);

      // Il backend imposta il cookie HttpOnly access_token.
      await api.login(u, p);

      router.push("/dashboard");
      router.refresh();
    } catch (err: unknown) {
      setError(getErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="container py-4 py-md-5">
      <div className="row justify-content-center">
        <div className="col-12 col-sm-10 col-md-7 col-lg-5">
          <h1 className="h3 mb-3 mb-md-4 text-center">Login</h1>

          {error && (
            <div className="alert alert-danger" role="alert">
              {error}
            </div>
          )}

          <div className="card">
            <div className="card-body">
              <form onSubmit={onSubmit}>
                <div className="mb-3">
                  <label className="form-label">Username</label>
                  <input
                    className="form-control"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    autoComplete="username"
                    inputMode="email"
                    placeholder="es. nome.cognome@email.it"
                  />
                </div>

                <div className="mb-3">
                  <label className="form-label">Password</label>
                  <input
                    className="form-control"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="current-password"
                    placeholder="Inserisci la tua password"
                  />
                  <div className="form-text">
                    Min 8 caratteri, 1 maiuscola, 1 carattere speciale.
                  </div>
                </div>

                <div className="d-grid d-sm-flex gap-2">
                  <button className="btn btn-primary" type="submit" disabled={loading}>
                    {loading ? "Accesso..." : "Accedi"}
                  </button>

                  <Link className="btn btn-outline-primary" href="/forgot-password">
                    Recupera password
                  </Link>
                </div>
              </form>
            </div>
          </div>

          <div className="text-center mt-3">
            <Link className="link-primary" href="/">
              Torna alla Home
            </Link>
          </div>
        </div>
      </div>
    </main>
  );
}
