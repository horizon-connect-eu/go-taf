<template>
  <Teleport to="#toolbar">
    <v-btn icon title="Refresh" size="small" @click="refresh" variant="text">
      <v-icon icon="mdi-reload" />
    </v-btn>
  </Teleport>

  <v-slider v-model="version" :min="0" :max="trustModelInstance.latestVersion" :step="1" show-ticks="always" tick-size="4" class="mt-1 ml-4">
    <template #append>
      <v-chip class="pr-0">
        Version
        <v-chip class="ml-2 font-weight-bold">{{ version }}</v-chip>
      </v-chip>
    </template>
  </v-slider>

  <div style="height: calc(100vh - 114px)" v-if="state">
    <trust-graph :state="state" always-show-opinions />
  </div>
</template>

<script lang='ts' setup>
import { computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAppStore } from '@/stores/app';

const route = useRoute();
const store = useAppStore();
const router = useRouter();

const version = computed({
  get() {
    return Number(route.params.version as string);
  },

  set(newVersion) {
    router.replace(`/tmis/${route.params.client}/${route.params.sessionID}/${route.params.template}/${route.params.id}/${newVersion}`);
  }
});

watch(() => route.params.version, () => refresh());

const state = computed(() => trustModelInstance.value.states?.[version.value]);

async function refresh() {
  await store.fetchTrustModelInstance(
    route.params.client as string,
    route.params.sessionID as string,
    route.params.template as string,
    route.params.id as string,
    route.params.version as string
  );
}

const trustModelInstance = computed(() => store.trustModelInstances[`//${route.params.client}/${route.params.sessionID}/${route.params.template}/${route.params.id}`]);

refresh();
</script>
