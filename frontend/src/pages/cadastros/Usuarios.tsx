import * as React from "react";
import { Plus, KeyRound, UserCheck, UserX } from "lucide-react";
import { PageHeader } from "@/components/PageHeader";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FormField } from "@/components/FormField";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { EstadoVazio } from "@/components/EstadoVazio";
import { CarregandoTela, Spinner } from "@/components/ui/spinner";
import { usuariosApi } from "@/services/api";
import { camposInvalidos, mensagemErro } from "@/services/api/client";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/contexts/AuthContext";
import { PERFIS, rotuloPerfil } from "@/lib/rotulos";
import type { Perfil, Usuario, UsuarioPayload } from "@/types";

const VAZIO: UsuarioPayload = {
  nome: "",
  email: "",
  senha: "",
  perfil: "operador",
  ativo: true,
};

export default function Usuarios() {
  const { toast } = useToast();
  const { usuario: usuarioAtual } = useAuth();
  const [lista, setLista] = React.useState<Usuario[]>([]);
  const [carregando, setCarregando] = React.useState(true);

  // Criação.
  const [modalAberto, setModalAberto] = React.useState(false);
  const [form, setForm] = React.useState<UsuarioPayload>(VAZIO);
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [salvando, setSalvando] = React.useState(false);

  // Reset de senha.
  const [resetAlvo, setResetAlvo] = React.useState<Usuario | null>(null);
  const [novaSenha, setNovaSenha] = React.useState("");
  const [erroSenha, setErroSenha] = React.useState<string | undefined>();
  const [redefinindo, setRedefinindo] = React.useState(false);

  // Ativar/desativar.
  const [ativoAlvo, setAtivoAlvo] = React.useState<Usuario | null>(null);
  const [processandoAtivo, setProcessandoAtivo] = React.useState(false);

  const carregar = React.useCallback(() => {
    setCarregando(true);
    usuariosApi
      .listar()
      .then(setLista)
      .catch((err) =>
        toast({
          titulo: "Erro ao carregar usuários",
          descricao: mensagemErro(err),
          variant: "destructive",
        })
      )
      .finally(() => setCarregando(false));
  }, [toast]);

  React.useEffect(carregar, [carregar]);

  function abrir() {
    setForm(VAZIO);
    setErros({});
    setModalAberto(true);
  }

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    const novosErros: Record<string, string> = {};
    if (!form.nome.trim()) novosErros.nome = "Informe o nome.";
    if (!form.email.trim()) novosErros.email = "Informe o e-mail.";
    if (form.senha.length < 6)
      novosErros.senha = "A senha deve ter ao menos 6 caracteres.";
    if (Object.keys(novosErros).length) {
      setErros(novosErros);
      return;
    }
    setErros({});
    setSalvando(true);
    try {
      await usuariosApi.criar(form);
      toast({ titulo: "Usuário criado.", variant: "success" });
      setModalAberto(false);
      carregar();
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else
        toast({
          titulo: "Não foi possível criar o usuário",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setSalvando(false);
    }
  }

  function abrirReset(u: Usuario) {
    setResetAlvo(u);
    setNovaSenha("");
    setErroSenha(undefined);
  }

  async function confirmarReset(e: React.FormEvent) {
    e.preventDefault();
    if (!resetAlvo) return;
    if (novaSenha.length < 6) {
      setErroSenha("A nova senha deve ter ao menos 6 caracteres.");
      return;
    }
    setRedefinindo(true);
    try {
      await usuariosApi.redefinirSenha(resetAlvo.id, novaSenha);
      toast({
        titulo: "Senha redefinida.",
        descricao: `Informe a nova senha a ${resetAlvo.nome}.`,
        variant: "success",
      });
      setResetAlvo(null);
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos?.senha) setErroSenha(campos.senha);
      else
        toast({
          titulo: "Não foi possível redefinir a senha",
          descricao: mensagemErro(err),
          variant: "destructive",
        });
    } finally {
      setRedefinindo(false);
    }
  }

  async function confirmarAtivo() {
    if (!ativoAlvo) return;
    const novoAtivo = !ativoAlvo.ativo;
    setProcessandoAtivo(true);
    try {
      await usuariosApi.definirAtivo(ativoAlvo.id, novoAtivo);
      toast({
        titulo: novoAtivo ? "Usuário ativado." : "Usuário desativado.",
        variant: "success",
      });
      setAtivoAlvo(null);
      carregar();
    } catch (err) {
      toast({
        titulo: "Não foi possível alterar a situação",
        descricao: mensagemErro(err),
        variant: "destructive",
      });
    } finally {
      setProcessandoAtivo(false);
    }
  }

  return (
    <div>
      <PageHeader
        titulo="Usuários"
        descricao="Contas de acesso ao sistema (administradores e operadores)."
        acao={
          <Button onClick={abrir}>
            <Plus className="h-4 w-4" /> Novo usuário
          </Button>
        }
      />

      <Card>
        <CardContent className="p-0">
          {carregando ? (
            <CarregandoTela />
          ) : lista.length === 0 ? (
            <EstadoVazio />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nome</TableHead>
                  <TableHead>E-mail</TableHead>
                  <TableHead>Perfil</TableHead>
                  <TableHead>Situação</TableHead>
                  <TableHead className="w-48 text-right">Ações</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {lista.map((u) => {
                  const ehProprio = u.id === usuarioAtual?.id;
                  return (
                    <TableRow key={u.id}>
                      <TableCell className="font-medium">
                        {u.nome}
                        {ehProprio && (
                          <span className="ml-2 text-xs text-muted-foreground">
                            (você)
                          </span>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {u.email}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            u.perfil === "administrador"
                              ? "default"
                              : "secondary"
                          }
                        >
                          {rotuloPerfil(u.perfil)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {u.ativo ? (
                          <Badge variant="success">Ativo</Badge>
                        ) : (
                          <Badge variant="muted">Inativo</Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => abrirReset(u)}
                            aria-label="Redefinir senha"
                            title="Redefinir senha"
                          >
                            <KeyRound className="h-4 w-4" /> Senha
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={ehProprio}
                            onClick={() => setAtivoAlvo(u)}
                            aria-label={u.ativo ? "Desativar" : "Ativar"}
                            title={
                              ehProprio
                                ? "Você não pode alterar a própria situação"
                                : u.ativo
                                ? "Desativar usuário"
                                : "Ativar usuário"
                            }
                          >
                            {u.ativo ? (
                              <>
                                <UserX className="h-4 w-4" /> Desativar
                              </>
                            ) : (
                              <>
                                <UserCheck className="h-4 w-4" /> Ativar
                              </>
                            )}
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Criação de usuário */}
      <Dialog open={modalAberto} onOpenChange={setModalAberto}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Novo usuário</DialogTitle>
          </DialogHeader>
          <form onSubmit={salvar} className="space-y-4" noValidate>
            <FormField label="Nome" htmlFor="nome" obrigatorio erro={erros.nome}>
              <Input
                id="nome"
                value={form.nome}
                onChange={(e) => setForm({ ...form, nome: e.target.value })}
              />
            </FormField>
            <FormField
              label="E-mail"
              htmlFor="email"
              obrigatorio
              erro={erros.email}
            >
              <Input
                id="email"
                type="email"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                placeholder="usuario@dominio.gov.br"
              />
            </FormField>
            <FormField
              label="Senha"
              htmlFor="senha"
              obrigatorio
              erro={erros.senha}
            >
              <Input
                id="senha"
                type="password"
                value={form.senha}
                onChange={(e) => setForm({ ...form, senha: e.target.value })}
                placeholder="Mínimo de 6 caracteres"
              />
            </FormField>
            <FormField label="Perfil" obrigatorio erro={erros.perfil}>
              <Select
                value={form.perfil}
                onValueChange={(v) => setForm({ ...form, perfil: v as Perfil })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PERFIS.map((p) => (
                    <SelectItem key={p.valor} value={p.valor}>
                      {p.rotulo}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <DialogFooter className="gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setModalAberto(false)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={salvando}>
                {salvando ? <Spinner className="h-4 w-4" /> : "Salvar"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Redefinir senha */}
      <Dialog open={!!resetAlvo} onOpenChange={(o) => !o && setResetAlvo(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Redefinir senha de {resetAlvo?.nome}</DialogTitle>
          </DialogHeader>
          <form onSubmit={confirmarReset} className="space-y-4" noValidate>
            <FormField
              label="Nova senha"
              htmlFor="nova-senha"
              obrigatorio
              erro={erroSenha}
            >
              <Input
                id="nova-senha"
                type="password"
                value={novaSenha}
                onChange={(e) => setNovaSenha(e.target.value)}
                placeholder="Mínimo de 6 caracteres"
              />
            </FormField>
            <p className="text-xs text-muted-foreground">
              Informe a nova senha ao usuário por um canal seguro. Recomende a
              troca no primeiro acesso.
            </p>
            <DialogFooter className="gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setResetAlvo(null)}
              >
                Cancelar
              </Button>
              <Button type="submit" disabled={redefinindo}>
                {redefinindo ? <Spinner className="h-4 w-4" /> : "Redefinir"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Confirmar ativar/desativar */}
      <ConfirmDialog
        aberto={!!ativoAlvo}
        titulo={
          ativoAlvo?.ativo
            ? `Desativar ${ativoAlvo?.nome}?`
            : `Ativar ${ativoAlvo?.nome}?`
        }
        descricao={
          ativoAlvo?.ativo
            ? "O usuário não conseguirá mais acessar o sistema. A conta é preservada e pode ser reativada depois."
            : "O usuário voltará a poder acessar o sistema."
        }
        textoConfirmar={ativoAlvo?.ativo ? "Desativar" : "Ativar"}
        destrutivo={ativoAlvo?.ativo}
        processando={processandoAtivo}
        onConfirmar={confirmarAtivo}
        onCancelar={() => setAtivoAlvo(null)}
      />
    </div>
  );
}
