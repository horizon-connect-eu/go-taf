<template>
  <Teleport to="#toolbar">
    <faceted-search v-model="filteredItems" :columns="headers" :items="items" sync-to-query v-model:sort-by="sortBy" :defaults="{ sort: '-version' }" />

    <v-btn icon title="Refresh" size="small" @click="refresh" variant="text">
      <v-icon icon="mdi-reload" />
    </v-btn>
  </Teleport>

  <v-data-table-virtual :headers="headers" :items="filteredItems" :height="height" v-resize="onResize" v-model:sort-by="sortBy" ref="table">
    <template #[`item.version`]="{ item }">
      <v-chip class="pr-0 mt-1" :to="`/tmis/${route.params.client as string}/${route.params.sessionID as string}/${route.params.template as string}/${route.params.id as string}/${item.version}`">
        Version
        <v-chip class="ml-1 font-weight-bold">{{ item.version }}</v-chip>
      </v-chip>
    </template>
    <template #[`item.updates`]="{ item }">
      <ul class="mt-1">
        <li v-for="(update, i) in item.updates" :key="i"><pre>{{ update }}</pre></li>
      </ul>
    </template>
    <template #[`item.atls`]="{ item }">
      <v-card variant="outlined" class="mx-2 my-1" v-for="(_, scope) in item.atls?.SlResults" :key="scope">
        <template #title>
          <div class="text-overline mt-n2">ATL Scope: <code>{{ scope }}</code></div>
        </template>

        <v-table density="compact" class="mt-n4">
          <tbody>
            <tr>
              <th class="text-left">Subjective Logic</th>
              <td v-if="item.atls.SlResults[scope]?.belief !== undefined"><code>({{ item.atls.SlResults[scope].belief?.toFixed(2) }} / {{ item.atls.SlResults[scope].disbelief.toFixed(2) }} / {{ item.atls.SlResults[scope].uncertainty.toFixed(2) }} / {{ item.atls.SlResults[scope].base_rate.toFixed(2) }})</code></td>
              <td v-else><code>{{ item.atls.SlResults[scope] }}</code></td>
            </tr>
            <tr>
              <th class="text-left">Projected Probability</th>
              <td><code>{{ item.atls.PpResults[scope] }}</code></td>
            </tr>
            <tr>
              <th class="text-left">Trust Decision</th>
              <td><code>{{ item.atls.TdResults[scope] }}</code></td>
            </tr>
          </tbody>
        </v-table>
      </v-card>
    </template>
    <template #[`item.state`]="{ item }">
      <v-card variant="outlined" class="mx-2 my-1" v-for="(rows, scope) in item.state.Values" :key="scope">
        <template #title>
          <div class="text-overline mt-n2">Scope: <code>{{ scope }}</code></div>
        </template>

        <v-table density="compact" class="mt-n4">
          <thead>
            <tr>
              <th class="text-left">Source</th>
              <th class="text-left">Destination</th>
              <th class="text-left">Opinion</th>
              <!--
              <th class="text-left">Belief</th>
              <th class="text-left">Disbelief</th>
              <th class="text-left">Uncertainty</th>
              <th class="text-left">Base Rate</th>
              -->
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, i) in rows" :key="i">
              <td>{{ row.source }}</td>
              <td>{{ row.destination }}</td>
              <td><code>({{ row.opinion.belief.toFixed(2) }} / {{ row.opinion.disbelief.toFixed(2) }} / {{ row.opinion.uncertainty.toFixed(2) }} / {{ row.opinion.base_rate.toFixed(2) }})</code></td>
              <!--
              <td><code>{{ row.opinion.belief.toFixed(2) }}</code></td>
              <td><code>{{ row.opinion.disbelief.toFixed(2) }}</code></td>
              <td><code>{{ row.opinion.uncertainty.toFixed(2) }}</code></td>
              <td><code>{{ row.opinion.base_rate.toFixed(2) }}</code></td>
              -->
            </tr>
          </tbody>
        </v-table>
      </v-card>
    </template>
    <template #[`item.graph`]="{ item }">
      <div style="min-height: 500px; min-width: 500px; height: 100%;">
        <trust-graph :state="item.state" :zoom-enabled="false" :pan-enabled="false" :zoom-level="2.5" />
      </div>
    </template>
  </v-data-table-virtual>
</template>

<style>
.v-data-table__tr {
  vertical-align: top;
}
</style>

<script lang='ts' setup>
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';

import { useAppStore } from '@/stores/app';
import { VDataTableVirtual } from 'vuetify/components';
import { Column, SortItem } from '@/types';
import router from '@/router';

const filteredItems = ref<any[]>([]);
const sortBy = ref<SortItem[]>([]);
const table = ref<null|VDataTableVirtual>(null);

const route = useRoute();

const store = useAppStore();
const headers: Column[] = [{
  maxWidth: '100px',
  filterable: true,
  title: 'Version',
  key: 'version',
}, {
  filterable: true,
  maxWidth: '300px',
  title: 'Updates',
  key: 'updates'
}, {
  filterable: false,
  title: 'Graph',
  key: 'graph'
}, {
  filterable: false,
  maxWidth: '350px',
  title: 'State',
  key: 'state'
}, {
  maxWidth: '350px',
  filterable: false,
  title: 'ATLs',
  key: 'atls'
}];

const trustModelInstance = computed(() => store.trustModelInstances[`//${route.params.client}/${route.params.sessionID}/${route.params.template}/${route.params.id}`]);
// const items = computed(() => Object.values(trustModelInstance.value.states || {}));
const items = computed(() => Object.entries(trustModelInstance.value.states || {}).map(([k, v]) => ({
  version: k,
  updates: trustModelInstance.value.updates?.[k]?.map?.((e: any) => JSON.stringify(e, null, 2)),
  atls: trustModelInstance.value.atls?.[k],
  state: v
})));

const height = ref(document.body.clientHeight - 48);

function onResize() {
  height.value = document.body.clientHeight - 48;
}

async function refresh() {
  try {
    await store.fetchTrustModelInstance(
      route.params.client as string,
      route.params.sessionID as string,
      route.params.template as string,
      route.params.id as string,
    );
  } catch {
    router.push('/');
  }
}

refresh();
</script>
