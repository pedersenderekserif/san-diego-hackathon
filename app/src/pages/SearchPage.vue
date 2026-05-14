<template>
  <div class="max-w-4xl mx-auto px-6 py-12">
    <h1 class="text-3xl font-bold text-white mb-2">Search Employers</h1>
    <p class="text-slate-400 mb-8">Find a self-funded employer by name or EIN to explore their health network choices.</p>

    <!-- Search input -->
    <div class="relative mb-8">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        class="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-slate-500 pointer-events-none"
        fill="none" viewBox="0 0 24 24" stroke="currentColor"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input
        v-model="query"
        @input="onInput"
        type="search"
        placeholder="Search by employer name or EIN…"
        class="w-full bg-brand-950/40 border border-brand-900/30 focus:border-brand-500 focus:ring-1 focus:ring-brand-500 rounded-xl pl-12 pr-4 py-4 text-white placeholder-slate-500 outline-none transition-colors"
      />
      <div v-if="loading" class="absolute right-4 top-1/2 -translate-y-1/2">
        <svg class="animate-spin h-5 w-5 text-brand-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
        </svg>
      </div>
    </div>

    <!-- Filter pills -->
    <div class="flex flex-wrap gap-2 mb-8">
      <button
        v-for="filter in filters"
        :key="filter.value"
        @click="activeFilter = filter.value"
        :class="[
          'px-4 py-1.5 rounded-full text-sm font-medium transition-colors border',
          activeFilter === filter.value
            ? 'bg-brand-500 border-brand-500 text-white'
            : 'bg-transparent border-slate-700 text-slate-400 hover:border-slate-500 hover:text-slate-200'
        ]"
      >
        {{ filter.label }}
      </button>
    </div>

    <!-- Error state -->
    <div v-if="error" class="mb-6 bg-red-500/10 border border-red-500/30 text-red-400 rounded-xl px-5 py-4 text-sm">
      {{ error }}
    </div>

    <!-- Results -->
    <div v-if="results.length > 0" class="space-y-3">
      <div
        v-for="employer in results"
        :key="employer.ein"
        class="bg-brand-950/40 border border-brand-900/20 hover:border-brand-800/40 rounded-xl p-5 transition-colors cursor-pointer group"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <h3 class="text-white font-semibold truncate group-hover:text-brand-400 transition-colors">{{ employer.name }}</h3>
            <div class="flex items-center gap-3 mt-1 text-sm text-slate-500">
              <span>EIN: {{ employer.ein }}</span>
              <span>·</span>
              <span>{{ employer.state }}</span>
              <span v-if="employer.industry">·</span>
              <span v-if="employer.industry">{{ employer.industry }}</span>
            </div>
          </div>
          <div class="flex flex-col items-end gap-1.5 shrink-0">
            <span
              v-for="network in employer.networks"
              :key="network"
              class="inline-block bg-brand-500/10 text-brand-400 text-xs font-medium px-2.5 py-0.5 rounded-full border border-brand-500/20"
            >
              {{ network }}
            </span>
          </div>
        </div>
        <div class="mt-3 flex items-center gap-4 text-xs text-slate-600">
          <span>{{ employer.planType }} plan</span>
          <span v-if="employer.employees">· {{ employer.employees.toLocaleString() }} employees</span>
          <span v-if="employer.hasPriceData" class="text-emerald-500">✓ Price data available</span>
          <span v-else class="text-amber-500">⚠ No price data</span>
        </div>
      </div>
    </div>

    <!-- Empty state (after search) -->
    <div v-else-if="hasSearched && !loading" class="text-center py-16 text-slate-500">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-4 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p class="text-lg font-medium mb-1">No results found</p>
      <p class="text-sm">Try a different name or EIN</p>
    </div>

    <!-- Initial empty state -->
    <div v-else-if="!hasSearched" class="text-center py-16 text-slate-600">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-4 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <p class="text-sm">Type an employer name or EIN above to get started</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const query = ref('')
const loading = ref(false)
const hasSearched = ref(false)
const activeFilter = ref('all')
const results = ref([])
const error = ref(null)

const filters = [
  { label: 'All Plans', value: 'all' },
  { label: 'Self-Funded', value: 'self-funded' },
  { label: 'Level-Funded', value: 'level-funded' },
  { label: 'Has Price Data', value: 'has-price-data' }
]

function mapFiling(f) {
  return {
    ein: f.spons_dfe_ein,
    name: f.sponsor_dfe_name || f.spons_dfe_dba_name,
    state: f.spons_dfe_mail_us_state,
    planType: f.plan_name || 'Unknown',
    employees: f.tot_act_rtd_sep_benef_cnt ? parseInt(f.tot_act_rtd_sep_benef_cnt, 10) || null : null,
    networks: [],
    industry: null,
    hasPriceData: false
  }
}

let debounceTimer = null

function onInput() {
  clearTimeout(debounceTimer)
  error.value = null
  if (!query.value.trim()) {
    results.value = []
    hasSearched.value = false
    return
  }
  loading.value = true
  debounceTimer = setTimeout(() => {
    search(query.value.trim())
  }, 350)
}

async function search(q) {
  try {
    const res = await fetch(`/api/v1/form-5500?q=${encodeURIComponent(q)}`)
    const json = await res.json()
    if (!res.ok) {
      error.value = json?.error?.message ?? 'An error occurred'
      results.value = []
    } else {
      results.value = (json.data ?? []).map(mapFiling)
    }
    hasSearched.value = true
  } catch (e) {
    error.value = 'Failed to reach the API. Is the server running?'
    results.value = []
    hasSearched.value = true
  } finally {
    loading.value = false
  }
}
</script>
