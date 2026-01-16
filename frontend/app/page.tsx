export default function Home() {
  return (
    <main className="container py-5">
      <h1 className="mb-4">Bootstrap Italia test</h1>

      <button className="btn btn-primary me-2">
        Bottone primario
      </button>

      <button className="btn btn-outline-primary">
        Bottone outline
      </button>

      <div className="alert alert-success mt-4" role="alert">
        Se vedi questo alert verde e i bottoni stilati, Bootstrap Italia è caricato ✅
      </div>
    </main>
  );
}
