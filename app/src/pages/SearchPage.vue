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

    <div class="mb-4">
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
      <template v-for="employer in results" :key="employer._key">
        <!-- Employer card -->
        <div
          @click="selectEmployer(employer)"
          @keydown="onEmployerKeydown($event, employer)"
          role="button"
          tabindex="0"
          :class="[
            'bg-slate-900 border rounded-xl p-5 transition-colors cursor-pointer group',
            selectedEmployer?._key === employer._key
              ? 'border-brand-500 rounded-b-none'
              : 'border-slate-800 hover:border-slate-600'
          ]"
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
              <!-- Expand/collapse chevron -->
              <svg
                :class="['h-4 w-4 text-slate-500 transition-transform mt-1', selectedEmployer?._key === employer._key ? 'rotate-180 text-brand-400' : '']"
                xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </div>
          </div>
          <div class="mt-3 flex items-center gap-4 text-xs text-slate-600">
            <span>{{ employer.planType }} plan</span>
            <span v-if="employer.employees">· {{ employer.employees.toLocaleString() }} employees</span>
            <span v-if="employer.hasPriceData" class="text-emerald-500">✓ Price data available</span>
            <span v-else class="text-amber-500">⚠ No price data</span>
          </div>
        </div>

        <!-- Expansion row: reporting plans inline below selected employer -->
        <div
          v-if="selectedEmployer?._key === employer._key"
          class="bg-slate-950 border border-brand-500 border-t-0 rounded-b-xl px-5 py-4 -mt-3"
        >
          <h2 class="text-sm font-semibold text-brand-400 mb-3">Reporting Plans</h2>

          <div v-if="reportingPlansLoading" class="text-slate-400 text-sm">Loading reporting plans…</div>

          <div v-else-if="reportingPlansError" class="bg-red-500/10 border border-red-500/30 text-red-400 rounded-xl px-4 py-3 text-sm">
            {{ reportingPlansError }}
          </div>

          <div v-else-if="reportingPlans.length === 0" class="text-slate-500 text-sm">
            No reporting plans found for this employer.
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="plan in reportingPlans"
              :key="plan.id"
              class="bg-slate-900 border border-slate-800 rounded-xl px-4 py-3"
            >
              <div class="flex items-center justify-between gap-3">
                <h3 class="text-sm font-medium text-white truncate">{{ plan.plan_name || 'Unnamed Plan' }}</h3>
                <span class="text-xs text-slate-400 shrink-0">{{ plan.plan_market_type }}</span>
              </div>
              <div class="mt-1 text-xs text-slate-500">
                <span>ID: {{ plan.plan_id }}</span>
                <span> · </span>
                <span>Type: {{ plan.plan_id_type }}</span>
                <span v-if="plan.issuer_name"> · Issuer: {{ plan.issuer_name }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
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
      <p class="text-sm">Select a payor network or type an employer name / EIN above to get started</p>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'

const query = ref('')
const loading = ref(false)
const hasSearched = ref(false)
const activeFilter = ref('all')
const results = ref([])
const error = ref(null)
const selectedEmployer = ref(null)
const reportingPlans = ref([])
const reportingPlansLoading = ref(false)
const reportingPlansError = ref(null)
const reportingPlanFilters = ref(null)
const payorOptions = ref([])
const payorLoading = ref(false)
const payorError = ref(null)
const selectedPayorId = ref('')
const aetnaEINs = ref(new Set())
const bcbsilEINs = ref(new Set())
const bcbstxEINs = ref(new Set())

const filters = [
  { label: 'All Plans', value: 'all' },
  { label: 'Self-Funded', value: 'self-funded' },
  { label: 'Level-Funded', value: 'level-funded' },
  { label: 'Has Price Data', value: 'has-price-data' }
]

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
    industry: null,
    hasPriceData: networks.length > 0
  }
}

function mapReportingPlan(plan) {
  return {
    id: plan.id,
    plan_id: plan.plan_id,
    plan_name: plan.plan_name,
    plan_id_type: plan.plan_id_type,
    plan_market_type: plan.plan_market_type,
    issuer_name: plan.issuer_name
  }
}

let debounceTimer = null

onMounted(() => {
  ensurePayorOptions()
  loadAetnaEINs()
  loadBCBSILEINs()
  loadBCBSTXEINs()
})

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

function onInput() {
  clearTimeout(debounceTimer)
  error.value = null
  if (!query.value.trim()) {
    selectedEmployer.value = null
    reportingPlans.value = []
    reportingPlansError.value = null
    if (selectedPayorId.value) {
      loading.value = true
      debounceTimer = setTimeout(() => search(''), 350)
    } else {
      results.value = []
      hasSearched.value = false
    }
    return
  }
  loading.value = true
  debounceTimer = setTimeout(() => {
    search(query.value.trim())
  }, 350)
}

async function ensureReportingPlanFilters() {
  if (reportingPlanFilters.value) {
    return reportingPlanFilters.value
  }

  const res = await fetch('/api/v1/reporting-plans/filters')
  const json = await res.json()
  if (!res.ok) {
    throw new Error(json?.error?.message ?? 'Failed to load reporting plan filters')
  }

  reportingPlanFilters.value = json?.data ?? null
  return reportingPlanFilters.value
}

async function ensurePayorOptions() {
  if (payorOptions.value.length > 0) {
    return payorOptions.value
  }

  payorLoading.value = true
  payorError.value = null

  try {
    const res = await fetch('/api/v1/index-templates/payors')
    const json = await res.json()
    if (!res.ok) {
      throw new Error(json?.error?.message ?? 'Failed to load payors')
    }

    payorOptions.value = json?.data ?? []
    return payorOptions.value
  } catch (fetchError) {
    payorError.value = fetchError?.message || 'Failed to load payors'
    return []
  } finally {
    payorLoading.value = false
  }
}

function reportingPlanParams(ein, filters, selectedPayor) {
  const params = new URLSearchParams()
  params.set('eins', ein)

  if (selectedPayor) {
    params.append('ingestor_ids', selectedPayor)
  } else {
    for (const ingestorID of filters?.ingestor_ids ?? []) {
      params.append('ingestor_ids', ingestorID)
    }
  }
  for (const planIDType of filters?.plan_id_types ?? []) {
    params.append('plan_id_types', planIDType)
  }
  for (const planMarketType of filters?.plan_market_types ?? []) {
    params.append('plan_market_types', planMarketType)
  }

  return params
}

function onPayorChange() {
  selectedEmployer.value = null
  reportingPlans.value = []
  reportingPlansError.value = null
  error.value = null

  if (selectedPayorId.value === 'select_payor') {
    results.value = []
    hasSearched.value = false
    return
  }

  const trimmed = query.value.trim()
  if (!trimmed && !selectedPayorId.value) {
    results.value = []
    hasSearched.value = false
    return
  }

  loading.value = true
  search(trimmed)
}

function onEmployerKeydown(event, employer) {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    selectEmployer(employer)
  }
}

async function selectEmployer(employer) {
  // Toggle: clicking the same employer again collapses the expansion row
  if (selectedEmployer.value?._key === employer._key) {
    selectedEmployer.value = null
    reportingPlans.value = []
    reportingPlansError.value = null
    return
  }

  selectedEmployer.value = employer
  reportingPlansError.value = null
  reportingPlans.value = []
  reportingPlansLoading.value = true

  try {
    const filters = await ensureReportingPlanFilters()
    const effectivePayorId = selectedPayorId.value !== 'select_payor' ? selectedPayorId.value : null
    const params = reportingPlanParams(employer.ein, filters, effectivePayorId)
    const res = await fetch(`/api/v1/reporting-plans?${params.toString()}`)
    const json = await res.json()
    if (!res.ok) {
      reportingPlansError.value = json?.error?.message ?? 'Failed to fetch reporting plans'
      return
    }

    reportingPlans.value = (json?.data ?? []).map(mapReportingPlan)
  } catch (fetchError) {
    reportingPlansError.value = fetchError?.message || 'Failed to fetch reporting plans'
  } finally {
    reportingPlansLoading.value = false
  }
}

async function search(q) {
  results.value = []
  try {
    const params = new URLSearchParams()
    params.set('q', q)
    if (selectedPayorId.value && selectedPayorId.value !== 'select_payor') {
      params.append('payor_id', selectedPayorId.value)
    }

    const res = await fetch(`/api/v1/form-5500?${params.toString()}`)
    const json = await res.json()
    if (!res.ok) {
      error.value = json?.error?.message ?? 'An error occurred'
      results.value = []
    } else {
      results.value = (json.data ?? []).map((f, i) => mapFiling(f, i))
      selectedEmployer.value = null
      reportingPlans.value = []
      reportingPlansError.value = null
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
