<template>
  <v-network-graph :zoom-level="props.zoomLevel || 2" :nodes="graph.nodes" :edges="graph.edges" :layouts="graph.layouts" :layers="graph.layers" :configs="configs" :style="{ filter: theme.global.current.value.dark ? 'invert(1)' : null }">
    <template #override-node="{ scale, nodeId, config }">
      <polygon :fill="config.color" :stroke="config.strokeColor" :stroke-width="2 * scale" :points="`0,${config.height * scale} -${config.width * (Math.max(5, String(nodeId).length) / 8) * scale},0 0,-${config.height * scale} ${config.width * (Math.max(5, String(nodeId).length) / 8) * scale},0`"></polygon>
    </template>
    <template #edge-label="{ edge, hovered, area, scale }">
      <g :style="`transform: translate(${(area.source.above.x + area.target.above.x) / 2 - (Math.max(edge.source.length, edge.target.length) * 7 + 28) / 2 * scale}px, ${(area.source.above.y + area.target.above.y) / 2 + -10 * scale}px);z-index:999;`" @pointerenter.passive="handleNodePointerEvent(true, $event)" @pointerleave.passive="handleNodePointerEvent(false, $event)">
        <rect v-if="!hovered && alwaysShowOpinions" :width="(Math.max(alwaysShowOpinions ? 90 : 25, Math.max(edge.source.length, edge.target.length) * 7 + 28)) * scale" :height="(alwaysShowOpinions ? 45 : 32) * scale" :rx="5 * scale" :ry="5 * scale" :style="`fill: #ebeef0; stroke: #dfe3e6; stroke-width: ${2 * scale}px`" />
        <template v-else-if="hovered">
          <rect :width="Math.max(Math.max(edge.source.length, edge.target.length) * 7 + 28, 130) * scale" :height="100 * scale" :rx="5 * scale" :ry="5 * scale" :style="`fill: #e0e6f7; stroke: #3355bb; stroke-width: ${2 * scale}px`" />
          <text :x="18 * scale" :y="45 * scale" :font-size="`${12 * scale}px`">Belief</text>
          <text :x="18 * scale" :y="61 * scale" :font-size="`${12 * scale}px`">Disbelief</text>
          <text :x="18 * scale" :y="76 * scale" :font-size="`${12 * scale}px`">Uncertainty</text>
          <text :x="18 * scale" :y="91 * scale" :font-size="`${12 * scale}px`">Base rate</text>
          <text :x="98 * scale" :y="45 * scale" :font-size="`${12 * scale}px`">{{ edge.belief.toFixed(2) }}</text>
          <text :x="98 * scale" :y="61 * scale" :font-size="`${12 * scale}px`">{{ edge.disbelief.toFixed(2) }}</text>
          <text :x="98 * scale" :y="76 * scale" :font-size="`${12 * scale}px`">{{ edge.uncertainty.toFixed(2) }}</text>
          <text :x="98 * scale" :y="91 * scale" :font-size="`${12 * scale}px`">{{ edge.base_rate.toFixed(2) }}</text>
        </template>
        <text v-if="!hovered && alwaysShowOpinions" :x="5 * scale" :y="40 * scale" :font-size="`${8 * scale}px`" font-family="monospace">
          [{{ edge.belief.toFixed(1) }} / {{ edge.disbelief.toFixed(1) }} / {{ edge.uncertainty.toFixed(1) }}]
        </text>
        <text v-if="hovered || alwaysShowOpinions" font-style="italic" :x="5 * scale" :y="20 * scale" :font-size="`${16 * scale}px`">ω</text>
        <text v-if="hovered || alwaysShowOpinions" font-style="italic" :x="23 * scale" :y="12 * scale" :font-size="`${12 * scale}px`">{{ edge.source }}</text>
        <text v-if="hovered || alwaysShowOpinions" font-style="italic" :x="23 * scale" :y="28 * scale" :font-size="`${12 * scale}px`">{{ edge.target }}</text>
      </g>
    </template>
  </v-network-graph>
</template>

<script lang='ts' setup>
import ColorHash from 'color-hash';
import { useTheme } from 'vuetify';
import { computed } from 'vue';

const props = defineProps<{
  panEnabled?: boolean,
  zoomEnabled?: boolean,
  zoomLevel?: Number,
  state: TrustModelInstanceState,
  alwaysShowOpinions?: boolean,
}>()

const theme = useTheme();

const nodeSize = 40;
const colorHash = new ColorHash();

import * as vNG from 'v-network-graph';

import dagre from 'dagre';
import { TrustModelInstanceState } from '@/stores/app';

const graph = computed<{ nodes: vNG.Nodes, edges: vNG.Edges, layouts: vNG.Layouts, layers: vNG.Layers}>(() => {
  const state = props.state;
  const nodes = new Set(state?.Structure.adjacency_list.flatMap(e => [e.sourceNode, ...e.targetNodes]) || []);
  const values = Object.fromEntries(
    Object.values(state?.Values || {})
      .flat()
      .map(e => [`e-${e.source}-${e.destination}`, e.opinion])
  );
  const edges = state?.Structure.adjacency_list.flatMap(e => e.targetNodes.map(t => ({
    source: e.sourceNode,
    target: t,
    belief: values[`e-${e.sourceNode}-${t}`].belief,
    disbelief: values[`e-${e.sourceNode}-${t}`].disbelief,
    uncertainty: values[`e-${e.sourceNode}-${t}`].uncertainty,
    base_rate: values[`e-${e.sourceNode}-${t}`].base_rate,
    label: `[${[
      values[`e-${e.sourceNode}-${t}`].belief,
      values[`e-${e.sourceNode}-${t}`].disbelief,
      values[`e-${e.sourceNode}-${t}`].uncertainty,
      values[`e-${e.sourceNode}-${t}`].base_rate
    ].map(e => e.toFixed(2)).join(' / ')}]`
  }))) || [];

  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({
    rankdir: 'TB',
    edgesep: nodeSize,
    nodesep: nodeSize * 4,
    ranksep: nodeSize * 4,
  });

  for (const node of nodes) {
    g.setNode(node, {
      label: node,
      width: nodeSize,
      height: nodeSize
    });
  }

  for (const edge of edges) {
    g.setEdge(edge.source, edge.target);
  }

  dagre.layout(g);

  return {
    nodes: Object.fromEntries([...nodes].map(e => [e, {
      name: e,
      color: colorHash.hex(e),
    }])),
    edges: Object.fromEntries(edges.map(e => [`e-${e.source}-${e.target}`, e])),
    layers: {
      tooltips: 'paths'
    },
    layouts: {
      nodes: Object.fromEntries([...g.nodes()].map(e => [e, { x: g.node(e).x, y: g.node(e).y }]))
    }
  };
});

const configs = vNG.defineConfigs({
  view: {
    builtInLayerOrder: ['edge-labels', 'paths'],
    autoPanAndZoomOnLoad: 'fit-content',
    zoomEnabled: props.zoomEnabled !== false,
    panEnabled: props.panEnabled !== false,
    autoPanOnResize: true,
  },
  node: {
    draggable: false,
    normal: {
      type: 'rect',
      color: e => `${e.color || '#c2c6ca'}80`,
      strokeColor: e => `${e.color || '#c2c6ca'}60`,
      width: nodeSize,
      height: nodeSize * 0.7,
      borderRadius: 0,
    },
    hover: {
      color: e => `${e.color || '#c2c6ca'}a0`,
      strokeColor: e => `${e.color || '#c2c6ca'}40`,
    },
    label: {
      visible: true,
      direction: 'center',
      color: '#000'
    },
  },
  edge: {
    normal: {
      color: '#aaa',
      width: 3,
    },
    margin: 10,
    marker: {
      target: {
        type: 'arrow',
        width: 4,
        height: 4,
      },
    },
  },
});

function handleNodePointerEvent(hover: boolean, event: PointerEvent) {
  const state = (event.target as any)?.__vueParentComponent?.props?.state;
  if (state) {
    state.hovered = hover;
  }
}
</script>
