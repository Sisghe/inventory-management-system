"use client";

import { useEffect, useMemo, useState } from "react";
import { api, ProductDTO, ProductTypeDTO } from "@/lib/api";

function getErrMsg(e: unknown) {
  return e instanceof Error ? e.message : "Errore";
}

function isoToDate(iso?: string) {
  if (!iso) return "";
  return iso.length >= 10 ? iso.slice(0, 10) : iso;
}

type EditState = {
  open: boolean;
  productId: number | null;
  nome_oggetto: string;
  descrizione: string;
  tipo_prodotto_id: string; // string per select
};

export default function InventoryPage() {
  const [types, setTypes] = useState<ProductTypeDTO[]>([]);
  const [products, setProducts] = useState<ProductDTO[]>([]);
  const [loading, setLoading] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Create form
  const [cNome, setCNome] = useState("");
  const [cDescrizione, setCDescrizione] = useState("");
  const [cTipoId, setCTipoId] = useState(""); // "" = non selezionato

  // Edit modal
  const [edit, setEdit] = useState<EditState>({
    open: false,
    productId: null,
    nome_oggetto: "",
    descrizione: "",
    tipo_prodotto_id: "",
  });

  const typeMap = useMemo(() => {
    const m = new Map<number, string>();
    for (const t of types) m.set(t.id, t.tipo);
    return m;
  }, [types]);

  const sortedProducts = useMemo(() => {
    return [...products].sort((a, b) => (a.id ?? 0) - (b.id ?? 0));
  }, [products]);

  async function loadAll() {
    setError(null);
    setSuccess(null);
    setLoading(true);
    try {
      const [t, p] = await Promise.all([api.productTypes.list(), api.products.list()]);
      setTypes(t);
      setProducts(p);
    } catch (e) {
      setError(getErrMsg(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadAll();
  }, []);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    const nome = cNome.trim();
    if (!nome) {
      setError("Il campo Nome oggetto è obbligatorio.");
      return;
    }

    const payload: ProductDTO = {
      nome_oggetto: nome,
    };

    const desc = cDescrizione.trim();
    if (desc) payload.descrizione = desc;

    if (cTipoId) payload.tipo_prodotto_id = Number(cTipoId);

    try {
      setLoading(true);
      await api.products.create(payload);
      setSuccess("Prodotto creato correttamente.");
      setCNome("");
      setCDescrizione("");
      setCTipoId("");
      await loadAll();
    } catch (e2) {
      setError(getErrMsg(e2));
    } finally {
      setLoading(false);
    }
  }

  function openEdit(p: ProductDTO) {
    setError(null);
    setSuccess(null);

    setEdit({
      open: true,
      productId: p.id ?? null,
      nome_oggetto: p.nome_oggetto ?? "",
      descrizione: (p.descrizione ?? "") as string,
      tipo_prodotto_id: p.tipo_prodotto_id != null ? String(p.tipo_prodotto_id) : "",
    });
  }

  function closeEdit() {
    setEdit((s) => ({ ...s, open: false, productId: null }));
  }

  async function onUpdate(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!edit.productId) {
      setError("ID prodotto non valido.");
      return;
    }

    const nome = edit.nome_oggetto.trim();
    if (!nome) {
      setError("Il campo Nome oggetto è obbligatorio.");
      return;
    }

    const payload: ProductDTO = {
      nome_oggetto: nome,
      descrizione: edit.descrizione.trim() ? edit.descrizione.trim() : null,
      tipo_prodotto_id: edit.tipo_prodotto_id ? Number(edit.tipo_prodotto_id) : null,
    };

    try {
      setLoading(true);
      await api.products.update(edit.productId, payload);
      setSuccess("Prodotto aggiornato correttamente.");
      closeEdit();
      await loadAll();
    } catch (e2) {
      setError(getErrMsg(e2));
    } finally {
      setLoading(false);
    }
  }

  async function onDelete(p: ProductDTO) {
    setError(null);
    setSuccess(null);

    const id = p.id;
    if (!id) {
      setError("ID prodotto non valido.");
      return;
    }

    const ok = window.confirm(`Vuoi davvero eliminare il prodotto "${p.nome_oggetto}" (id: ${id})?`);
    if (!ok) return;

    try {
      setLoading(true);
      await api.products.delete(id);
      setSuccess("Prodotto eliminato correttamente.");
      await loadAll();
    } catch (e2) {
      setError(getErrMsg(e2));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="container-fluid p-0">
      <div className="d-flex align-items-center justify-content-between mb-3">
        <h1 className="h4 mb-0">Gestione inventario</h1>
        <button className="btn btn-outline-primary" onClick={loadAll} disabled={loading} type="button">
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

      {/* Create form */}
      <div className="card mb-4">
        <div className="card-body">
          <h2 className="h6 mb-3">Inserisci nuovo prodotto</h2>

          <form onSubmit={onCreate}>
            <div className="row g-3">
              <div className="col-md-4">
                <label className="form-label">Nome oggetto *</label>
                <input className="form-control" value={cNome} onChange={(e) => setCNome(e.target.value)} />
              </div>

              <div className="col-md-4">
                <label className="form-label">Tipo prodotto</label>
                <select className="form-select" value={cTipoId} onChange={(e) => setCTipoId(e.target.value)}>
                  <option value="">Seleziona…</option>
                  {types.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.tipo}
                    </option>
                  ))}
                </select>
                <div className="form-text">Tipi ammessi: Buste, Carta, Toner.</div>
              </div>

              <div className="col-md-4">
                <label className="form-label">Descrizione</label>
                <input
                  className="form-control"
                  value={cDescrizione}
                  onChange={(e) => setCDescrizione(e.target.value)}
                />
              </div>

              <div className="col-12">
                <button className="btn btn-primary" type="submit" disabled={loading}>
                  {loading ? "Salvataggio..." : "Crea prodotto"}
                </button>
              </div>
            </div>
          </form>
        </div>
      </div>

      {/* Products table */}
      <div className="card">
        <div className="card-body">
          <h2 className="h6 mb-3">Elenco prodotti</h2>

          {sortedProducts.length === 0 ? (
            <p className="mb-0">Nessun prodotto trovato.</p>
          ) : (
            <div className="table-responsive">
              <table className="table">
                <thead>
                  <tr>
                    <th style={{ width: 80 }}>ID</th>
                    <th>Nome oggetto</th>
                    <th>Tipo</th>
                    <th>Descrizione</th>
                    <th style={{ width: 140 }}>Data inserimento</th>
                    <th style={{ width: 200 }}>Azioni</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedProducts.map((p) => (
                    <tr key={p.id}>
                      <td>{p.id}</td>
                      <td>{p.nome_oggetto}</td>
                      <td>
                        {p.tipo_prodotto_id != null ? typeMap.get(p.tipo_prodotto_id) ?? `ID ${p.tipo_prodotto_id}` : ""}
                      </td>
                      <td>{(p.descrizione ?? "") as string}</td>
                      <td>{isoToDate(p.data_inserimento)}</td>
                      <td>
                        <div className="d-flex gap-2">
                          <button className="btn btn-sm btn-outline-primary" type="button" onClick={() => openEdit(p)}>
                            Modifica
                          </button>
                          <button className="btn btn-sm btn-outline-danger" type="button" onClick={() => onDelete(p)}>
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

      {/* Modal edit */}
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
                <h3 className="modal-title h6 mb-0">Modifica prodotto</h3>
                <button type="button" className="btn-close" aria-label="Close" onClick={closeEdit} />
              </div>

              <form onSubmit={onUpdate}>
                <div className="modal-body">
                  <div className="mb-3">
                    <label className="form-label">Nome oggetto *</label>
                    <input
                      className="form-control"
                      value={edit.nome_oggetto}
                      onChange={(e) => setEdit((s) => ({ ...s, nome_oggetto: e.target.value }))}
                    />
                  </div>

                  <div className="mb-3">
                    <label className="form-label">Tipo prodotto</label>
                    <select
                      className="form-select"
                      value={edit.tipo_prodotto_id}
                      onChange={(e) => setEdit((s) => ({ ...s, tipo_prodotto_id: e.target.value }))}
                    >
                      <option value="">Seleziona…</option>
                      {types.map((t) => (
                        <option key={t.id} value={t.id}>
                          {t.tipo}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="mb-3">
                    <label className="form-label">Descrizione</label>
                    <input
                      className="form-control"
                      value={edit.descrizione}
                      onChange={(e) => setEdit((s) => ({ ...s, descrizione: e.target.value }))}
                    />
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
