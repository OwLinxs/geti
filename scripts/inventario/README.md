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

## Uso manual (teste)
```powershell
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
