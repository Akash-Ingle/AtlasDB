import { create } from "zustand";

interface AuthState {
  token: string | null;
  email: string | null;
  isAuthenticated: boolean;
  setAuth: (token: string, email: string) => void;
  logout: () => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  email: null,
  isAuthenticated: false,

  setAuth: (token: string, email: string) => {
    localStorage.setItem("atlas_token", token);
    localStorage.setItem("atlas_email", email);
    set({ token, email, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem("atlas_token");
    localStorage.removeItem("atlas_email");
    set({ token: null, email: null, isAuthenticated: false });
  },

  hydrate: () => {
    if (typeof window === "undefined") return;
    const token = localStorage.getItem("atlas_token");
    const email = localStorage.getItem("atlas_email");
    if (token) {
      set({ token, email, isAuthenticated: true });
    }
  },
}));
