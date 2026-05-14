<template>
  <div class="max-w-6xl mx-auto px-6 py-12">
    <h1 class="text-3xl font-bold text-white mb-2">Explore by Issuer</h1>
    <p class="text-slate-400 mb-8">Browse self-funded plan sponsors grouped by insurance issuer.</p>

    <!-- Controls row -->
    <div class="flex flex-col sm:flex-row gap-6 mb-8">
      <!-- Payor dropdown -->
      <div class="flex-1">
        <label for="payor-filter" class="block text-sm font-medium text-slate-300 mb-2">Payor Network</label>
        <select
          id="payor-filter"
          v-model="selectedPayorId"
          @change="onPayorChange"
          class="w-full bg-slate-900 border border-slate-700 focus:border-brand-500 focus:ring-1 focus:ring-brand-500 rounded-xl px-4 py-3 text-white outline-none transition-colors"
        >
          <option value="select_payor">Select Payor</option>
          <option
            v-for="payor in payorOptions"
            :key="payor.payor_id"
            :value="payor.payor_id"
          >
            {{ payor.payor_name }}
          </option>
        </select>
        <p v-if="payorLoading" class="mt-2 text-xs text-slate-500">Loading payors...</p>
        <p v-else-if="payorError" class="mt-2 text-xs text-red-400">{{ payorError }}</p>
      </div>

      <!-- Stats card -->
      <div
        v-if="results.length > 0"
        class="bg-purple-600/10 border border-purple-500/20 rounded-xl p-6 flex flex-col items-center justify-center min-w-[220px]"
      >
        <div class="text-3xl font-bold text-purple-400">{{ results.length }}</div>
        <div class="text-sm text-purple-300/70 text-center mt-1">Total Number of Self-Funded Plans</div>
      </div>
    </div>

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

    <!-- Scrollable results table -->
    <div v-else-if="results.length > 0" class="rounded-xl border border-brand-900/20">
      <div class="max-h-[400px] overflow-y-auto overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="sticky top-0 z-10">
            <tr class="border-b border-brand-900/30 bg-brand-950">
              <th class="px-4 py-3 text-slate-400 font-medium">Plan Sponsor</th>
              <th class="px-4 py-3 text-slate-400 font-medium">EIN</th>
              <th class="px-4 py-3 text-slate-400 font-medium">State</th>
              <th class="px-4 py-3 text-slate-400 font-medium">Plan Type</th>
              <th class="px-4 py-3 text-slate-400 font-medium">Employees</th>
              <th class="px-4 py-3 text-slate-400 font-medium">Networks</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, i) in results"
              :key="i"
              class="border-b border-brand-900/10 hover:bg-brand-950/40 transition-colors"
            >
              <td class="px-4 py-3 text-white font-medium">{{ row.name }}</td>
              <td class="px-4 py-3 text-slate-300 font-mono text-xs">{{ row.ein }}</td>
              <td class="px-4 py-3 text-slate-300">{{ row.state }}</td>
              <td class="px-4 py-3 text-slate-300">{{ row.planType }}</td>
              <td class="px-4 py-3 text-slate-300">{{ row.employees?.toLocaleString() ?? '—' }}</td>
              <td class="px-4 py-3">
                <span
                  v-for="network in row.networks"
                  :key="network"
                  class="inline-block bg-brand-500/10 text-brand-400 text-xs font-medium px-2 py-0.5 rounded-full border border-brand-500/20 mr-1"
                >
                  {{ network }}
                </span>
                <span v-if="row.networks.length === 0" class="text-slate-600 text-xs">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="text-xs text-slate-500 px-4 py-3 bg-brand-950/30 border-t border-brand-900/20">
        Showing {{ results.length }} plan{{ results.length !== 1 ? 's' : '' }}
      </div>
    </div>

    <!-- Empty: payor selected but no rows -->
    <div v-else-if="hasSearched && !loading" class="text-center py-16 text-slate-500">
      <p>No plans found for this payor.</p>
    </div>

    <!-- Prompt to select -->
    <div v-else class="text-center py-16 text-slate-600">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-4 opacity-30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
      </svg>
      <p class="text-sm">Select a payor network above to view their self-funded plan sponsors.</p>
    </div>

    <!-- Choropleth map -->
    <StateMap v-if="results.length > 0" :stateCounts="stateCounts" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import StateMap from '../components/StateMap.vue'

const results = ref([])
const loading = ref(false)
const hasSearched = ref(false)
const error = ref(null)

const payorOptions = ref([])
const payorLoading = ref(false)
const payorError = ref(null)
const selectedPayorId = ref('select_payor')

const aetnaEINs = ref(new Set())
const bcbsilEINs = ref(new Set())
const bcbstxEINs = ref(new Set())

const stateCounts = computed(() => {
  const counts = {}
  for (const row of results.value) {
    const st = (row.state || '').trim().toUpperCase()
    if (st) counts[st] = (counts[st] || 0) + 1
  }
  return counts
})

function mapFiling(f, index) {
  const plainEIN = f.spons_dfe_ein ? f.spons_dfe_ein.replace(/-/g, '') : ''
  const networks = []
  if (aetnaEINs.value.has(plainEIN)) networks.push('Aetna')
  if (bcbsilEINs.value.has(plainEIN)) networks.push('BCBS IL')
  if (bcbstxEINs.value.has(plainEIN)) networks.push('BCBS TX')
  const ein = f.spons_dfe_ein
  const name = f.sponsor_dfe_name || f.spons_dfe_dba_name
  return {
    _key: `${ein ?? ''}-${name ?? ''}-${index}`,
    ein,
    name,
    state: f.spons_dfe_mail_us_state,
    planType: f.plan_name || 'Unknown',
    employees: f.tot_act_rtd_sep_benef_cnt ? parseInt(f.tot_act_rtd_sep_benef_cnt, 10) || null : null,
    networks,
    hasPriceData: networks.length > 0
  }
}

onMounted(() => {
  ensurePayorOptions()
  loadAetnaEINs()
  loadBCBSILEINs()
  loadBCBSTXEINs()
})

async function ensurePayorOptions() {
  if (payorOptions.value.length > 0) return payorOptions.value

  payorLoading.value = true
  payorError.value = null

  try {
    const res = await fetch('/api/v1/index-templates/payors')
    const json = await res.json()
    if (!res.ok) throw new Error(json?.error?.message ?? 'Failed to load payors')
    payorOptions.value = json?.data ?? []
    return payorOptions.value
  } catch (fetchError) {
    payorError.value = fetchError?.message || 'Failed to load payors'
    return []
  } finally {
    payorLoading.value = false
  }
}

async function loadAetnaEINs() {
  try {
    const res = await fetch('/api/v1/aetna-mrf/plans')
    if (!res.ok) return
    const json = await res.json()
    const plans = json?.data ?? []
    const einSet = new Set()
    for (const plan of plans) {
      if (plan.plan_id_type === 'ein' || plan.plan_id_type === 'EIN') {
        einSet.add(plan.plan_id.replace(/-/g, ''))
      }
    }
    aetnaEINs.value = einSet
  } catch {
    // non-critical: silently ignore
  }
}

async function loadBCBSILEINs() {
  try {
    const res = await fetch('/api/v1/bcbsil-mrf/entries')
    if (!res.ok) return
    const json = await res.json()
    const entries = json?.data ?? []
    const einSet = new Set()
    for (const entry of entries) {
      if (entry.ein) einSet.add(entry.ein.replace(/-/g, ''))
    }
    bcbsilEINs.value = einSet
  } catch {
    // non-critical: silently ignore
  }
}

async function loadBCBSTXEINs() {
  try {
    const res = await fetch('/api/v1/bcbstx-mrf/entries')
    if (!res.ok) return
    const json = await res.json()
    const entries = json?.data ?? []
    const einSet = new Set()
    for (const entry of entries) {
      if (entry.ein) einSet.add(entry.ein.replace(/-/g, ''))
    }
    bcbstxEINs.value = einSet
  } catch {
    // non-critical: silently ignore
  }
}

function onPayorChange() {
  error.value = null
  results.value = []

  if (selectedPayorId.value === 'select_payor') {
    hasSearched.value = false
    return
  }

  loading.value = true
  fetchByPayor()
}

async function fetchByPayor() {
  try {
    const params = new URLSearchParams()
    params.set('q', '')
    params.append('payor_id', selectedPayorId.value)

    const res = await fetch(`/api/v1/form-5500?${params.toString()}`)
    const json = await res.json()
    if (!res.ok) {
      error.value = json?.error?.message ?? 'An error occurred'
      results.value = []
    } else {
      results.value = (json.data ?? []).map((f, i) => mapFiling(f, i))
    }
    hasSearched.value = true
  } catch {
    error.value = 'Failed to reach the API. Is the server running?'
    results.value = []
    hasSearched.value = true
  } finally {
    loading.value = false
  }
}
</script>
