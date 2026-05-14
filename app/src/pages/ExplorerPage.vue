<template>
  <div class="max-w-6xl mx-auto px-6 py-12">
    <h1 class="text-3xl font-bold text-white mb-2">Explore by Issuer</h1>
    <p class="text-slate-400 mb-8">Browse self-funded plan sponsors grouped by insurance issuer.</p>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-16">
      <svg class="animate-spin h-8 w-8 text-brand-500 mx-auto mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
      </svg>
      <p class="text-slate-400">Loading data...</p>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="bg-red-500/10 border border-red-500/30 text-red-400 rounded-xl px-5 py-4 text-sm mb-6">
      {{ error }}
    </div>

    <!-- Main content -->
    <template v-else>
      <!-- Controls row -->
      <div class="flex flex-col sm:flex-row gap-6 mb-8">
        <!-- Issuer dropdown -->
        <div class="flex-1">
          <label for="issuer-select" class="block text-sm font-medium text-slate-300 mb-2">Select Issuer</label>
          <select
            id="issuer-select"
            v-model="selectedIssuer"
            class="w-full bg-brand-950/40 border border-brand-900/30 text-white rounded-xl px-4 py-3 focus:border-brand-500 focus:ring-1 focus:ring-brand-500 outline-none transition-colors"
          >
            <option value="" disabled>Choose an issuer...</option>
            <option v-for="name in issuerNames" :key="name" :value="name">{{ name }}</option>
          </select>
        </div>

        <!-- Stats card -->
        <div
          v-if="selectedIssuer"
          class="bg-purple-600/10 border border-purple-500/20 rounded-xl p-6 flex flex-col items-center justify-center min-w-[220px]"
        >
          <div class="text-3xl font-bold text-purple-400">{{ uniqueEinCount }}</div>
          <div class="text-sm text-purple-300/70 text-center mt-1">Total Number of Self-Funded Plans</div>
        </div>
      </div>

      <!-- Scrollable results table -->
      <div v-if="filteredRows.length > 0" class="rounded-xl border border-brand-900/20">
        <div class="max-h-[400px] overflow-y-auto overflow-x-auto">
          <table class="w-full text-sm text-left">
            <thead class="sticky top-0 z-10">
              <tr class="border-b border-brand-900/30 bg-brand-950">
                <th class="px-4 py-3 text-slate-400 font-medium">Plan Sponsor</th>
                <th class="px-4 py-3 text-slate-400 font-medium">Plan Name</th>
                <th class="px-4 py-3 text-slate-400 font-medium">EIN</th>
                <th class="px-4 py-3 text-slate-400 font-medium">City</th>
                <th class="px-4 py-3 text-slate-400 font-medium">State</th>
                <th class="px-4 py-3 text-slate-400 font-medium">Plan Year Begin</th>
                <th class="px-4 py-3 text-slate-400 font-medium">Effective Date</th>
                <th class="px-4 py-3 text-slate-400 font-medium">Admin Signed Name</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, i) in filteredRows"
                :key="i"
                class="border-b border-brand-900/10 hover:bg-brand-950/40 transition-colors"
              >
                <td class="px-4 py-3 text-white font-medium">{{ row.plan_sponsor_name }}</td>
                <td class="px-4 py-3 text-slate-300">{{ row.plan_name }}</td>
                <td class="px-4 py-3 text-slate-300 font-mono text-xs">{{ row.SPONS_DFE_EIN }}</td>
                <td class="px-4 py-3 text-slate-300">{{ row.SPONS_DFE_MAIL_US_CITY }}</td>
                <td class="px-4 py-3 text-slate-300">{{ row.SPONS_DFE_MAIL_US_STATE }}</td>
                <td class="px-4 py-3 text-slate-300">{{ row.FORM_PLAN_YEAR_BEGIN_DATE }}</td>
                <td class="px-4 py-3 text-slate-300">{{ row.PLAN_EFF_DATE }}</td>
                <td class="px-4 py-3 text-slate-300">{{ row.ADMIN_SIGNED_NAME }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="text-xs text-slate-500 px-4 py-3 bg-brand-950/30 border-t border-brand-900/20">
          Showing {{ filteredRows.length }} plan{{ filteredRows.length !== 1 ? 's' : '' }}
        </div>
      </div>

      <!-- Empty: issuer selected but no rows -->
      <div v-else-if="selectedIssuer" class="text-center py-16 text-slate-500">
        <p>No plans found for this issuer.</p>
      </div>

      <!-- Prompt to select -->
      <div v-else class="text-center py-16 text-slate-600">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-4 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
        </svg>
        <p class="text-sm">Select an issuer above to view their self-funded plan sponsors.</p>
      </div>

      <!-- Choropleth map -->
      <StateMap v-if="selectedIssuer && filteredRows.length > 0" :stateCounts="stateCounts" />
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Papa from 'papaparse'
import StateMap from '../components/StateMap.vue'

const allRows = ref([])
const selectedIssuer = ref('')
const loading = ref(true)
const error = ref(null)

const issuerNames = computed(() => {
  const names = new Set(allRows.value.map(r => r.issuer_name).filter(Boolean))
  return [...names].sort()
})

const filteredRows = computed(() => {
  if (!selectedIssuer.value) return []
  return allRows.value.filter(r => r.issuer_name === selectedIssuer.value)
})

const uniqueEinCount = computed(() => {
  const eins = new Set(filteredRows.value.map(r => r.SPONS_DFE_EIN).filter(Boolean))
  return eins.size
})

const stateCounts = computed(() => {
  const counts = {}
  for (const row of filteredRows.value) {
    const st = (row.SPONS_DFE_MAIL_US_STATE || '').trim().toUpperCase()
    if (st) counts[st] = (counts[st] || 0) + 1
  }
  return counts
})

onMounted(async () => {
  try {
    const res = await fetch('/data/output.csv')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const text = await res.text()

    const parsed = Papa.parse(text, {
      header: true,
      skipEmptyLines: true,
      transformHeader: (h) => h.trim()
    })

    allRows.value = parsed.data.filter(r => r.issuer_name)

    if (issuerNames.value.length === 1) {
      selectedIssuer.value = issuerNames.value[0]
    }
  } catch (e) {
    error.value = `Failed to load data: ${e.message}`
  } finally {
    loading.value = false
  }
})
</script>
