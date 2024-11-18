/**
 * plugins/index.ts
 *
 * Automatically included in `./src/main.ts`
 */

// Plugins
import vuetify from './vuetify';
import stores from '../stores';
import router from '../router';

import VNetworkGraph from 'v-network-graph';
import 'v-network-graph/lib/style.css';

// Types
import type { App } from 'vue';

export function registerPlugins (app: App) {
  app
    .use(VNetworkGraph)
    .use(vuetify)
    .use(stores)
    .use(router);
};
