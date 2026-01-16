"use client";

import { useEffect, useMemo, useState } from "react";
import { api, UserDTO } from "@/lib/api";

function validatePasswordAgID(pw: string): string | null {
  if (pw.length < 8) return "La password deve essere lunga almeno 8 caratteri.";
  if (!/[A-Z]/.test(pw)) return "La password deve contenere almeno una lettera maiuscola.";
  if (!/[!@#$%^&*()_\-+=\[\]{};:'\",.<>/?\\|`~]/.test(pw))
    return "La password deve contenere almeno un carattere speciale.";
  return null;
}

function normalizeDate(value?: string) {
  if (!value) return "";
  return value.length >= 10 ? value.slice(0, 10) : value;
}

function getErrMsg(e: unknown) {
  return e instanceof Error ? e.message : "Errore";
}

type EditState = {
  open: boolean;
  userId: number | null;
  username: string;
  password: string; // opzionale: se vuota non aggiorna password
  nome: string;
  cognome: string;
  data_nascita: string; // YYYY-MM-DD (se vuota => non inviare)
};

export default function UsersPage() {
  const [users, setUsers] = useState<UserDTO[]>([]);
  const [loading, setLoading] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Form create
  const [cUsername, setCUsername] = useState("");
  const [cPassword, setCPassword] = useState("");
  const [cNome, setCNome] = useState("");
  const [cCognome, setCCognome] = useState("");
  const [cDataNascita, setCDataNascita] = useState("");

  // Modal edit
  const [edit, setEdit] = useState<EditState>({
    open: false,
    userId: null,
    username: "",
    password: "",
    nome: "",
    cognome: "",
    data_nascita: "",
  });

  const sortedUsers = useMemo(() => {
    return [...users].sort((a, b) => (a.id ?? 0) - (b.id ?? 0));
  }, [users]);

  async function loadUsers() {
    setError(null);
    setSuccess(null);
    setLoading(true);
    try {
      const list = await api.users.list();
      setUsers(list);
    } catch (e) {
      setError(getErrMsg(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadUsers();
  }, []);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    const username = cUsername.trim();
    const password = cPassword.trim();

    if (!username || !password) {
      setError("Username e password sono obbligatori.");
      return;
    }

    const pwErr = validatePasswordAgID(password);
    if (pwErr) {
      setError(pwErr);
      return;
    }

    const payload: UserDTO = { username, password };
    if (cNome.trim()) payload.nome = cNome.trim();
    if (cCognome.trim()) payload.cognome = cCognome.trim();
    if (cDataNascita.trim()) payload.data_nascita = cDataNascita.trim();

    try {
      setLoading(true);
      await api.users.create(payload);
      setSuccess("Utente creato correttamente.");

      setCUsername("");
      setCPassword("");
      setCNome("");
      setCCognome("");
      setCDataNascita("");

      await loadUsers();
    } catch (e) {
      setError(getErrMsg(e));
    } finally {
      setLoading(false);
    }
  }

  function openEdit(u: UserDTO) {
    setError(null);
    setSuccess(null);

    setEdit({
      open: true,
      userId: u.id ?? null,
      username: u.username ?? "",
      password: "",
      nome: u.nome ?? "",
      cognome: u.cognome ?? "",
      data_nascita: normalizeDate(u.data_nascita),
    });
  }

  function closeEdit() {
    setEdit((s) => ({ ...s, open: false, userId: null }));
  }

  async function onUpdate(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!edit.userId) {
      setError("ID utente non valido.");
      return;
    }

    const username = edit.username.trim();
    if (!username) {
      setError("Lo username non può essere vuoto.");
      return;
    }

    const password = edit.password.trim();
    if (password) {
      const pwErr = validatePasswordAgID(password);
      if (pwErr) {
        setError(pwErr);
        return;
      }
    }

    const payload: UserDTO = { username };
    if (password) payload.password = password;
    payload.nome = edit.nome.trim() || undefined;
    payload.cognome = edit.cognome.trim() || undefined;

    // IMPORTANT: il tuo backend rifiuta data_nascita vuota: se vuota NON inviarla
    if (edit.data_nascita.trim()) payload.data_nascita = edit.data_nascita.trim();

    try {
      setLoading(true);
      await api.users.update(edit.userId, payload);
      setSuccess("Utente aggiornato correttamente.");
      closeEdit();
      await loadUsers();
    } catch (e) {
      setError(getErrMsg(e));
    } finally {
      setLoading(false);
    }
  }

  async function onDelete(u: UserDTO) {
    setError(null);
    setSuccess(null);

    const id = u.id;
    if (!id) {
      setError("ID utente non valido.");
      return;
    }

    const ok = window.confirm(`Vuoi davvero eliminare l'utente "${u.username}" (id: ${id})?`);
    if (!ok) return;

    try {
      setLoading(true);
      await api.users.delete(id);
      setSuccess("Utente eliminato correttamente.");
      await loadUsers();
    } catch (e) {
      setError(getErrMsg(e));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="container-fluid p-0">
      <div className="d-flex align-items-center justify-content-between mb-3">
        <h1 className="h4 mb-0">Gestione utenti</h1>
        <button className="btn btn-outline-primary" onClick={loadUsers} disabled={loading} type="button">
          {loading ? "Caricamento..." : "Ricarica"}
        </button>
      </div>

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

      {/* Form creazione */}
      <div className="card mb-4">
        <div className="card-body">
          <h2 className="h6 mb-3">Crea nuovo utente</h2>
          <form onSubmit={onCreate}>
            <div className="row g-3">
              <div className="col-md-6">
                <label className="form-label">Username *</label>
                <input className="form-control" value={cUsername} onChange={(e) => setCUsername(e.target.value)} />
              </div>
              <div className="col-md-6">
                <label className="form-label">Password *</label>
                <input
                  className="form-control"
                  type="password"
                  value={cPassword}
                  onChange={(e) => setCPassword(e.target.value)}
                  autoComplete="new-password"
                />
                <div className="form-text">Min 8 caratteri, 1 maiuscola, 1 carattere speciale.</div>
              </div>

              <div className="col-md-4">
                <label className="form-label">Nome</label>
                <input className="form-control" value={cNome} onChange={(e) => setCNome(e.target.value)} />
              </div>
              <div className="col-md-4">
                <label className="form-label">Cognome</label>
                <input className="form-control" value={cCognome} onChange={(e) => setCCognome(e.target.value)} />
              </div>
              <div className="col-md-4">
                <label className="form-label">Data nascita</label>
                <input
                  className="form-control"
                  type="date"
                  value={cDataNascita}
                  onChange={(e) => setCDataNascita(e.target.value)}
                />
              </div>

              <div className="col-12">
                <button className="btn btn-primary" type="submit" disabled={loading}>
                  {loading ? "Salvataggio..." : "Crea utente"}
                </button>
              </div>
            </div>
          </form>
        </div>
      </div>

      {/* Tabella utenti */}
      <div className="card">
        <div className="card-body">
          <h2 className="h6 mb-3">Elenco utenti</h2>

          {sortedUsers.length === 0 ? (
            <p className="mb-0">Nessun utente trovato.</p>
          ) : (
            <div className="table-responsive">
              <table className="table">
                <thead>
                  <tr>
                    <th style={{ width: 80 }}>ID</th>
                    <th>Username</th>
                    <th>Nome</th>
                    <th>Cognome</th>
                    <th style={{ width: 140 }}>Data nascita</th>
                    <th style={{ width: 200 }}>Azioni</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedUsers.map((u) => (
                    <tr key={u.id}>
                      <td>{u.id}</td>
                      <td>{u.username}</td>
                      <td>{u.nome ?? ""}</td>
                      <td>{u.cognome ?? ""}</td>
                      <td>{normalizeDate(u.data_nascita)}</td>
                      <td>
                        <div className="d-flex gap-2">
                          <button className="btn btn-sm btn-outline-primary" type="button" onClick={() => openEdit(u)}>
                            Modifica
                          </button>
                          <button className="btn btn-sm btn-outline-danger" type="button" onClick={() => onDelete(u)}>
                            Elimina
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Modal modifica */}
      {edit.open && (
        <div
          className="modal d-block"
          tabIndex={-1}
          role="dialog"
          aria-modal="true"
          style={{ background: "rgba(0,0,0,0.5)" }}
          onClick={closeEdit}
        >
          <div className="modal-dialog" role="document" onClick={(e) => e.stopPropagation()}>
            <div className="modal-content">
              <div className="modal-header">
                <h3 className="modal-title h6 mb-0">Modifica utente</h3>
                <button type="button" className="btn-close" aria-label="Close" onClick={closeEdit} />
              </div>

              <form onSubmit={onUpdate}>
                <div className="modal-body">
                  <div className="mb-3">
                    <label className="form-label">Username *</label>
                    <input
                      className="form-control"
                      value={edit.username}
                      onChange={(e) => setEdit((s) => ({ ...s, username: e.target.value }))}
                    />
                  </div>

                  <div className="mb-3">
                    <label className="form-label">Nuova password (opzionale)</label>
                    <input
                      className="form-control"
                      type="password"
                      value={edit.password}
                      onChange={(e) => setEdit((s) => ({ ...s, password: e.target.value }))}
                      autoComplete="new-password"
                    />
                    <div className="form-text">Se lasci vuoto, la password non viene modificata.</div>
                  </div>

                  <div className="row g-3">
                    <div className="col-md-6">
                      <label className="form-label">Nome</label>
                      <input
                        className="form-control"
                        value={edit.nome}
                        onChange={(e) => setEdit((s) => ({ ...s, nome: e.target.value }))}
                      />
                    </div>
                    <div className="col-md-6">
                      <label className="form-label">Cognome</label>
                      <input
                        className="form-control"
                        value={edit.cognome}
                        onChange={(e) => setEdit((s) => ({ ...s, cognome: e.target.value }))}
                      />
                    </div>
                    <div className="col-md-12">
                      <label className="form-label">Data nascita</label>
                      <input
                        className="form-control"
                        type="date"
                        value={edit.data_nascita}
                        onChange={(e) => setEdit((s) => ({ ...s, data_nascita: e.target.value }))}
                      />
                      <div className="form-text">
                        Se lasci vuoto, la data resta invariata (non viene inviata).
                      </div>
                    </div>
                  </div>
                </div>

                <div className="modal-footer">
                  <button type="button" className="btn btn-outline-secondary" onClick={closeEdit}>
                    Annulla
                  </button>
                  <button type="submit" className="btn btn-primary" disabled={loading}>
                    {loading ? "Salvataggio..." : "Salva"}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
