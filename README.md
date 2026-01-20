# inventory-management-system
Full-stack web app for authenticated user and shop inventory management. Built with Next.js, Go (Gin), PostgreSQL, and Bootstrap Italia.


I'm working on the inventory management project and tracking the roadmap/step-by-step progress on Excalidraw.  
Here's the live diagram:  https://excalidraw.com/#room=f2bccebf6a2e3edd35af,SCTM4zUqIK3P7-ErG7yHYA

# Inventory Management System (Next.js + Go + PostgreSQL)

Applicativo web per gestione utenti e inventario.
- Frontend: Next.js (App Router) + Bootstrap Italia
- Backend: Go (Gin) + JWT
- Database: PostgreSQL

## Requisiti
- Node.js + npm
- Go (versione compatibile con `backend/go.mod`)
- PostgreSQL
- (Opzionale) Git

## Struttura
- `frontend/` → Next.js
- `backend/` → API Go (Gin)
- DB PostgreSQL con tabelle: `utente`, `prodotto`, `tipo_prodotto`

## Avvio in locale

### 1) Database
Crea un database PostgreSQL (es. `inventory_db`) e configura le variabili del backend .
Assicurati che esistano anche i 3 tipi prodotto:
- Buste
- Carta
- Toner

> Nota: il prodotto ha un riferimento a `tipo_prodotto_id` (FK logica).

### 2) Backend (Go)
Apri un terminale:

```bash
cd backend
go run .
