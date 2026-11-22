param(
  [string]$BaseUrl = "http://localhost:8080",
  [string]$AdminEmail = "",
  [string]$AdminPassword = ""
)

$ErrorActionPreference = "Stop"
$script:Pass = 0
$script:Fail = 0

# ---------------------------------------------------------------- Admin env
# Bootstrap RF09: o servidor so promove ADMIN_EMAILS no boot, entao o script
# garante o e-mail do admin no ambiente (herdado por qualquer processo filho
# que ele stesso iniciar) e, se necessario, reinicia o backend depois do
# registro (ver Restart-Backend / bloco 0b).
$script:Root = Split-Path -Parent $PSScriptRoot
if (-not $AdminEmail) {
  if ($env:RADIOPX_ADMIN_EMAIL) { $AdminEmail = $env:RADIOPX_ADMIN_EMAIL }
  else { $AdminEmail = "admin@radiopx.test" }
}
if (-not $AdminPassword) {
  if ($env:RADIOPX_ADMIN_PASSWORD) { $AdminPassword = $env:RADIOPX_ADMIN_PASSWORD }
  else { $AdminPassword = "admin12345" }
}
$script:AdminEmail = $AdminEmail
$script:AdminPassword = $AdminPassword

$adminEmails = @()
if ($env:ADMIN_EMAILS) { $adminEmails += ($env:ADMIN_EMAILS -split ',') }
foreach ($f in @((Join-Path $script:Root "backend\.env"), (Join-Path $script:Root ".env"))) {
  if (Test-Path $f) {
    foreach ($line in (Get-Content $f -ErrorAction SilentlyContinue)) {
      if ($line -match '^\s*ADMIN_EMAILS\s*=\s*(.+)$') {
        $adminEmails += ($Matches[1].Trim().Trim('"').Trim("'") -split ',')
      }
    }
  }
}
$adminEmails += $AdminEmail
$env:ADMIN_EMAILS = (($adminEmails | ForEach-Object { $_.Trim() } | Where-Object { $_ }) | Select-Object -Unique) -join ','

function Write-Step($Text) {
  Write-Host ""
  Write-Host "== $Text ==" -ForegroundColor Cyan
}

function Check($Name, $Condition, $Detail = "") {
  if ($Condition) {
    Write-Host "[PASS] $Name" -ForegroundColor Green
    $script:Pass++
  }
  else {
    Write-Host "[FAIL] $Name -> $Detail" -ForegroundColor Red
    $script:Fail++
  }
}

function Invoke-Json($Method, $Path, $Body = $null, $Headers = @{}) {
  # ngrok: sem este header, o "browser warning" (ERR_NGROK_6024) substitui a resposta da API
  $Headers['ngrok-skip-browser-warning'] = '1'
  $params = @{
    Uri     = "$BaseUrl$Path"
    Method  = $Method
    Headers = $Headers
  }
  if ($null -ne $Body) {
    $params.Body = ($Body | ConvertTo-Json -Compress -Depth 5)
    $params.ContentType = "application/json"
  }
  try {
    $resp = Invoke-WebRequest @params -UseBasicParsing
    return @{ Code = [int]$resp.StatusCode; Body = $resp.Content }
  }
  catch {
    if ($_.Exception.Response) {
      $code = [int]$_.Exception.Response.StatusCode
      $body = $null
      try {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $body = $reader.ReadToEnd()
      }
      catch { }
      return @{ Code = $code; Body = $body }
    }
    return @{ Code = -1; Body = $_.Exception.Message }
  }
}

function New-Ws([string]$Uri) {
  $ws = [System.Net.WebSockets.ClientWebSocket]::new()
  $ws.Options.SetRequestHeader('ngrok-skip-browser-warning', '1')
  [void]$ws.ConnectAsync([Uri]::new($Uri), [System.Threading.CancellationToken]::None).GetAwaiter().GetResult()
  return $ws
}

function Send-WsText($ws, [string]$Json) {
  $bytes = [System.Text.Encoding]::UTF8.GetBytes($Json)
  $seg = [System.ArraySegment[byte]]::new($bytes)
  [void]$ws.SendAsync($seg, [System.Net.WebSockets.WebSocketMessageType]::Text, $true, [System.Threading.CancellationToken]::None).GetAwaiter().GetResult()
}

function Receive-WsText($ws, [int]$TimeoutMs = 6000) {
  $buf = New-Object byte[] 65536
  $seg = [System.ArraySegment[byte]]::new($buf)
  $task = $ws.ReceiveAsync($seg, [System.Threading.CancellationToken]::None)
  if (-not $task.Wait($TimeoutMs)) { return $null }
  $res = $task.Result
  if ($res.MessageType -eq [System.Net.WebSockets.WebSocketMessageType]::Close) { return $null }
  $text = [System.Text.Encoding]::UTF8.GetString($buf, 0, $res.Count)
  if ($res.EndOfMessage) { return $text }
  return $text
}

function Close-Ws($ws) {
  try {
    [void]$ws.CloseOutputAsync([System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure, "", [System.Threading.CancellationToken]::None).GetAwaiter().GetResult()
  }
  catch { }
  $ws.Dispose()
}

# Reinicia o backend para aplicar o bootstrap de ADMIN_EMAILS (so roda no boot).
# Garante que backend/server.exe exista ANTES de matar o processo atual.
function Restart-Backend {
  $exe = Join-Path $script:Root "backend\server.exe"
  if (-not (Test-Path $exe)) {
    Write-Host "server.exe ausente - compilando backend..."
    Push-Location (Join-Path $script:Root "backend")
    try {
      & go build -o server.exe ./cmd/server
      if ($LASTEXITCODE -ne 0 -or -not (Test-Path $exe)) {
        throw "Falha ao compilar backend/server.exe (go build exit=$LASTEXITCODE)"
      }
    }
    finally { Pop-Location }
  }

  $port = 8080
  try { $port = ([uri]$BaseUrl).Port } catch { }
  $conns = @(Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue)
  foreach ($c in $conns) {
    try { Stop-Process -Id $c.OwningProcess -Force -ErrorAction Stop }
    catch { Write-Host "AVISO: nao foi possivel parar PID $($c.OwningProcess): $($_.Exception.Message)" -ForegroundColor Yellow }
  }
  if ($conns.Count -gt 0) { Start-Sleep -Seconds 1 }

  # Filho herda $env:ADMIN_EMAILS (definido no topo do script).
  Start-Process -FilePath $exe -WorkingDirectory (Split-Path $exe) -WindowStyle Hidden | Out-Null
  $ok = $false
  for ($i = 0; $i -lt 20; $i++) {
    Start-Sleep -Milliseconds 500
    $h = Invoke-Json "GET" "/health"
    if ($h.Code -eq 200) { $ok = $true; break }
  }
  if (-not $ok) { throw "Backend nao voltou a responder apos reinicio" }
  Write-Host "Backend reiniciado com ADMIN_EMAILS=$env:ADMIN_EMAILS" -ForegroundColor Yellow
}

# ---------------------------------------------------------------- 0. Health
Write-Step "0. Servidor"
$health = Invoke-Json "GET" "/health"
$script:ServerStartedByScript = $false
if ($health.Code -ne 200) {
  $exe = "C:\RadioPx\backend\server.exe"
  if (Test-Path $exe) {
    Write-Host "Servidor parado. Iniciando $exe ..."
    Start-Process -FilePath $exe -WorkingDirectory (Split-Path $exe) -WindowStyle Hidden | Out-Null
    $script:ServerStartedByScript = $true
    Start-Sleep -Seconds 2
    $health = Invoke-Json "GET" "/health"
  }
}
$healthOk = ($health.Code -eq 200 -and ($health.Body -match '"status"\s*:\s*"ok"'))
Check "Health check $BaseUrl/health" $healthOk $health.Body

# ---------------------------------------------------------------- 0b. Admin bootstrap (RF09)
# O servidor so promove ADMIN_EMAILS no boot. Fluxo:
#   1. reinicia o backend (se o script nao acabou de subi-lo) para zerar o
#      rate limit de auth da rodada anterior (10/min por IP) e aplicar env;
#   2. tenta logar; se o usuario nao existir, registra;
#   3. se ainda nao for admin (registro apos o boot), reinicia de novo para
#      o bootstrap promover e re-login.
Write-Step "0b. Admin bootstrap (ADMIN_EMAILS)"
$adminReady = $false
$tokenAd = $null
$adminUserId = $null

if (-not $script:ServerStartedByScript) {
  try { Restart-Backend }
  catch { Write-Host "AVISO: reinicio inicial falhou: $($_.Exception.Message)" -ForegroundColor Yellow }
}

$adminLogin = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $script:AdminEmail; password = $script:AdminPassword }
if ($adminLogin.Code -eq 429) {
  # Janela de rate limit ainda cheia (reinicio falhou?): uma unica nova tentativa.
  try {
    Restart-Backend
    $adminLogin = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $script:AdminEmail; password = $script:AdminPassword }
  } catch { }
}
if ($adminLogin.Code -eq 401) {
  $adminReg = Invoke-Json "POST" "/api/v1/auth/register" @{ name = "Admin RadioPX"; email = $script:AdminEmail; password = $script:AdminPassword }
  if ($adminReg.Code -eq 201 -or $adminReg.Code -eq 409) {
    $adminLogin = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $script:AdminEmail; password = $script:AdminPassword }
  }
}
if ($adminLogin.Code -eq 200) {
  try { $tokenAd = ($adminLogin.Body | ConvertFrom-Json).token } catch { }
}
Check "Admin bootstrap: login de $script:AdminEmail" (-not [string]::IsNullOrEmpty($tokenAd)) "$($adminLogin.Code) $($adminLogin.Body)"

if (-not [string]::IsNullOrEmpty($tokenAd)) {
  $hAd = @{ Authorization = "Bearer $tokenAd" }
  $isAdmin = $false
  $aprof = Invoke-Json "GET" "/api/v1/user/profile" $null $hAd
  if ($aprof.Code -eq 200) {
    try {
      $p = $aprof.Body | ConvertFrom-Json
      $isAdmin = ($p.is_admin -eq $true)
      $adminUserId = $p.id
    } catch { }
  }

  if (-not $isAdmin) {
    Write-Host "Usuario admin ainda nao promovido - reiniciando backend para aplicar bootstrap..." -ForegroundColor Yellow
    try {
      Restart-Backend
      # Reinicio limpa o rate limit; re-login garante token pos-boot.
      $adminLogin = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $script:AdminEmail; password = $script:AdminPassword }
      if ($adminLogin.Code -eq 200) {
        $tokenAd = ($adminLogin.Body | ConvertFrom-Json).token
        $hAd = @{ Authorization = "Bearer $tokenAd" }
        $aprof = Invoke-Json "GET" "/api/v1/user/profile" $null $hAd
        if ($aprof.Code -eq 200) {
          try {
            $p = $aprof.Body | ConvertFrom-Json
            $isAdmin = ($p.is_admin -eq $true)
            $adminUserId = $p.id
          } catch { }
        }
      }
    }
    catch {
      Write-Host "ERRO no reinicio: $($_.Exception.Message)" -ForegroundColor Red
    }
  }

  Check "Admin bootstrap: is_admin=true apos ADMIN_EMAILS" $isAdmin $aprof.Body
  $adminReady = $isAdmin
}

# ---------------------------------------------------------------- 1. Auth
Write-Step "1. Auth"
$suffix = [DateTimeOffset]::Now.ToUnixTimeSeconds()
$emailA = "a$suffix@px.com"
$emailB = "b$suffix@px.com"
$passA = "senha123"
$passNew = "nova1234"

$regA = Invoke-Json "POST" "/api/v1/auth/register" @{ name = "Usuario A"; email = $emailA; password = $passA }
Check "Register usuario A (201)" $regA.Code -eq 201 $regA.Body
$userA = $null; $tokenA = $null
if ($regA.Code -eq 201) {
  $rA = $regA.Body | ConvertFrom-Json
  $tokenA = $rA.token
  $userA = $rA.user
}
Check "Register retornou token" -not [string]::IsNullOrEmpty($tokenA)

$dup = Invoke-Json "POST" "/api/v1/auth/register" @{ name = "Duplicado"; email = $emailA; password = $passA }
Check "Register email duplicado (409)" $dup.Code -eq 409

$regB = Invoke-Json "POST" "/api/v1/auth/register" @{ name = "Usuario B"; email = $emailB; password = $passA }
Check "Register usuario B (201)" $regB.Code -eq 201 $regB.Body
$tokenB = $null; $userB = $null
if ($regB.Code -eq 201) {
  $rB = $regB.Body | ConvertFrom-Json
  $tokenB = $rB.token
  $userB = $rB.user
}

$loginA = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $emailA; password = $passA }
Check "Login A (200)" $loginA.Code -eq 200 $loginA.Body

# Rota protegida sem token -> 401
$noAuth = Invoke-Json "GET" "/api/v1/channels"
Check "Channels sem token (401)" $noAuth.Code -eq 401

# ---------------------------------------------------------------- 2. Canais
Write-Step "2. Canais (resto do fluxo)"
$hA = @{ Authorization = "Bearer $tokenA" }
$hB = @{ Authorization = "Bearer $tokenB" }

$lat = -23.5505; $lng = -46.6333
$ch = Invoke-Json "POST" "/api/v1/channels" @{ name = "Canal Teste"; latitude = $lat; longitude = $lng } $hA
Check "Criar canal (201)" $ch.Code -eq 201 $ch.Body
$channelId = $null
if ($ch.Code -eq 201) { $channelId = ($ch.Body | ConvertFrom-Json).id }
Check "Canal criado com id" -not [string]::IsNullOrEmpty($channelId)

if ($channelId) {
  $joinA = Invoke-Json "POST" "/api/v1/channels/$channelId/join" $null $hA
  Check "Join A (200)" $joinA.Code -eq 200 $joinA.Body

  $near = Invoke-Json "GET" "/api/v1/channels/nearby?lat=$lat&lng=$lng&radius=10" $null $hA
  $nearOk = $false
  if ($near.Code -eq 200) {
    try { $nearOk = @($near.Body | ConvertFrom-Json | Where-Object { $_.id -eq $channelId }).Count -gt 0 } catch { }
  }
  Check "Nearby retorna o canal (raio 10km)" $nearOk $near.Body

  $mine = Invoke-Json "GET" "/api/v1/channels" $null $hA
  $mineOk = $false
  if ($mine.Code -eq 200) {
    try { $mineOk = @($mine.Body | ConvertFrom-Json | Where-Object { $_.id -eq $channelId }).Count -gt 0 } catch { }
  }
  Check "Meus canais contem o canal" $mineOk $mine.Body

  $users1 = Invoke-Json "GET" "/api/v1/channels/$channelId/users" $null $hA
  $users1Ok = $false
  if ($users1.Code -eq 200) {
    try {
      $list = $users1.Body | ConvertFrom-Json
      $users1Ok = $null -ne $list -and $list.Count -eq 1 -and $list[0] -eq $userA.id
    } catch { }
  }
  Check "Usuarios do canal = [A]" $users1Ok $users1.Body

  $joinB = Invoke-Json "POST" "/api/v1/channels/$channelId/join" $null $hB
  Check "Join B (200)" $joinB.Code -eq 200 $joinB.Body

  $users2 = Invoke-Json "GET" "/api/v1/channels/$channelId/users" $null $hA
  $users2Ok = $false
  if ($users2.Code -eq 200) {
    try { $users2Ok = ($users2.Body | ConvertFrom-Json).Count -eq 2 } catch { }
  }
  Check "Usuarios do canal = [A,B]" $users2Ok $users2.Body

  $joinB2 = Invoke-Json "POST" "/api/v1/channels/$channelId/join" $null $hB
  Check "Join B repetido (200 - idempotente)" $joinB2.Code -eq 200 $joinB2.Body

  # ---------------------------------------------------------------- 3. Voz (WebSocket)
  Write-Step "3. Voz (WebSocket - relay entre 2 radio)"
  $wsBase = $BaseUrl -replace '^http', 'ws'
  $wsA = $null; $wsB = $null
  try {
    $wsA = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenA))"
    $wsB = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenB))"
    Check "WS: A e B conectados ao canal" ($null -ne $wsA -and $null -ne $wsB)

    $sampleB64 = "QUJD" # bytes de exemplo (3 bytes)
    $sendJson = ('{{"type":"audio","payload":{{"room":"{0}","audio":"{1}","is_ptt":true}}}}' -f $channelId, $sampleB64)
    $sent = $null
    try {
      Send-WsText $wsA $sendJson
      $sent = $true
    } catch { $sent = $false }
    Check "WS: A enviou pacote de audio" $sent

    $rxB = Receive-WsText $wsB
    $relayOk = $false
    if ($rxB) {
      try {
        $mB = $rxB | ConvertFrom-Json
        $relayOk = ($mB.type -eq "audio" -and $mB.payload.audio -eq $sampleB64 -and $mB.payload.is_ptt -eq $true)
      } catch { }
    }
    Check "WS: B recebeu o audio de A" $relayOk $rxB

    # Com a correcao, o servidor NAO ecoa para o remetente: a primeira
    # mensagem que A recebe deve ser o "stop" de B (sem o proprio audio).
    $stopJson = ('{{"type":"audio","payload":{{"room":"{0}","audio":"","is_ptt":false}}}}' -f $channelId)
    Send-WsText $wsB $stopJson | Out-Null
    $rxA = Receive-WsText $wsA
    $stopOk = $false
    if ($rxA) {
      try {
        $mA = $rxA | ConvertFrom-Json
        $stopOk = ($mA.type -eq "audio" -and -not $mA.payload.is_ptt -and $mA.payload.audio -ne $sampleB64)
      } catch { }
    }
    Check "WS: A recebeu o fim de B (sem eco proprio)" $stopOk $rxA
  }
  catch {
    Check "WS: fluxo executado sem erro" $false $_.Exception.Message
  }
  finally {
    if ($wsA) { Close-Ws $wsA }
    if ($wsB) { Close-Ws $wsB }
  }

  # ------------------------------------------------------------ 3b. voiceGuard (RF03)
  # Apenas um usuario fala por vez: 2o falante e rejeitado (nao ha relay) e o
  # disconnect de A libera o canal (releaseVoice) para B transmitir.
  Write-Step "3b. voiceGuard (RF03 - um falante por vez)"
  $wgA = $null; $wgA2 = $null; $wgB = $null
  try {
    $wgA = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenA))"
    $wgB = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenB))"
    Start-Sleep -Milliseconds 300 # servidor registra a room apos o upgrade

    $wgSample1 = "QUJDAA=="
    Send-WsText $wgA ('{{"type":"audio","payload":{{"room":"{0}","audio":"{1}","is_ptt":true}}}}' -f $channelId, $wgSample1)
    $wgRxB = Receive-WsText $wgB
    $wgStartOk = $false
    if ($wgRxB) {
      try {
        $m = $wgRxB | ConvertFrom-Json
        $wgStartOk = ($m.type -eq "audio" -and $m.payload.audio -eq $wgSample1 -and $m.payload.is_ptt -eq $true)
      } catch { }
    }
    Check "voiceGuard: B recebe audio de A (A assumiu o canal)" $wgStartOk $wgRxB

    # B tenta falar com A ainda transmitindo -> rejeitado: nada chega em A.
    $wgSample2 = "QUJCAg=="
    Send-WsText $wgB ('{{"type":"audio","payload":{{"room":"{0}","audio":"{1}","is_ptt":true}}}}' -f $channelId, $wgSample2)
    $wgRxA = Receive-WsText $wgA 2000
    Check "voiceGuard: 2o falante (B) e rejeitado" ($null -eq $wgRxA) "A recebeu: $wgRxA"

    # A desconecta -> releaseVoice libera o canal.
    Close-Ws $wgA
    $wgA = $null
    Start-Sleep -Milliseconds 700

    # Reconecta A apenas como ouvinte; B agora consegue transmitir.
    $wgA2 = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenA))"
    Start-Sleep -Milliseconds 300
    $wgSample3 = "QUJDQw=="
    Send-WsText $wgB ('{{"type":"audio","payload":{{"room":"{0}","audio":"{1}","is_ptt":true}}}}' -f $channelId, $wgSample3)
    $wgRxA2 = Receive-WsText $wgA2
    $wgReleaseOk = $false
    if ($wgRxA2 -and $userB) {
      try {
        $m2 = $wgRxA2 | ConvertFrom-Json
        $wgReleaseOk = ($m2.type -eq "audio" -and $m2.payload.is_ptt -eq $true -and $m2.payload.user_id -eq $userB.id)
      } catch { }
    }
    Check "voiceGuard: apos disconnect de A, B consegue transmitir" $wgReleaseOk $wgRxA2
  }
  catch {
    Check "voiceGuard: fluxo executado sem erro" $false $_.Exception.Message
  }
  finally {
    if ($wgA) { Close-Ws $wgA }
    if ($wgA2) { Close-Ws $wgA2 }
    if ($wgB) { Close-Ws $wgB }
  }

  # ------------------------------------------------------------ 3c. request_replay (RF06)
  # B recebe 3 chunks (sequencia N), desconecta; A envia 2 a mais (N+1, N+2);
  # B reconecta e pede replay desde N -> recebe os 2 perdidos + sentinel.
  Write-Step "3c. request_replay (RF06 - replay pos-reconexao)"
  $rpA = $null; $rpB = $null; $rpB2 = $null
  try {
    $rpA = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenA))"
    $rpB = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenB))"
    Start-Sleep -Milliseconds 300

    $rpChunks = @("QUJDQQ==", "QUJDQg==", "QUJDQw==")
    foreach ($chunk in $rpChunks) {
      Send-WsText $rpA ('{{"type":"audio","payload":{{"room":"{0}","audio":"{1}","is_ptt":true}}}}' -f $channelId, $chunk)
    }

    $rpSeqs = @()
    $rpLiveOk = $true
    for ($i = 0; $i -lt 3; $i++) {
      $rx = Receive-WsText $rpB
      if (-not $rx) { $rpLiveOk = $false; break }
      try {
        $m = $rx | ConvertFrom-Json
        if ($m.type -ne "audio" -or -not $m.payload.is_ptt -or -not $m.payload.sequence) { $rpLiveOk = $false; break }
        $rpSeqs += [int64]$m.payload.sequence
      } catch { $rpLiveOk = $false; break }
    }
    Check "RF06: B recebeu 3 chunks ao vivo com sequence" ($rpLiveOk -and $rpSeqs.Count -eq 3) "seqs=$($rpSeqs -join ',')"
    $rpLastSeq = if ($rpSeqs.Count -gt 0) { $rpSeqs[-1] } else { 0 }

    # B desconecta; A grava 2 chunks "perdidos" no buffer.
    Close-Ws $rpB
    $rpB = $null
    $rpMissed = @("QUJDRA==", "QUJDRAA=")
    foreach ($chunk in $rpMissed) {
      Send-WsText $rpA ('{{"type":"audio","payload":{{"room":"{0}","audio":"{1}","is_ptt":true}}}}' -f $channelId, $chunk)
    }

    # B reconecta (entra na room automaticamente) e pede replay.
    $rpB2 = New-Ws "$wsBase/ws/audio/?channel=$channelId&token=$([uri]::EscapeDataString($tokenB))"
    Send-WsText $rpB2 ('{{"type":"request_replay","payload":{{"room":"{0}","since_sequence":{1}}}}}' -f $channelId, $rpLastSeq)

    $rpGot = @()
    for ($i = 0; $i -lt 3; $i++) {
      $rx = Receive-WsText $rpB2 8000
      if (-not $rx) { break }
      try { $rpGot += ($rx | ConvertFrom-Json) } catch { }
    }

    $rpReplayed = @($rpGot | Where-Object { $_.payload.is_replayed -eq $true -and $_.payload.is_ptt -eq $true })
    $rpTerm = @($rpGot | Where-Object { $_.payload.sender -eq "replay" -and $_.payload.is_ptt -eq $false })
    $rpSeqOk = ($rpReplayed.Count -eq 2 -and
      [int64]$rpReplayed[0].payload.sequence -eq ($rpLastSeq + 1) -and
      [int64]$rpReplayed[1].payload.sequence -eq ($rpLastSeq + 2))
    Check "RF06: replay entregou os 2 chunks perdidos (N+1, N+2)" $rpSeqOk "last=$rpLastSeq got=$($rpGot.Count)"
    Check "RF06: replay encerrado por sentinel is_ptt=false (sender=replay)" ($rpTerm.Count -eq 1) "got=$($rpGot.Count)"
  }
  catch {
    Check "RF06: fluxo executado sem erro" $false $_.Exception.Message
  }
  finally {
    if ($rpA) { Close-Ws $rpA }
    if ($rpB) { Close-Ws $rpB }
    if ($rpB2) { Close-Ws $rpB2 }
  }

  $leaveB = Invoke-Json "POST" "/api/v1/channels/$channelId/leave" $null $hB
  Check "Leave B (200)" $leaveB.Code -eq 200 $leaveB.Body

  $users3 = Invoke-Json "GET" "/api/v1/channels/$channelId/users" $null $hA
  $users3Ok = $false
  if ($users3.Code -eq 200) {
    try { $users3Ok = ($users3.Body | ConvertFrom-Json).Count -eq 1 } catch { }
  }
  Check "Usuarios do canal = [A] apos leave" $users3Ok $users3.Body
}

# ---------------------------------------------------------------- 4. Perfil
Write-Step "4. Perfil"
$prof = Invoke-Json "GET" "/api/v1/user/profile" $null $hA
$profOk = $false
if ($prof.Code -eq 200) {
  try { $profOk = (($prof.Body | ConvertFrom-Json).email -eq $emailA) } catch { }
}
Check "GET perfil A" $profOk $prof.Body

$upd = Invoke-Json "PUT" "/api/v1/user/profile" @{ name = "Usuario A Editado" } $hA
Check "PUT perfil (renomear)" $upd.Code -eq 200 $upd.Body

# ---------------------------------------------------------------- 5. Senha
Write-Step "5. Troca de senha"
$cp = Invoke-Json "POST" "/api/v1/user/change-password" @{ old_password = $passA; new_password = $passNew } $hA
Check "Troca de senha correta (200)" $cp.Code -eq 200 $cp.Body

$loginOld = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $emailA; password = $passA }
Check "Login com senha antiga (401)" $loginOld.Code -eq 401

$loginNew = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $emailA; password = $passNew }
Check "Login com senha nova (200)" $loginNew.Code -eq 200 $loginNew.Body

# ---------------------------------------------------------------- 6. Refresh
Write-Step "6. Refresh token"
$refreshT = ($regA.Body | ConvertFrom-Json).refresh_token
$ref = Invoke-Json "POST" "/api/v1/auth/refresh" @{ refresh_token = $refreshT }
Check "Refresh token (200)" $ref.Code -eq 200 $ref.Body

# ---------------------------------------------------------------- 7. Admin (RF09)
# TC-RF09: metrics, lista sem password_hash, authz (401/403), self-action 403,
# suspensao invalida login e delete. Usa o token do bloco 0b.
Write-Step "7. Admin (RF09 - painel administrativo)"
if (-not $adminReady -or [string]::IsNullOrEmpty($tokenAd)) {
  Check "Admin: bloqueio executado" $false "bootstrap do admin falhou (ver bloco 0b)"
}
else {
  $hAd = @{ Authorization = "Bearer $tokenAd" }

  # 1. Dashboard com as 6 metricas
  $stats = Invoke-Json "GET" "/api/v1/admin/stats" $null $hAd
  $statsOk = $false
  if ($stats.Code -eq 200) {
    try {
      $s = $stats.Body | ConvertFrom-Json
      $statsOk = ($null -ne $s.users_total -and $null -ne $s.users_active -and
                  $null -ne $s.channels_total -and $null -ne $s.ws_connections -and
                  $null -ne $s.live_rooms -and $null -ne $s.transmitting_channels)
    } catch { }
  }
  Check "Admin: GET stats 200 com as 6 metricas" $statsOk $stats.Body

  # 2. Lista de usuarios: sem password_hash, com campos de localizacao
  $usersL = Invoke-Json "GET" "/api/v1/admin/users" $null $hAd
  $usersOk = $false
  if ($usersL.Code -eq 200) {
    $usersOk = (-not ($usersL.Body -match 'password')) -and
               ($usersL.Body -match '"last_lat"') -and
               ($usersL.Body -match '"last_lng"') -and
               ($usersL.Body -match '"last_location_at"')
  }
  Check "Admin: GET users 200 sem password_hash e com last_lat/lng" $usersOk $usersL.Body

  # 3. Authz: sem token -> 401; usuario comum -> 403
  $statsNoTok = Invoke-Json "GET" "/api/v1/admin/stats"
  Check "Admin: stats sem token -> 401" $statsNoTok.Code -eq 401 $statsNoTok.Code
  $statsComum = Invoke-Json "GET" "/api/v1/admin/stats" $null $hB
  Check "Admin: stats com token comum -> 403" $statsComum.Code -eq 403 $statsComum.Code

  # 4. Self-action bloqueado: admin nao pode nem suspender nem excluir a si
  $selfPatch = Invoke-Json "PATCH" "/api/v1/admin/users/$adminUserId" @{ is_active = $false } $hAd
  Check "Admin: PATCH de si mesmo (is_active) -> 403" $selfPatch.Code -eq 403 "$($selfPatch.Code) $($selfPatch.Body)"
  $selfDel = Invoke-Json "DELETE" "/api/v1/admin/users/$adminUserId" $null $hAd
  Check "Admin: DELETE de si mesmo -> 403" $selfDel.Code -eq 403 "$($selfDel.Code) $($selfDel.Body)"

  # 5. Suspender B invalida o proximo login (401)
  if ($userB) {
    $sus = Invoke-Json "PATCH" "/api/v1/admin/users/$($userB.id)" @{ is_active = $false } $hAd
    Check "Admin: suspender B -> 200" $sus.Code -eq 200 "$($sus.Code) $($sus.Body)"

    $loginSus = Invoke-Json "POST" "/api/v1/auth/login" @{ email = $emailB; password = $passA }
    Check "Admin: login de suspenso -> 401" $loginSus.Code -eq 401 "$($loginSus.Code) $($loginSus.Body)"

    # 6. Excluir B e some da lista
    $delB = Invoke-Json "DELETE" "/api/v1/admin/users/$($userB.id)" $null $hAd
    Check "Admin: DELETE B -> 200" $delB.Code -eq 200 "$($delB.Code) $($delB.Body)"

    $usersL2 = Invoke-Json "GET" "/api/v1/admin/users" $null $hAd
    $goneOk = $false
    if ($usersL2.Code -eq 200) { $goneOk = -not ($usersL2.Body -match [regex]::Escape($userB.id)) }
    Check "Admin: B nao aparece apos delete" $goneOk $usersL2.Body
  }
  else {
    Check "Admin: usuario B disponivel para suspensao/delete" $false "registro de B falhou"
  }
}

# ---------------------------------------------------------------- Resultado
Write-Step "RESULTADO"
Write-Host ("  Testes passaram : {0}" -f $script:Pass) -ForegroundColor Green
Write-Host ("  Testes falharam : {0}" -f $script:Fail) -ForegroundColor Red
if ($script:Fail -eq 0) {
  Write-Host "  >>> TUDO OK <<<" -ForegroundColor Green
  exit 0
}
else {
  Write-Host "  >>> REVISAR FALHAS <<<" -ForegroundColor Red
  exit 1
}