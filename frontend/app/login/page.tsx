"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api, setToken } from "@/lib/api";

function validatePasswordAgID(pw: string): string | null {
  if (pw.length < 8) return "La password deve essere lunga almeno 8 caratteri.";
  if (!/[A-Z]/.test(pw)) return "La password deve contenere almeno una lettera maiuscola.";
  if (!/[!@#$%^&*()_\-+=\[\]{};:'",.<>/?\\|`~]/.test(pw))
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

    if (!username || !password) {
      setError("Username e password sono obbligatori.");
      return;
    }

    const pwErr = validatePasswordAgID(password);
    if (pwErr) {
      setError(pwErr);
      return;
    }

    try {
      setLoading(true);
      const res = await api.login(username, password);
      setToken(res.access_token);
      router.push("/dashboard");
    } catch (err: unknown) {
      setError(getErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="container py-5" style={{ maxWidth: 520 }}>
      <h1 className="mb-4">Login</h1>

      {error && (
        <div className="alert alert-danger" role="alert">
          {error}
        </div>
      )}

      <form onSubmit={onSubmit}>
        <div className="mb-3">
          <label className="form-label">Username</label>
          <input
            className="form-control"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
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
          />
          <div className="form-text">
            Min 8 caratteri, 1 maiuscola, 1 carattere speciale.
          </div>
        </div>

        <button className="btn btn-primary" type="submit" disabled={loading}>
          {loading ? "Accesso..." : "Accedi"}
        </button>
      </form>
    </main>
  );
}
