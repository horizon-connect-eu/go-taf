<template>
  <v-network-graph :zoom-level="2" :nodes="graph.nodes" :edges="graph.edges" :layouts="graph.layouts" :configs="configs" :style="{ filter: theme.global.current.value.dark ? 'invert(1)' : null }">
    <template #override-node="{ scale, config }">
      <rect :width="config.width * scale" :height="config.height * scale" :style="`transform: rotate(45deg) translate(-${config.width * scale * 0.5}px, -${config.height * scale * 0.5}px); fill: #c2c6ca; stroke: #616365; stroke-width: ${2 * scale}px`" />
    </template>
    <template #edge-overlay="{ hovered, edge, scale, center }">
      <!-- Place the triangle at the center of the edge -->
      <g :style="`transform: translate(${center.x + 25 * scale}px, ${center.y + -15 * scale}px);z-index:999;`" @pointerenter.passive="handleNodePointerEvent(true, $event)" @pointerleave.passive="handleNodePointerEvent(false, $event)">
        <rect v-if="!hovered" :width="(Math.max(edge.source.length, edge.target.length) * 8 + 40) * scale" :height="40 * scale" :rx="5 * scale" :ry="5 * scale" :style="`fill: #ebeef0; stroke: #dfe3e6; stroke-width: ${2 * scale}px`" />
        <template v-else>
          <rect :width="160 * scale" :height="Math.max(Math.max(edge.source.length, edge.target.length) * 8 + 40, 130) * scale" :rx="5 * scale" :ry="5 * scale" :style="`fill: #e0e6f7; stroke: #3355bb; stroke-width: ${2 * scale}px`" />
          <text :x="18 * scale" :y="56 * scale" :font-size="`${14 * scale}px`">Belief</text>
          <text :x="18 * scale" :y="76 * scale" :font-size="`${14 * scale}px`">Disbelief</text>
          <text :x="18 * scale" :y="96 * scale" :font-size="`${14 * scale}px`">Uncertainty</text>
          <text :x="18 * scale" :y="116 * scale" :font-size="`${14 * scale}px`">Base rate</text>
          <text :x="118 * scale" :y="56 * scale" :font-size="`${14 * scale}px`">{{ edge.belief.toFixed(2) }}</text>
          <text :x="118 * scale" :y="76 * scale" :font-size="`${14 * scale}px`">{{ edge.disbelief.toFixed(2) }}</text>
          <text :x="118 * scale" :y="96 * scale" :font-size="`${14 * scale}px`">{{ edge.uncertainty.toFixed(2) }}</text>
          <text :x="118 * scale" :y="116 * scale" :font-size="`${14 * scale}px`">{{ edge.base_rate.toFixed(2) }}</text>
        </template>
        <text font-style="italic" :x="5 * scale" :y="26 * scale" :font-size="`${20 * scale}px`">ω</text>
        <text font-style="italic" :x="23 * scale" :y="18 * scale" :font-size="`${16 * scale}px`">{{ edge.source }}</text>
        <text font-style="italic" :x="23 * scale" :y="34 * scale" :font-size="`${16 * scale}px`">{{ edge.target }}</text>
      </g>
    </template>
    <template #edge-label="{ edge, ...slotProps }">
      <v-edge-label :text="edge.label" align="center" vertical-align="above" v-bind="slotProps" />
    </template>
  </v-network-graph>
</template>

<script lang='ts' setup>
import { useTheme } from 'vuetify';
import { computed } from 'vue';

const props = defineProps<{
  panEnabled?: boolean,
  zoomEnabled?: boolean,
  state: TrustModelInstanceState
}>()

const theme = useTheme();

const nodeSize = 40;

import * as vNG from 'v-network-graph';

import dagre from 'dagre';
import { TrustModelInstanceState } from '@/stores/app';

const graph = computed<{ nodes: vNG.Nodes, edges: vNG.Edges, layouts: vNG.Layouts}>(() => {
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
    ranksep: nodeSize * 2,
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
    nodes: Object.fromEntries([...nodes].map(e => [e, { name: e }])),
    edges: Object.fromEntries(edges.map(e => [`e-${e.source}-${e.target}`, e])),
    layouts: {
      nodes: Object.fromEntries([...g.nodes()].map(e => [e, { x: g.node(e).x, y: g.node(e).y }]))
    }
  };
});

const configs = vNG.defineConfigs({
  view: {
    autoPanAndZoomOnLoad: 'fit-content',
    zoomEnabled: props.zoomEnabled !== false,
    panEnabled: props.panEnabled !== false,
    autoPanOnResize: true,
  },
  node: {
    draggable: false,
    normal: {
      type: 'rect',
      width: nodeSize,
      height: nodeSize,
      borderRadius: 0,
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
    margin: 4,
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
