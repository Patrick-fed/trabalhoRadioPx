param(
  [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"
$script:Pass = 0
$script:Fail = 0

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

# ---------------------------------------------------------------- 0. Health
Write-Step "0. Servidor"
$health = Invoke-Json "GET" "/health"
if ($health.Code -ne 200) {
  $exe = "C:\RadioPx\backend\server.exe"
  if (Test-Path $exe) {
    Write-Host "Servidor parado. Iniciando $exe ..."
    Start-Process -FilePath $exe -WorkingDirectory (Split-Path $exe) -WindowStyle Hidden | Out-Null
    Start-Sleep -Seconds 2
    $health = Invoke-Json "GET" "/health"
  }
}
$healthOk = ($health.Code -eq 200 -and ($health.Body -match '"status"\s*:\s*"ok"'))
Check "Health check $BaseUrl/health" $healthOk $health.Body

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
$tokenB = $null
if ($regB.Code -eq 201) {
  $rB = $regB.Body | ConvertFrom-Json
  $tokenB = $rB.token
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
    $wsA = New-Ws "$wsBase/ws/audio?channel=$channelId&user_id=$($userA.id)"
    $wsB = New-Ws "$wsBase/ws/audio?channel=$channelId&user_id=$($rB.user.id)"
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