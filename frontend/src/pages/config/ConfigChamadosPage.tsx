import * as React from "react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { FormField } from "@/components/FormField";
import { Card, CardContent } from "@/components/ui/card";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { configChamadosApi } from "@/services/api";
import { mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { cn } from "@/lib/utils";
import CategoriasChamado from "./CategoriasChamado";
import RespostasRapidas from "./RespostasRapidas";
import type { ConfigChamados } from "@/types";

type Aba = "notificacoes" | "sla" | "categorias" | "respostas";

export default function ConfigChamadosPage() {
  const [aba, setAba] = React.useState<Aba>("notificacoes");

  return (
    <div>
      <PageHeader
        titulo="Chamados — Configurações"
        descricao="Parâmetros do módulo de chamados: notificações e categorias."
      />

      <div className="mb-4 flex flex-wrap gap-2">
        {(
          [
            ["notificacoes", "Notificações"],
            ["sla", "SLA / Prazos"],
            ["categorias", "Categorias"],
            ["respostas", "Respostas rápidas"],
          ] as [Aba, string][]
        ).map(([v, rot]) => (
          <button
            key={v}
            onClick={() => setAba(v)}
            className={cn(
              "rounded-md border px-3 py-1.5 text-sm font-medium transition-colors",
              aba === v
                ? "border-primary bg-primary text-primary-foreground"
                : "border-border bg-card text-foreground hover:bg-accent"
            )}
          >
            {rot}
          </button>
        ))}
      </div>

      {aba === "notificacoes" && <NotificacoesConfig />}
      {aba === "sla" && <SLAConfig />}
      {aba === "categorias" && <CategoriasChamado embutido />}
      {aba === "respostas" && <RespostasRapidas />}
    </div>
  );
}

function SLAConfig() {
  const { toast } = useToast();
  const [cfg, setCfg] = React.useState<ConfigChamados | null>(null);
  const [carregando, setCarregando] = React.useState(true);
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    configChamadosApi
      .obter()
      .then(setCfg)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  async function salvar() {
    if (!cfg) return;
    setSalvando(true);
    try {
      const at = await configChamadosApi.salvar(cfg);
      setCfg(at);
      toast({ titulo: "Prazos salvos.", variant: "success" });
    } catch (err) {
      toast({
        titulo: "Não foi possível salvar",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setSalvando(false);
    }
  }

  if (carregando) return <CarregandoTela />;
  if (!cfg) return null;

  const prioridades: {
    rotulo: string;
    resp: keyof ConfigChamados;
    resol: keyof ConfigChamados;
  }[] = [
    { rotulo: "Alta", resp: "sla_resposta_alta_h", resol: "sla_resolucao_alta_h" },
    {
      rotulo: "Normal",
      resp: "sla_resposta_normal_h",
      resol: "sla_resolucao_normal_h",
    },
    {
      rotulo: "Baixa",
      resp: "sla_resposta_baixa_h",
      resol: "sla_resolucao_baixa_h",
    },
  ];

  return (
    <Card>
      <CardContent className="space-y-4 p-4">
        <p className="text-sm text-muted-foreground">
          Prazos em <strong>horas</strong> por prioridade. Resposta = 1ª resposta
          da equipe; Resolução = conclusão. Use <strong>0</strong> para não cobrar
          prazo.
        </p>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-muted-foreground">
                <th className="py-2">Prioridade</th>
                <th className="py-2">Resposta (h)</th>
                <th className="py-2">Resolução (h)</th>
              </tr>
            </thead>
            <tbody>
              {prioridades.map((p) => (
                <tr key={p.rotulo} className="border-b border-border/60">
                  <td className="py-2 font-medium">{p.rotulo}</td>
                  <td className="py-2 pr-3">
                    <Input
                      type="number"
                      min={0}
                      className="w-24"
                      value={String(cfg[p.resp] as number)}
                      onChange={(e) =>
                        setCfg({ ...cfg, [p.resp]: Number(e.target.value) })
                      }
                    />
                  </td>
                  <td className="py-2">
                    <Input
                      type="number"
                      min={0}
                      className="w-24"
                      value={String(cfg[p.resol] as number)}
                      onChange={(e) =>
                        setCfg({ ...cfg, [p.resol]: Number(e.target.value) })
                      }
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="flex justify-end">
          <Button onClick={salvar} disabled={salvando}>
            {salvando ? <Spinner className="h-4 w-4" /> : "Salvar"}
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Vale para chamados abertos a partir da alteração (o prazo é fixado na
          abertura).
        </p>
      </CardContent>
    </Card>
  );
}

function NotificacoesConfig() {
  const { toast } = useToast();
  const [cfg, setCfg] = React.useState<ConfigChamados | null>(null);
  const [carregando, setCarregando] = React.useState(true);
  const [salvando, setSalvando] = React.useState(false);

  React.useEffect(() => {
    configChamadosApi
      .obter()
      .then(setCfg)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar configurações",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  async function salvar() {
    if (!cfg) return;
    setSalvando(true);
    try {
      const atualizado = await configChamadosApi.salvar(cfg);
      setCfg(atualizado);
      toast({ titulo: "Configurações salvas.", variant: "success" });
    } catch (err) {
      toast({
        titulo: "Não foi possível salvar",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setSalvando(false);
    }
  }

  if (carregando) return <CarregandoTela />;
  if (!cfg) return null;

  type ChaveBool =
    | "notificar_equipe_novo_chamado"
    | "notificar_solicitante_resposta"
    | "notificar_solicitante_status";
  const linhas: { chave: ChaveBool; titulo: string; desc: string }[] =
    [
      {
        chave: "notificar_equipe_novo_chamado",
        titulo: "Avisar a equipe em novo chamado",
        desc: "Notifica administradores e operadores quando um chamado é aberto.",
      },
      {
        chave: "notificar_solicitante_resposta",
        titulo: "Avisar o solicitante em nova resposta",
        desc: "Notifica quem abriu o chamado quando a equipe responde.",
      },
      {
        chave: "notificar_solicitante_status",
        titulo: "Avisar o solicitante em mudança de status",
        desc: "Notifica quem abriu o chamado quando o status muda.",
      },
    ];

  return (
    <Card>
      <CardContent className="space-y-3 p-4">
        {linhas.map((l) => (
          <div
            key={l.chave}
            className="flex items-center justify-between gap-4 rounded-md border border-border px-3 py-2"
          >
            <div>
              <p className="text-sm font-medium">{l.titulo}</p>
              <p className="text-xs text-muted-foreground">{l.desc}</p>
            </div>
            <Switch
              checked={cfg[l.chave]}
              onCheckedChange={(v) => setCfg({ ...cfg, [l.chave]: v })}
            />
          </div>
        ))}
        {/* Envio por e-mail (SMTP). */}
        <div className="mt-4 space-y-3 border-t border-border pt-4">
          <div className="flex items-center justify-between gap-4 rounded-md border border-border px-3 py-2">
            <div>
              <p className="text-sm font-medium">Enviar também por e-mail</p>
              <p className="text-xs text-muted-foreground">
                Além do sino, envia as notificações por e-mail (requer SMTP).
              </p>
            </div>
            <Switch
              checked={cfg.email_ativo}
              onCheckedChange={(v) => setCfg({ ...cfg, email_ativo: v })}
            />
          </div>

          {cfg.email_ativo && (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <FormField label="Servidor SMTP (host)">
                <Input
                  value={cfg.smtp_host}
                  onChange={(e) => setCfg({ ...cfg, smtp_host: e.target.value })}
                  placeholder="smtp.gmail.com"
                />
              </FormField>
              <FormField label="Porta">
                <Input
                  type="number"
                  value={String(cfg.smtp_porta || 0)}
                  onChange={(e) =>
                    setCfg({ ...cfg, smtp_porta: Number(e.target.value) })
                  }
                  placeholder="587"
                />
              </FormField>
              <FormField label="Usuário">
                <Input
                  value={cfg.smtp_usuario}
                  onChange={(e) =>
                    setCfg({ ...cfg, smtp_usuario: e.target.value })
                  }
                  placeholder="usuario@dominio.gov.br"
                />
              </FormField>
              <FormField label="Senha">
                <Input
                  type="password"
                  value={cfg.smtp_senha}
                  onChange={(e) =>
                    setCfg({ ...cfg, smtp_senha: e.target.value })
                  }
                  placeholder="(inalterada)"
                />
              </FormField>
              <FormField label="Remetente (From)" className="sm:col-span-2">
                <Input
                  value={cfg.smtp_remetente}
                  onChange={(e) =>
                    setCfg({ ...cfg, smtp_remetente: e.target.value })
                  }
                  placeholder="T.I. Prefeitura <ti@dominio.gov.br>"
                />
              </FormField>
            </div>
          )}
        </div>

        <div className="flex justify-end pt-2">
          <Button onClick={salvar} disabled={salvando}>
            {salvando ? <Spinner className="h-4 w-4" /> : "Salvar"}
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          Notificações in-app (sino) sempre ativas. E-mail usa STARTTLS (porta
          587). A senha nunca é exibida; deixe em branco para manter a atual.
        </p>
      </CardContent>
    </Card>
  );
}
