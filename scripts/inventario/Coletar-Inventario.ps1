<#
.SYNOPSIS
    Coletor de inventário do SIGE-TI. Lê a configuração da máquina (via CIM/WMI)
    e envia para o endpoint de ingestão do SIGE, que faz o upsert do item pelo
    número de série.

.DESCRIPTION
    Rode manualmente, por logon script ou por Tarefa Agendada (de preferência
    como SYSTEM). Reexecutar atualiza o mesmo item — não duplica.

    Segurança:
      - Use SEMPRE a URL HTTPS do SIGE. A chave e o inventário trafegam no corpo.
      - A chave (-ApiKey) é a INVENTARIO_API_KEY do servidor, dedicada e de
        alcance mínimo (só faz upsert de item). Não use a chave de admin/JWT.

.PARAMETER BaseUrl
    URL base do SIGE, ex.: https://sige.prefeitura.gov.br

.PARAMETER ApiKey
    Chave do coletor (INVENTARIO_API_KEY configurada no servidor).

.EXAMPLE
    .\Coletar-Inventario.ps1 -BaseUrl "https://sige.prefeitura.gov.br" -ApiKey "xxxx"
#>

[CmdletBinding()]
param(
    [string] $BaseUrl,
    [string] $ApiKey,
    [int]    $SetorId,
    [switch] $ListarSetores
)

# ============================================================================
# CONFIGURAÇÃO — preencha aqui para rodar sem passar parâmetros.
# Os parâmetros -BaseUrl / -ApiKey / -SetorId, se informados, têm prioridade.
# ============================================================================
$ConfigBaseUrl = "https://sige.prefeitura.gov.br"   # URL HTTPS do SIGE
$ConfigApiKey  = "COLE_A_INVENTARIO_API_KEY_AQUI"   # INVENTARIO_API_KEY do servidor
$ConfigSetorId = 0                                   # ID do departamento desta máquina (0 = não define)
# ============================================================================

if ([string]::IsNullOrWhiteSpace($BaseUrl)) { $BaseUrl = $ConfigBaseUrl }
if ([string]::IsNullOrWhiteSpace($ApiKey))  { $ApiKey  = $ConfigApiKey }
if ($SetorId -le 0) { $SetorId = $ConfigSetorId }

if ([string]::IsNullOrWhiteSpace($BaseUrl) -or [string]::IsNullOrWhiteSpace($ApiKey) -or
    $ApiKey -eq "COLE_A_INVENTARIO_API_KEY_AQUI") {
    Write-Error "Configure \$ConfigBaseUrl e \$ConfigApiKey no topo do script (ou use -BaseUrl / -ApiKey)."
    exit 2
}

$ErrorActionPreference = "Stop"
# Garante TLS 1.2 em Windows/PowerShell antigos.
try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 } catch {}

# Modo descoberta: lista os departamentos (id/nome/sigla) para preencher o
# -SetorId / $ConfigSetorId. Rode UMA vez no PC do admin.
if ($ListarSetores) {
    try {
        $uri = ($BaseUrl.TrimEnd("/")) + "/api/v1/inventario/setores"
        $setores = Invoke-RestMethod -Uri $uri -Method Get -Headers @{ "X-API-Key" = $ApiKey }
        $setores | Select-Object id, sigla, nome | Format-Table -AutoSize
        exit 0
    }
    catch {
        Write-Error "Falha ao listar setores: $($_.Exception.Message)"
        exit 1
    }
}

function Get-Especificacao {
    param([string] $Nome, [string] $Valor)
    $v = ("" + $Valor).Trim()
    if ([string]::IsNullOrWhiteSpace($v)) { return $null }
    return [pscustomobject]@{ nome = $Nome; valor = $v }
}

try {
    $bios     = Get-CimInstance -ClassName Win32_BIOS
    $cs       = Get-CimInstance -ClassName Win32_ComputerSystem
    $os       = Get-CimInstance -ClassName Win32_OperatingSystem
    $cpu      = Get-CimInstance -ClassName Win32_Processor | Select-Object -First 1
    $enclos   = Get-CimInstance -ClassName Win32_SystemEnclosure | Select-Object -First 1
    $discos   = Get-CimInstance -ClassName Win32_DiskDrive
    $nics     = Get-CimInstance -ClassName Win32_NetworkAdapterConfiguration -Filter "IPEnabled = True"

    # Desktop x Notebook pelo tipo de chassi (8,9,10,11,12,14,18,21 = portáteis).
    $portateis = 8, 9, 10, 11, 12, 14, 18, 21
    $categoria = "Desktop"
    if ($enclos -and $enclos.ChassisTypes) {
        foreach ($t in $enclos.ChassisTypes) {
            if ($portateis -contains [int]$t) { $categoria = "Notebook"; break }
        }
    }

    # Memória RAM total (GB).
    $ramGB = [math]::Round(($cs.TotalPhysicalMemory / 1GB), 1)

    # Armazenamento: soma dos discos físicos (GB), com modelos.
    $discoGB = 0
    $discoModelos = @()
    foreach ($d in $discos) {
        if ($d.Size) { $discoGB += [math]::Round(($d.Size / 1GB), 0) }
        if ($d.Model) { $discoModelos += $d.Model.Trim() }
    }
    $discoTexto = ""
    if ($discoGB -gt 0) {
        $discoTexto = "$discoGB GB"
        if ($discoModelos.Count -gt 0) {
            $discoTexto += " (" + (($discoModelos | Select-Object -Unique) -join ", ") + ")"
        }
    }

    # Primeiro MAC com IP.
    $mac = ""
    if ($nics) {
        $nic = $nics | Where-Object { $_.MACAddress } | Select-Object -First 1
        if ($nic) { $mac = $nic.MACAddress }
    }

    $serie = ("" + $bios.SerialNumber).Trim()
    # Séries "lixo" comuns de placas sem SMBIOS preenchido.
    $lixo = @("", "To be filled by O.E.M.", "Default string", "System Serial Number", "None", "0")
    if ($lixo -contains $serie) { $serie = "" }

    $especificacoes = @(
        Get-Especificacao "Processador"        $cpu.Name
        Get-Especificacao "Núcleos"            ("" + $cpu.NumberOfCores)
        Get-Especificacao "Memória RAM"        ("$ramGB GB")
        Get-Especificacao "Armazenamento"      $discoTexto
        Get-Especificacao "Sistema Operacional" $os.Caption
        Get-Especificacao "Versão do SO"       $os.Version
        Get-Especificacao "Endereço MAC"       $mac
        Get-Especificacao "Usuário"            $cs.UserName
    ) | Where-Object { $_ -ne $null }

    $payload = [ordered]@{
        hostname       = $env:COMPUTERNAME
        numero_serie   = $serie
        fabricante     = ("" + $cs.Manufacturer).Trim()
        modelo         = ("" + $cs.Model).Trim()
        categoria      = $categoria
        especificacoes = $especificacoes
    }
    # Define o local (departamento) só quando houver um ID configurado.
    if ($SetorId -gt 0) { $payload.setor_id = [int]$SetorId }

    $json = ([pscustomobject]$payload) | ConvertTo-Json -Depth 5
    $uri  = ($BaseUrl.TrimEnd("/")) + "/api/v1/inventario/coleta"

    $resp = Invoke-RestMethod -Uri $uri -Method Post -Body $json `
        -ContentType "application/json; charset=utf-8" `
        -Headers @{ "X-API-Key" = $ApiKey }

    if ($resp.criado) {
        Write-Output "OK: item criado (#$($resp.item.id)) - $($resp.item.descricao)"
    } else {
        Write-Output "OK: item atualizado (#$($resp.item.id)) - $($resp.item.descricao)"
    }
    exit 0
}
catch {
    Write-Error "Falha na coleta de inventário: $($_.Exception.Message)"
    exit 1
}
