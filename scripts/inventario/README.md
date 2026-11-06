# Coletor de inventário — SIGE-TI

Lê a configuração de cada computador (via CIM/WMI) e envia ao SIGE, que faz
**upsert do item pelo número de série** (reexecutar atualiza, não duplica).

## O que é coletado
- Fabricante, modelo, número de série (BIOS)
- Categoria automática: **Desktop** ou **Notebook** (pelo tipo de chassi)
- Specs (viram campos customizados da categoria): Processador, Núcleos,
  Memória RAM, Armazenamento, Sistema Operacional, Versão do SO, MAC, Usuário

Os campos customizados são **criados sozinhos** na categoria na primeira coleta.
Setor, responsável, nº de patrimônio e estado **nunca** são sobrescritos — ficam
sob controle do operador no SIGE.

## Pré-requisitos no servidor
No `.env` do backend, configure a chave dedicada e rebuild:

```
INVENTARIO_API_KEY=<gere: openssl rand -hex 32>
```

Vazio = coleta desabilitada (endpoint responde 503).

## Segurança
- Use **sempre** a URL **HTTPS** do SIGE (a chave vai no header).
- A `INVENTARIO_API_KEY` é dedicada e de alcance mínimo: o endpoint só faz
  upsert de item. Mesmo exposta, não escala acesso.
- Em logon script de GPO (SYSVOL), qualquer usuário do domínio lê o `.ps1` e a
  chave. Prefira **Tarefa Agendada como SYSTEM** (abaixo) e rotacione a chave.

## Configuração embutida (sem passar parâmetros)
No topo do `.ps1`, preencha:

```powershell
$ConfigBaseUrl = "https://sige.prefeitura.gov.br"
$ConfigApiKey  = "<INVENTARIO_API_KEY>"
```

Assim roda só com `.\Coletar-Inventario.ps1`. Os parâmetros `-BaseUrl` / `-ApiKey`,
se informados, ainda têm prioridade sobre o que está no arquivo.

> **Atenção:** com a chave dentro do arquivo, quem ler o `.ps1` tem a chave.
> Mantenha o `.ps1` numa pasta com ACL restrita (ex.: `C:\SIGE`, só Administradores
> + SYSTEM) e **não** o coloque no SYSVOL de um logon script. Rotacione a chave se
> vazar. Como a chave só permite upsert de inventário, o risco é limitado.

## Local da máquina (departamento) por ID
Defina o departamento da máquina pelo **ID** do setor — sem digitar nome.

1. Descubra os IDs (rode uma vez, no PC do admin):
   ```powershell
   .\Coletar-Inventario.ps1 -ListarSetores
   ```
   Imprime `id / sigla / nome` de todos os departamentos.
2. Configure o ID no script (por departamento) ou passe na linha de comando:
   ```powershell
   $ConfigSetorId = 12        # no topo do .ps1
   # ou
   .\Coletar-Inventario.ps1 -SetorId 12
   ```
`0` (padrão) = não mexe no setor. Com um ID, a coleta **define** o local do item
(inclusive sobrescreve numa recoleta — o script vira a fonte da verdade do local,
então use um ID por departamento). `-SetorId` tem prioridade sobre `$ConfigSetorId`.

## Uso manual (teste)
```powershell
# Com config embutida:
.\Coletar-Inventario.ps1
# Ou passando na linha de comando:
.\Coletar-Inventario.ps1 -BaseUrl "https://sige.prefeitura.gov.br" -ApiKey "<chave>"
```

## Deploy recomendado — Tarefa Agendada como SYSTEM
Roda sozinha no boot/diariamente, sem expor a chave no logon do usuário.
Copie o `.ps1` para `C:\SIGE\` em cada máquina (ou via GPO de arquivos) e crie a
tarefa (uma vez, pode ir por GPO Preferences → Scheduled Tasks):

```cmd
schtasks /Create /TN "SIGE-TI Inventario" /RU SYSTEM /SC DAILY /ST 09:00 /RL HIGHEST /F ^
  /TR "powershell.exe -NoProfile -ExecutionPolicy Bypass -File C:\SIGE\Coletar-Inventario.ps1 -BaseUrl https://sige.prefeitura.gov.br -ApiKey <chave>"
```

Para coletar também no logon, adicione uma 2ª tarefa com `/SC ONLOGON`.

## Deploy alternativo — Logon Script (GPO)
`Configuração do Usuário → Políticas → Scripts → Logon` apontando para o `.ps1`
com os parâmetros. Mais simples, porém expõe a chave no SYSVOL (ver Segurança).

## Código de saída
`0` = sucesso · `1` = falha (mensagem no stderr / Visualizador de Eventos).
