# SIGE-TI — Frontend

Interface web do **Sistema de Gestão de Estoque de T.I. (SIGE-TI)**, construída em
React + Vite + TypeScript, com Tailwind CSS e componentes no padrão shadcn/ui.
Toda a interface está em **português do Brasil**.

Consome a API REST do backend Go (`/api/v1`). Veja o `README.md` do backend para
subir a API.

---

## Stack

- **React 18 + TypeScript** (Vite)
- **Tailwind CSS** + componentes shadcn/ui (Radix UI)
- **React Router** com rotas protegidas por perfil
- **Axios** centralizado com interceptors de JWT e tratamento global de 401
- **lucide-react** para ícones

## Pré-requisitos

- Node.js 18+ (testado em Node 20/22)
- Backend SIGE-TI em execução (por padrão em `http://localhost:8080`)

## Instalação

```bash
cd frontend
cp .env.example .env   # ajuste se necessário
npm install
```

## Configuração (variáveis de ambiente)

| Variável             | Descrição                                                                 | Padrão                  |
| -------------------- | ------------------------------------------------------------------------- | ----------------------- |
| `VITE_API_BASE_URL`  | URL base da API. **Vazio em dev** (usa o proxy do Vite para `/api`).      | _(vazio)_               |
| `VITE_PROXY_TARGET`  | Alvo do proxy de desenvolvimento (apenas em dev).                          | `http://localhost:8080` |

**Em desenvolvimento**, deixe `VITE_API_BASE_URL` vazio: o Vite faz proxy de
`/api` para o backend (definido em `VITE_PROXY_TARGET`), evitando problemas de
CORS no navegador.

**Em produção**, defina `VITE_API_BASE_URL` com a URL pública da API
(ex.: `https://sige-ti.prefeitura.gov.br`). O frontend chamará
`${VITE_API_BASE_URL}/api/v1`.

## Execução em desenvolvimento

```bash
npm run dev
```

Acesse `http://localhost:5173`.

### Credenciais de teste (geradas pelo seed do backend)

Disponíveis em desenvolvimento com o seed demo (`SEED_DEMO=true` no backend):

| Perfil        | E-mail                     | Senha         |
| ------------- | -------------------------- | ------------- |
| Administrador | `admin@sige-ti.local`      | `admin123`    |
| Operador      | `operador@sige-ti.local`   | `operador123` |

> Em produção, o admin é criado com `ADMIN_EMAIL`/`ADMIN_SENHA` do backend e o
> operador de exemplo não existe.

## Build de produção

```bash
npm run build      # gera ./dist
npm run preview    # serve o build localmente para conferência
```

Os arquivos estáticos ficam em `dist/` e podem ser servidos por qualquer
servidor web (Nginx, Apache) ou CDN.

### Deploy em rede interna (LAN) via Docker

O modo recomendado para uso interno é o `docker compose` da raiz do projeto: o
frontend é servido por **Nginx na porta 80** do servidor e faz **proxy de
`/api`** para o backend (sem CORS, mesmo host). Não é preciso definir
`VITE_API_BASE_URL` (fica vazio e usa o proxy).

```bash
cd ..                      # raiz do projeto
cp .env.example .env       # defina JWT_SECRET e ADMIN_SENHA
docker compose up -d --build
```

Acesse de qualquer máquina da rede em `http://IP-do-servidor`
(ex.: `http://192.168.0.50`). Ajuste `CORS_ALLOWED_ORIGINS` no `.env` para esse
endereço.

---

## Estrutura de pastas

```
src/
├── components/          # Componentes reutilizáveis
│   ├── ui/              # Primitivos no padrão shadcn/ui (button, input, dialog, ...)
│   ├── layout/          # AppShell, Sidebar, Header
│   ├── FormField.tsx, ConfirmDialog.tsx, PageHeader.tsx, ...
│   ├── MovimentacaoForm.tsx, TermoForm.tsx, HistoricoTabela.tsx
├── contexts/            # AuthContext (sessão/JWT)
├── hooks/               # useReferencias (categorias/setores/servidores)
├── lib/                 # utils, format (datas/moeda PT-BR), rotulos (enums)
├── pages/               # Telas (Login, Dashboard, Itens, Movimentações, ...)
│   ├── itens/           # Lista, detalhe e formulário de itens
│   └── cadastros/       # Categorias, Setores, Servidores, Usuários
├── routes/              # ProtectedRoute (autenticação + perfil)
├── services/api/        # Camada de API (client axios + módulos por recurso)
├── types/               # Tipos TS espelhando os DTOs do backend
├── App.tsx              # Roteamento
└── main.tsx             # Entrypoint (Providers)
```

## Perfis de acesso (RBAC no frontend)

O frontend respeita os mesmos níveis de acesso do backend:

- **Administrador**: acesso total, incluindo gestão de **Usuários** (criar,
  **redefinir senha**, **ativar/desativar**), a escrita de **Categorias** e
  **Setores**, e a **exclusão de itens** cadastrados por engano.
- **Operador**: opera o dia a dia (itens, movimentações, servidores, termos,
  relatórios). Vê categorias/setores, mas não os altera; não acessa Usuários e
  não exclui itens (apenas cadastra/edita).

### Gestão de contas e correção de cadastros

- Na tela **Usuários** (admin), além de criar, é possível **redefinir a senha**
  de um usuário e **ativar/desativar** a conta. Não há exclusão física: usuário
  inativo simplesmente não consegue mais entrar. O sistema impede desativar o
  último administrador ativo.
- Na lista de **Itens**, o admin tem a ação **Excluir** (com confirmação) para
  remover um item cadastrado por engano. Itens com histórico de movimentações ou
  termos **não podem ser excluídos** — nesses casos, use a **baixa patrimonial**.

A rota `/usuarios` é protegida por `ProtectedRoute somenteAdmin`. Botões de
escrita exclusivos de administrador são ocultados conforme o perfil. A
autorização **definitiva** é sempre do backend — o frontend apenas melhora a
experiência.

## Validações

As validações ocorrem **nos dois lados**:

- No frontend, antes de enviar (campos obrigatórios, quantidades, saída maior
  que o estoque, etc.), com mensagens claras em português.
- No backend (autoritativo). Erros de validação por campo (HTTP 422) são
  exibidos junto a cada campo; conflitos de regra de negócio (HTTP 409, ex.:
  estoque insuficiente) e demais erros são exibidos como notificação (toast).

## Notas de LGPD

A interface coleta e exibe apenas os dados estritamente necessários do servidor
responsável (**nome e matrícula**). Não há coleta de dados pessoais sensíveis. O
acesso é protegido por autenticação JWT e segregado por perfil.

## Integração com a API

- O token JWT é armazenado em `localStorage` (`sige-ti.token`) e injetado no
  header `Authorization: Bearer ...` por um interceptor do axios.
- Respostas `401` limpam o token e redirecionam para o login automaticamente.
- Downloads (CSV/PDF de relatórios e termos) são feitos via `responseType: blob`
  e disparados no navegador com o nome de arquivo apropriado.
