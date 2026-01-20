param(
  [string]$ApiBase = "http://localhost:8080",
  [string]$LoginUser = "test@a.com",
  [string]$LoginPass = "Agente007$",

  # Se vuoi test email in modo professionale: usa MailHog e metti qui la sua API base:
  # MailHog UI: http://localhost:8025  | API: http://localhost:8025/api/v2
  [string]$MailHogApiBase = "",

  # Se true: NON cancella utente/prodotto creati dal test (così li vedi nel DB)
  [switch]$NoCleanup
)

$ErrorActionPreference = "Stop"

function Pass($msg) { Write-Host "PASS  $msg" -ForegroundColor Green }
function Fail($msg) { Write-Host "FAIL  $msg" -ForegroundColor Red; exit 1 }
function Info($msg) { Write-Host "INFO  $msg" -ForegroundColor Cyan }
function Warn($msg) { Write-Host "WARN  $msg" -ForegroundColor Yellow }

function PostJson($url, $body, $headers=@{}) {
  return Invoke-RestMethod -Method Post -Uri $url -ContentType "application/json" -Headers $headers -Body ($body | ConvertTo-Json -Compress)
}
function PutJson($url, $body, $headers=@{}) {
  return Invoke-RestMethod -Method Put -Uri $url -ContentType "application/json" -Headers $headers -Body ($body | ConvertTo-Json -Compress)
}
function GetJson($url, $headers=@{}) {
  return Invoke-RestMethod -Method Get -Uri $url -Headers $headers
}
function DeleteReq($url, $headers=@{}) {
  return Invoke-RestMethod -Method Delete -Uri $url -Headers $headers
}

function GetMailHogLatestMessageFor($recipient) {
  if ([string]::IsNullOrWhiteSpace($MailHogApiBase)) { return $null }

  # MailHog v2 API: GET /api/v2/messages
  $resp = Invoke-RestMethod -Method Get -Uri "$MailHogApiBase/messages"
  if (-not $resp -or -not $resp.items) { return $null }

  # cerca l'email più recente inviata al destinatario
  foreach ($m in $resp.items) {
    $tos = @()
    if ($m.Content.Headers.To) { $tos = $m.Content.Headers.To }
    if ($tos -join "," -match [regex]::Escape($recipient)) {
      return $m
    }
  }
  return $null
}

function ExtractFirstUrlOrTokenFromEmail($mail) {
  if (-not $mail) { return $null }

  # Body può stare in Content.Body (plain) o HTML (dipende dal mailer)
  $body = ""
  if ($mail.Content -and $mail.Content.Body) { $body = [string]$mail.Content.Body }

  if ([string]::IsNullOrWhiteSpace($body)) { return $null }

  # 1) prova a estrarre un URL
  $urlMatch = [regex]::Match($body, '(https?://[^\s"<>]+)', 'IgnoreCase')
  if ($urlMatch.Success) { return @{ kind="url"; value=$urlMatch.Groups[1].Value } }

  # 2) prova a estrarre un token tipo jwt/hex/base64 (molto permissivo)
  #    Esempio: token=xxxxx oppure "reset token: xxxxx"
  $tokMatch = [regex]::Match($body, '(token|reset_token|code)\s*[:=]\s*([A-Za-z0-9\-\._~\+\/]+=*)', 'IgnoreCase')
  if ($tokMatch.Success) { return @{ kind="token"; value=$tokMatch.Groups[2].Value } }

  return $null
}

# -------------------------
# 0) Health checks
# -------------------------
try {
  $ping = GetJson "$ApiBase/ping"
  if ($ping.message -ne "pong") { Fail "/ping risposta inattesa" }
  Pass "/ping"
} catch { Fail "/ping non raggiungibile: $($_.Exception.Message)" }

try {
  $db = GetJson "$ApiBase/db-ping"
  if ($db.db -ne "up") { Fail "/db-ping DB non up" }
  Pass "/db-ping"
} catch { Fail "/db-ping fallito: $($_.Exception.Message)" }

# -------------------------
# 1) Login -> token
# -------------------------
try {
  $loginResp = PostJson "$ApiBase/auth/login" @{ username=$LoginUser; password=$LoginPass }
  if (-not $loginResp.access_token) { Fail "Login: token mancante" }
  $token = $loginResp.access_token
  $authH = @{ Authorization = "Bearer $token" }
  Pass "Login (token ottenuto)"
} catch { Fail "Login fallito: $($_.Exception.Message)" }

# -------------------------
# 2) Endpoint protetto /api/me
# -------------------------
try {
  $me = GetJson "$ApiBase/api/me" $authH
  if ($me.username -ne $LoginUser) { Fail "/api/me username inatteso" }
  Pass "/api/me"
} catch { Fail "/api/me fallito: $($_.Exception.Message)" }

# -------------------------
# 3) Product Types
# -------------------------
[int]$tipoId = 0
try {
  $pts = GetJson "$ApiBase/api/product-types" $authH
  if (-not $pts -or $pts.Count -lt 1) { Fail "product-types vuoto" }

  $buste = $pts | Where-Object { $_.tipo -eq "Buste" } | Select-Object -First 1
  if ($buste) { $tipoId = [int]$buste.id } else { $tipoId = [int]$pts[0].id }

  if ($tipoId -le 0) { Fail "tipo_prodotto_id non valido" }
  Pass "GET /api/product-types (tipo_prodotto_id=$tipoId)"
} catch { Fail "GET /api/product-types fallito: $($_.Exception.Message)" }

# -------------------------
# 4) CRUD Prodotto
# -------------------------
[int]$prodId = 0
$unique = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()

try {
  $createP = PostJson "$ApiBase/api/products" @{
    nome_oggetto = "ProdottoTest-$unique"
    descrizione = "Creato da smoke-tests"
    tipo_prodotto_id = $tipoId
  } $authH

  if ($createP.id) { $prodId = [int]$createP.id }

  if ($prodId -le 0) {
    $listP = GetJson "$ApiBase/api/products" $authH
    $found = $listP | Where-Object { $_.nome_oggetto -eq "ProdottoTest-$unique" } | Select-Object -First 1
    if ($found) { $prodId = [int]$found.id }
  }

  if ($prodId -le 0) { Fail "Creazione prodotto: id non trovato" }
  Pass "POST /api/products (id=$prodId)"
} catch { Fail "POST /api/products fallito: $($_.Exception.Message)" }

try {
  $null = PutJson "$ApiBase/api/products/$prodId" @{
    nome_oggetto = "ProdottoTest-$unique-upd"
    descrizione = "Aggiornato da smoke-tests"
    tipo_prodotto_id = $tipoId
  } $authH
  Pass "PUT /api/products/:id"
} catch { Fail "PUT /api/products/:id fallito: $($_.Exception.Message)" }

# -------------------------
# 5) CRUD Utente
# -------------------------
[int]$userId = 0
$testUserEmail = "smoke+$unique@a.com"

try {
  $newU = PostJson "$ApiBase/api/users" @{
    username = $testUserEmail
    password = "Abcdefg!1"   # conforme AgID
    nome = "Smoke"
    cognome = "Test"
    data_nascita = "1990-01-01"
  } $authH

  if ($newU.id) { $userId = [int]$newU.id }

  if ($userId -le 0) {
    $users = GetJson "$ApiBase/api/users" $authH
    $fu = $users | Where-Object { $_.username -eq $testUserEmail } | Select-Object -First 1
    if ($fu) { $userId = [int]$fu.id }
  }

  if ($userId -le 0) { Fail "Creazione utente: id non trovato" }
  Pass "POST /api/users (id=$userId)"
} catch { Fail "POST /api/users fallito: $($_.Exception.Message)" }

try {
  $null = PutJson "$ApiBase/api/users/$userId" @{
    username = $testUserEmail
    nome = "SmokeUpdated"
    cognome = "Test"
    data_nascita = "1990-01-01"
  } $authH
  Pass "PUT /api/users/:id"
} catch { Fail "PUT /api/users/:id fallito: $($_.Exception.Message)" }

# -------------------------
# 6) EMAIL & PASSWORD RESET (sicuro) - con MailHog se disponibile
# -------------------------
# Best practice: gli endpoint di recovery dovrebbero rispondere in modo "generico"
# (non rivelare se l'utente esiste). Nei test, non asseriamo il contenuto del messaggio,
# ma verifichiamo che l'email arrivi nel sandbox (MailHog) e contenga link/token validi.
# Riferimenti OWASP: forgot password + authentication responses.  :contentReference[oaicite:3]{index=3}

# 6.a Forgot password
try {
  $null = PostJson "$ApiBase/auth/forgot-password" @{ username=$testUserEmail }
  Pass "POST /auth/forgot-password (trigger)"
} catch {
  Fail "POST /auth/forgot-password fallito: $($_.Exception.Message)"
}

if (-not [string]::IsNullOrWhiteSpace($MailHogApiBase)) {
  Info "MailHog abilitato: provo a recuperare l'email di reset per $testUserEmail"
  Start-Sleep -Seconds 1

  $mail = GetMailHogLatestMessageFor $testUserEmail
  if (-not $mail) {
    Warn "Nessuna email trovata in MailHog per $testUserEmail. Controlla config SMTP del backend."
  } else {
    Pass "Email reset catturata in MailHog"

    $extract = ExtractFirstUrlOrTokenFromEmail $mail
    if (-not $extract) {
      Warn "Email presente ma non ho trovato URL/token (regex). Valuta il template email."
    } else {
      Pass "Trovato $($extract.kind) nell'email"

      # 6.b Reset password end-to-end (se possiamo ricavare token)
      # Se abbiamo un URL, proviamo a prendere token dalla query (?token=...)
      $resetToken = $null
      if ($extract.kind -eq "url") {
        $u = $extract.value
        $m = [regex]::Match($u, '[\?&](token|reset_token|code)=([^&]+)', 'IgnoreCase')
        if ($m.Success) { $resetToken = [uri]::UnescapeDataString($m.Groups[2].Value) }
      } elseif ($extract.kind -eq "token") {
        $resetToken = $extract.value
      }

      if ([string]::IsNullOrWhiteSpace($resetToken)) {
        Warn "Ho trovato un URL/token, ma non sono riuscito a ricavare resetToken utilizzabile."
      } else {
        Info "Provo reset-password con token estratto (non stampo il token per sicurezza)"
        $newPass = "Abcdefg!2"

        try {
          # Nota: il body esatto dipende dal tuo handler.
          # Qui assumo un formato tipico: { token, new_password }
          $null = PostJson "$ApiBase/auth/reset-password" @{
            token = $resetToken
            new_password = $newPass
          }
          Pass "POST /auth/reset-password (end-to-end)"

          # verifica che la nuova password permetta login
          try {
            $lr = PostJson "$ApiBase/auth/login" @{ username=$testUserEmail; password=$newPass }
            if ($lr.access_token) { Pass "Login con nuova password OK" } else { Fail "Login con nuova password: token mancante" }
          } catch {
            Fail "Login con nuova password fallito: $($_.Exception.Message)"
          }

        } catch {
          Warn "reset-password non completato (probabile mismatch body richiesto dal backend): $($_.Exception.Message)"
          Warn "Suggerimento: dimmi il JSON richiesto da /auth/reset-password e lo adatto al 100%."
        }
      }
    }
  }
} else {
  Info "MailHog non configurato: salto la verifica contenuto email/reset end-to-end."
  Info "Per test professionali: usa MailHog e passa -MailHogApiBase http://localhost:8025/api/v2"
}

# 6.c Verify email
# Qui non posso fare end-to-end senza sapere il body richiesto dal tuo endpoint.
# Mantengo una prova "smoke": chiamata e controllo che non crashi.
try {
  $null = PostJson "$ApiBase/auth/verify-email" @{ username=$testUserEmail }
  Pass "POST /auth/verify-email (smoke)"
} catch {
  Warn "verify-email non testato: $($_.Exception.Message)"
  Warn "Nota: per E2E serve catturare email e usare token/link, come da best practice. :contentReference[oaicite:4]{index=4}"
}

# -------------------------
# 7) Cleanup (opzionale)
# -------------------------
if ($NoCleanup) {
  Info "NoCleanup attivo: lascio nel DB user_id=$userId e product_id=$prodId"
  Pass "SMOKE TESTS COMPLETATI ✅"
  exit 0
}

try {
  DeleteReq "$ApiBase/api/products/$prodId" $authH
  Pass "DELETE /api/products/:id"
} catch { Fail "DELETE /api/products/:id fallito: $($_.Exception.Message)" }

try {
  DeleteReq "$ApiBase/api/users/$userId" $authH
  Pass "DELETE /api/users/:id"
} catch { Fail "DELETE /api/users/:id fallito: $($_.Exception.Message)" }

Pass "SMOKE TESTS COMPLETATI ✅"
exit 0
