// Utilities
import router from '../router';

import { markRaw } from 'vue';
import { createPinia } from 'pinia';

const pinia = createPinia();

pinia.use(({ store }) => {
  store.$router = markRaw(router);
});

export default pinia;
