/**
 * router/index.ts
 *
 * Automatic routes for `./src/pages/*.vue`
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router';

import sessions from '@/pages/sessions/index.vue';
import version from '@/pages/tmis/client/sessionID/template/id/version/index.vue';
import template from '@/pages/tmis/client/sessionID/template/id/index.vue';
import tmis from '@/pages/tmis/index.vue';

const routes = [{
  path: '/sessions',
  children: [{
    path: '',
    name: '/sessions/',
    component: sessions
  }]
}, {
  path: '/tmis',
  children: [{
    path: '',
    name: '/tmis/',
    component: tmis
  }, {
    path: ':client',
    children: [{
      path: ':sessionID',
      children: [{
        path: ':template',
        children: [{
          path: ':id',
          children: [{
            path: '',
            name: '/tmis/:client/:sessionID/:template/:id/',
            component: template
          }, {
            path: ':version',
            children: [{
              path: '',
              name: '/tmis/:client/:sessionID/:template/:id/:version/',
              component: version
            }]
          }]
        }]
      }]
    }]
  }]
}, {
  path: '/',
  redirect: '/tmis'
}];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: routes,
});

// Workaround for https://github.com/vitejs/vite/issues/11804
router.onError((err, to) => {
  if (err?.message?.includes?.('Failed to fetch dynamically imported module')) {
    if (!localStorage.getItem('vuetify:dynamic-reload')) {
      console.log('Reloading page to fix dynamic import error');
      localStorage.setItem('vuetify:dynamic-reload', 'true');
      location.assign(to.fullPath);
    } else {
      console.error('Dynamic import error, reloading page did not fix it', err);
    }
  } else {
    console.error(err)
  }
});

router.isReady().then(() => {
  localStorage.removeItem('vuetify:dynamic-reload');
});

export default router;
