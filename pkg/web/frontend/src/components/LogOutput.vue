<template>
  <div class="ansioutput">
    <span v-for="(chunk, i) in chunks" :key="i" :style="chunk.css">{{ chunk.text }}</span>
  </div>
</template>

<script setup lang="ts">
import { parse } from 'ansicolor';
import { computed } from 'vue';

const props = defineProps<{ output: string }>();
const chunks = computed(() => parse(props.output) as any);
</script>

<style scoped>
.ansioutput * {
  font-family: monospace;
  white-space: pre;
}
</style>