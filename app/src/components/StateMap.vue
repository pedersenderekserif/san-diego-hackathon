<template>
  <div class="mt-10">
    <h2 class="text-xl font-bold text-white mb-4">Plans by State</h2>
    <div class="rounded-xl border border-brand-900/20 bg-brand-950/40 p-4">
      <div ref="plotEl" class="w-full" style="min-height: 450px"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, nextTick } from 'vue'
import Plotly from 'plotly.js-dist-min'

const props = defineProps({
  stateCounts: { type: Object, required: true }
})

const plotEl = ref(null)

function render() {
  if (!plotEl.value) return
  const entries = Object.entries(props.stateCounts).filter(([, v]) => v > 0)
  const locations = entries.map(([k]) => k)
  const values = entries.map(([, v]) => v)

  const data = [{
    type: 'choropleth',
    locationmode: 'USA-states',
    locations,
    z: values,
    colorscale: [
      [0, 'rgb(247,251,255)'],
      [0.125, 'rgb(222,235,247)'],
      [0.25, 'rgb(198,219,239)'],
      [0.375, 'rgb(158,202,225)'],
      [0.5, 'rgb(107,174,214)'],
      [0.625, 'rgb(66,146,198)'],
      [0.75, 'rgb(33,113,181)'],
      [0.875, 'rgb(8,81,156)'],
      [1, 'rgb(8,48,107)']
    ],
    colorbar: {
      title: { text: 'Unique Plans', font: { color: '#94a3b8' } },
      tickfont: { color: '#94a3b8' },
      outlinewidth: 0
    },
    hovertemplate: 'State=%{location}<br>Unique Plans=%{z}<extra></extra>'
  }]

  const layout = {
    geo: {
      scope: 'usa',
      bgcolor: 'rgba(0,0,0,0)',
      landcolor: '#0B1E76',
      lakecolor: '#041739',
      showlakes: true,
      subunitcolor: '#1446B9',
      subunitwidth: 1,
      showsubunits: true,
      countrycolor: '#55B1FF',
      countrywidth: 1.5,
      showcountries: true
    },
    paper_bgcolor: 'rgba(0,0,0,0)',
    plot_bgcolor: 'rgba(0,0,0,0)',
    margin: { t: 10, b: 10, l: 10, r: 10 },
    dragmode: false
  }

  Plotly.newPlot(plotEl.value, data, layout, {
    responsive: true,
    displayModeBar: false
  })
}

onMounted(() => nextTick(render))
watch(() => props.stateCounts, render, { deep: true })
</script>
