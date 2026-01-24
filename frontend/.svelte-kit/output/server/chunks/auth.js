import { w as writable } from "./index.js";
function createAuthStore() {
  let initialToken = null;
  let initialUser = null;
  const { subscribe, set, update } = writable({
    token: initialToken,
    user: initialUser
  });
  return {
    subscribe,
    login: (token, user) => {
      set({ token, user });
    },
    logout: () => {
      set({ token: null, user: null });
    }
  };
}
const auth = createAuthStore();
export {
  auth as a
};
