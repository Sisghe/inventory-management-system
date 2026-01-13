package models

import "time"

type Product struct {
	ID              int       `json:"id"`
	NomeOggetto     string    `json:"nome_oggetto"`
	Descrizione     *string   `json:"descrizione"`
	DataInserimento time.Time `json:"data_inserimento"`
	TipoProdottoID  *int      `json:"tipo_prodotto_id"`
}
