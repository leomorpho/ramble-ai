import { browser } from '$app/environment';
import { pb } from '$lib/pocketbase.js';

/**
 * PocketBase authentication store for Svelte 5 runes
 */
class AuthStore {
  #user = $state(null);
  #isValid = $state(false);
  #isLoading = $state(true);

  constructor() {
    if (browser) {
      this.#initializeAuth();
    }
  }

  #initializeAuth() {
    // PocketBase automatically loads from localStorage on instantiation
    this.#user = pb.authStore.model;
    this.#isValid = pb.authStore.isValid;
    this.#isLoading = false;

    // Listen for PocketBase auth changes
    pb.authStore.onChange(() => {
      this.#user = pb.authStore.model;
      this.#isValid = pb.authStore.isValid;
    }, true);

    // Try to refresh auth if we have a token
    if (pb.authStore.isValid) {
      this.#refreshAuth();
    }
  }

  async #refreshAuth() {
    try {
      if (pb.authStore.isValid) {
        await pb.collection('users').authRefresh();
        this.#user = pb.authStore.model;
        this.#isValid = pb.authStore.isValid;
      }
    } catch (error) {
      console.warn('Failed to refresh auth:', error);
      this.logout();
    } finally {
      this.#isLoading = false;
    }
  }

  // Getters
  get user() {
    return this.#user;
  }

  get isValid() {
    return this.#isValid;
  }

  get isLoading() {
    return this.#isLoading;
  }

  get isAuthenticated() {
    return this.#isValid && this.#user;
  }

  // Actions
  async login(email, password) {
    try {
      const authData = await pb.collection('users').authWithPassword(email, password);
      return { success: true, user: authData.record };
    } catch (error) {
      console.error('Login failed:', error);
      return { success: false, error: error.message };
    }
  }

  logout() {
    pb.authStore.clear();
  }
}

// Create and export the singleton auth store instance
export const authStore = new AuthStore();