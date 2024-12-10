<template>
  <Teleport to="#toolbar">
    <faceted-search v-model="filteredItems" :columns="headers" :items="items" sync-to-query v-model:sort-by="sortBy" :defaults="{}" />

    <v-btn icon title="Refresh" size="small" @click="store.fetchSessions()" variant="text">
      <v-icon icon="mdi-reload" />
    </v-btn>
  </Teleport>

  <v-data-table-virtual :headers="headers" :items="filteredItems" :height="height" v-resize="onResize" multi-sort v-model:sort-by="sortBy" ref="table">
    <template #[`item.Client`]="{ item }">
      <code class="mt-1">{{ item.Client }}</code>
    </template>
    <template #[`item.SessionID`]="{ item }">
      <code class="mt-1">{{ item.SessionID }}</code>
    </template>
    <template #[`item.Template`]="{ item }">
      <code class="mt-1">{{ item.Template }}</code>
    </template>
    <template #[`item.IsActive`]="{ item }">
      <v-checkbox-btn v-model="item.IsActive" readonly />
    </template>
    <template #[`item.TMIs`]="{ item }">
      <v-chip v-for="(tmi, i) in item.TMIs" :key="i" :to="`/tmis/${item.Client}/${item.SessionID}/${item.Template}/${tmi}`" class="mx-1" :title="`//${item.Client}/${item.SessionID}/${item.Template}/${tmi}`">
        <code>/{{ tmi }}</code>
      </v-chip>
    </template>
  </v-data-table-virtual>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';

import { useAppStore } from '@/stores/app';
import { VDataTableVirtual } from 'vuetify/components';
import { Column, SortItem } from '@/types';

const sortBy = ref<SortItem[]>([]);
const filteredItems = ref<any[]>([]);
const table = ref<null|VDataTableVirtual>(null);

const store = useAppStore();
const headers: Column[] = [{
  maxWidth: '125px',
  filterable: true,
  title: 'Client',
  key: 'Client'
}, {
  filterable: true,
  title: 'Session ID',
  key: 'SessionID'
}, {
  filterable: true,
  title: 'Template',
  key: 'Template'
}, {
  filterable: true,
  title: 'Active',
  key: 'IsActive'
}, {
  filterable: true,
  title: 'TMIs',
  key: 'TMIs'
}];

const items = computed(() => Object.values(store.sessions));

const height = ref(document.body.clientHeight - 48);

function onResize() {
  height.value = document.body.clientHeight - 48;
}

</script>
