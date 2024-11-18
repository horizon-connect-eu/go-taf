<template>
  <v-menu v-model="menu" :close-on-content-click="false" transition="scale-transition">
    <template #activator="{ props }">
      <v-text-field variant="solo" density="compact" hide-details single-line v-model="search" prepend-inner-icon="mdi-magnify" append-icon="" label="Search" :items="items" autocomplete="off" multiple v-bind="props" width="20%" @input="update()" placeholder="Search" max-width="30em">
        <template #append-inner>
          <v-chip v-for="(filter, i) in filters" :key="i" label @click.stop="editFilter(filter)" @click:close.stop="deleteFilter(filter)" small closable class="ml-2">
            <b>
              <v-icon left small>mdi-filter-outline</v-icon>
              {{ filter.text }}:
            </b>
            <span :class="`pl-2 ${filter?.filterSettings?.items?.filter?.(e => e.value === filter.filter && e.color)?.map?.(e => `text-${e.color}`)?.join('')}`">
              {{ filter.filter }}
            </span>
          </v-chip>
        </template>
      </v-text-field>
    </template>
    <v-list density="compact">
      <v-list-subheader>Add filter</v-list-subheader>
      <v-list-item v-for="(item, i) in items" :key="i" @click="addFilter(item)" prepend-icon="mdi-filter-outline">
        <v-list-item-title>{{ item.title }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script lang="ts" setup>
import { onMounted, ref, inject, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Dialog } from '@/symbols';
import { Column, FilterSettings, SortItem } from '@/types';

const regexify = (value: string) => /^\/.*\/$/.test(value) ? new RegExp(value.slice(1, -1)) : value;

const router = useRouter();
const route = useRoute();

const $dialog = inject(Dialog);

const props = defineProps<{
  items: any[],
  modelValue: any[],
  columns: Column[],
  sortBy: SortItem[],
  syncToQuery: boolean,
  defaults?: { [key: string]: string },
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', payload: any[]): void,
  (event: 'update:sortBy', payload: SortItem[]): void
}>();

type Filter = {
  text: string,
  value: string,
  filter: string | RegExp,
  filterSettings?: FilterSettings
};

const search = ref('');
const filters = ref<Filter[]>([]);
const menu = ref(false);

onMounted(() => {
  const useDefaults = !Object.keys(route.query).some(e => e === 'sort' || e === 'filter_term' || e.startsWith('filter_by_'));
  const query = useDefaults ? props.defaults : route.query;

  if (query?.['filter_term']) {
    search.value = String(query['filter_term']);
  }

  if (query?.['sort']) {
    const sortBy: SortItem[] = String(query['sort'])
      .split(',')
      .map((e) => e.startsWith('-') ? { key: e.slice(1), order: 'desc' } : { key: e, order: 'asc' });

    emit('update:sortBy', sortBy);
  }

  for (const item of props.columns) {
    if (query?.[`filter_by_${item.key}`]) {
      filters.value.push({
        text: item.title,
        value: item.key,
        filter: regexify(String(query[`filter_by_${item.key}`])),
        filterSettings: item.filterSettings
      });
    }
  }

  update(false);
});


function promptFilter(item: Filter) {
  switch (item.filterSettings?.type) {
    case 'select':
      return $dialog?.select(`Filter Entries by "${item.text}"`, String(item.filter || ''), {
        confirmText: 'Filter',
        confirmIcon: 'mdi-filter-plus-outline',
        cancelIcon: 'mdi-cancel',
        icon: 'mdi-filter-outline',
        items: item.filterSettings.items,
        itemText: item.filterSettings.itemText,
        itemValue: item.filterSettings.itemValue
      });

    default:
      return $dialog?.prompt(`Filter Entries by "${item.text}"`, String(item.filter || ''), {
        confirmText: 'Filter',
        confirmIcon: 'mdi-filter-plus-outline',
        cancelIcon: 'mdi-cancel',
        icon: 'mdi-filter-outline'
      });
  }
}

async function addFilter(item: Column) {
  menu.value = false;

  const filter: Filter = {
    text: item.title,
    value: item.key,
    filter: '',
    filterSettings: item.filterSettings
  };

  const value = await promptFilter(filter);

  if (value) {
    filter.filter = regexify(value);
    filters.value.push(filter);

    update();
  }
}

async function editFilter(item: Filter) {
  const value = await promptFilter(item);

  if (value) {
    item.filter = regexify(value);
    update();
  }
}

function deleteFilter(item: Filter) {
  filters.value.splice(filters.value.indexOf(item), 1);
  update();
}

function update(syncToQuery: boolean=true) {
  if (syncToQuery) {
    syncQuery();
  }

  let items = props.items;

  if (search.value?.length) {
    const term = search.value.toLocaleLowerCase();
    items = items.filter((item) => props.columns.some((e) => {
        if (!e.filterable) {
          return false;
        }

        const value = getObjectValueByPath(item, e.key);

        return value != null && typeof value !== 'boolean' && value.toString()
          .toLocaleLowerCase()
          .indexOf(term) !== -1;
    }));
  }

  if (filters.value?.length && Array.isArray(props.items)) {
    items = items.filter((item) => filters.value.every((f) => {
      const value = getObjectValueByPath(item, f.value);
      if (f.filter instanceof RegExp) {
        return f.filter.test(value);
      }

      return value != null && f.filter != null && typeof value !== 'boolean' &&
        value.toString().toLocaleLowerCase().indexOf(f.filter.toLocaleLowerCase()) !== -1;
    }));
  }

  emit('update:modelValue', items);
}

function syncQuery() {
  if (props.syncToQuery) {
    const queryUpdates: {[key: string]: string} = {};
    if (search.value !== route.query['filter_term']) {
      queryUpdates['filter_term'] = search.value;
    }

    let sort = '';
    if (props.sortBy?.length) {
      sort = props.sortBy
        .map(e => `${e.order === 'desc' ? '-' : ''}${e.key}`)
        .join(',');
    }

    if (sort !== route.query['sort']) {
      queryUpdates['sort'] = sort;
    }

    const definedFilters: {[key: string]: boolean} = {};
    for (const filter of filters.value) {
      definedFilters[`filter_by_${filter.value}`] = true;

      const tmp = String(filter.filter);
      if (tmp !== route.query[`filter_by_${filter.value}`]) {
        queryUpdates[`filter_by_${filter.value}`] = tmp;
      }
    }

    const filteredQuery = Object.fromEntries(
      Object.entries(route.query)
        .filter(e => !e[0].startsWith('filter_by_') || definedFilters[e[0]])
    );

    if (Object.keys(queryUpdates).length || Object.keys(filteredQuery).length !== Object.keys(route.query).length) {
      router.replace({
        query: { ...filteredQuery, ...queryUpdates }
      });
    }
  }
}

const items = computed(() => {
  return props.columns.filter((e) => e.filterable && !filters.value.some(f => f.value === e.key));
});

watch(() => props.items, () => update(false));
watch(() => props.sortBy, () => syncQuery(), { deep: true });

function getObjectValueByPath(obj: any, path: string, fallback?: any): any {
  // credit: http://stackoverflow.com/questions/6491463/accessing-nested-javascript-objects-with-string-key#comment55278413_6491621
  if (obj == null || !path || typeof path !== 'string') {
    return fallback;
  }
  if (obj[path] !== undefined) {
    return obj[path];
  }
  path = path.replace(/\[(\w+)\]/g, '.$1'); // convert indexes to properties
  path = path.replace(/^\./, ''); // strip a leading dot
  return getNestedValue(obj, path.split('.'), fallback);
}

function getNestedValue(obj: any, path: (string | number)[], fallback?: any): any {
  // credit: https://github.com/vuetifyjs/vuetify/blob/master/packages/vuetify/src/util/helpers.ts
  const last = path.length - 1;

  if (last < 0) {
    return obj === undefined ? fallback : obj;
  }

  for (let i = 0; i < last; i++) {
    if (obj == null) {
      return fallback;
    }
    obj = obj[path[i]];
  }

  if (obj == null) {
    return fallback;
  }

  return obj[path[last]] === undefined ? fallback : obj[path[last]];
}

</script>
