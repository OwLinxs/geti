import axios, {
  AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from "axios";
import type { ErroValidacao } from "@/types";

const TOKEN_KEY = "sige-ti.token";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

// Em dev, VITE_API_BASE_URL vazio => usa o proxy do Vite (/api).
// Em produção, defina VITE_API_BASE_URL com a URL pública da API.
const baseURL = `${import.meta.env.VITE_API_BASE_URL ?? ""}/api/v1`;

export const api: AxiosInstance = axios.create({
  baseURL,
  headers: { "Content-Type": "application/json" },
});

// Interceptor de requisição: injeta o token JWT (Bearer) quando disponível.
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Callback para reagir a 401 (sessão expirada/invalida) — definido pelo AuthContext.
let onUnauthorized: (() => void) | null = null;
export function setOnUnauthorized(cb: () => void) {
  onUnauthorized = cb;
}

// Interceptor de resposta: trata 401 globalmente.
api.interceptors.response.use(
  (res) => res,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      clearToken();
      if (onUnauthorized) onUnauthorized();
    }
    return Promise.reject(error);
  }
);

// Extrai uma mensagem de erro legível (PT-BR) a partir da resposta da API.
export function mensagemErro(error: unknown, padrao = "Ocorreu um erro."): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data as
      | { erro?: string; detalhe?: string }
      | undefined;
    if (data?.erro) return data.erro;
    if (error.message === "Network Error") {
      return "Não foi possível conectar ao servidor. Verifique se o backend está em execução.";
    }
  }
  return padrao;
}

// Extrai o mapa de erros por campo (HTTP 422), se houver.
export function camposInvalidos(error: unknown): Record<string, string> | null {
  if (axios.isAxiosError(error) && error.response?.status === 422) {
    const data = error.response.data as ErroValidacao;
    return data.campos ?? null;
  }
  return null;
}
