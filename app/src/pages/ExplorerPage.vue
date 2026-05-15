<template>
  <div class="max-w-6xl mx-auto px-6 py-12">
    <h1 class="text-3xl font-bold text-white mb-2">Explore by Issuer</h1>
    <p class="text-slate-400 mb-8">Browse self-funded plan sponsors grouped by insurance issuer.</p>

    <!-- Controls row -->
    <div class="flex flex-col sm:flex-row gap-6 mb-8">
      <!-- Payor dropdown + funding checkbox -->
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
        <div class="flex items-center gap-2 mt-3">
          <input
            id="funding-gen-asset-explorer"
            type="checkbox"
            v-model="fundingGenAssetOnly"
            @change="onPayorChange"
            class="h-4 w-4 rounded border-slate-600 bg-slate-900 text-brand-500 focus:ring-brand-500 focus:ring-offset-0"
          />
          <label for="funding-gen-asset-explorer" class="text-sm text-slate-400 select-none cursor-pointer">
            General assets only (<code class="text-xs text-slate-500">FUNDING_GEN_ASSET_IND = 1</code>)
          </label>
        </div>
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

    <!-- Results cards -->
    <div v-else-if="results.length > 0" class="max-h-[50vh] overflow-y-auto space-y-3 pr-1">
      <template v-for="(row, i) in results" :key="row._key ?? i">
        <!-- Employer card -->
        <div
          @click="selectEmployer(row)"
          @keydown="onEmployerKeydown($event, row)"
          role="button"
          tabindex="0"
          :class="[
            'bg-slate-900 border rounded-xl p-5 transition-colors cursor-pointer group',
            selectedEmployer?._key === row._key
              ? 'border-brand-500 rounded-b-none'
              : 'border-slate-800 hover:border-slate-600'
          ]"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="min-w-0">
              <h3 class="text-white font-semibold truncate group-hover:text-brand-400 transition-colors">{{ row.name }}</h3>
              <div class="flex items-center gap-3 mt-1 text-sm text-slate-500">
                <span>EIN: {{ row.ein }}</span>
                <span>·</span>
                <span>{{ row.state }}</span>
              </div>
            </div>
            <div class="flex flex-col items-end gap-1.5 shrink-0">
              <span
                v-for="network in row.networks"
                :key="network"
                class="inline-block bg-brand-500/10 text-brand-400 text-xs font-medium px-2.5 py-0.5 rounded-full border border-brand-500/20"
              >
                {{ network }}
              </span>
              <!-- Expand/collapse chevron -->
              <svg
                :class="['h-4 w-4 text-slate-500 transition-transform mt-1', selectedEmployer?._key === row._key ? 'rotate-180 text-brand-400' : '']"
                xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </div>
          </div>
          <div class="mt-3 flex items-center gap-4 text-xs text-slate-600">
            <span>{{ row.planType }} plan</span>
            <span v-if="row.planName">· Plan: {{ row.planName }}</span>
            <span v-if="row.planIdType">· Plan ID Type: {{ row.planIdType }}</span>
            <span v-if="row.employees">· {{ row.employees.toLocaleString() }} covered lives</span>
            <span v-if="row.hasPriceData" class="text-emerald-500">✓ Price data available</span>
            <span v-else class="text-amber-500">⚠ No price data</span>
          </div>
          <div v-if="row.adminName || row.adminPhone" class="mt-2 flex items-center gap-3 text-xs text-slate-500">
            <span class="text-slate-600 font-medium">Admin:</span>
            <span v-if="row.adminName">{{ row.adminName }}</span>
            <span v-if="row.adminCity || row.adminState">· {{ [row.adminCity, row.adminState, row.adminZip].filter(Boolean).join(', ') }}</span>
            <span v-if="row.adminPhone">· {{ row.adminPhone }}</span>
          </div>
        </div>

        <!-- Expansion row: reporting plans inline below selected employer -->
        <div
          v-if="selectedEmployer?._key === row._key"
          class="bg-slate-950 border border-brand-500 border-t-0 rounded-b-xl px-5 py-4 -mt-3"
        >
          <h2 class="text-sm font-semibold text-brand-400 mb-3">Reporting Plans</h2>

          <div v-if="reportingPlansLoading" class="text-slate-400 text-sm">Loading reporting plans…</div>

          <div v-else-if="reportingPlansError" class="bg-red-500/10 border border-red-500/30 text-red-400 rounded-xl px-4 py-3 text-sm">
            {{ reportingPlansError }}
          </div>

          <div v-else-if="expandedReportingPlans.length === 0" class="text-slate-500 text-sm">
            No reporting plans found for this employer.
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="plan in expandedReportingPlans"
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

      <div class="text-xs text-slate-500 px-1 py-2">
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
const fundingGenAssetOnly = ref(false)

const selectedEmployer = ref(null)
const expandedReportingPlans = ref([])
const reportingPlansLoading = ref(false)
const reportingPlansError = ref(null)

const payorOptions = ref([])
const payorLoading = ref(false)
const payorError = ref(null)
const selectedPayorId = ref('select_payor')

const reportingPlanFilters = ref(null)

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
    planName: null,
    planIdType: null,
    planMarketType: null,
    employees: f.tot_act_rtd_sep_benef_cnt ? parseInt(f.tot_act_rtd_sep_benef_cnt, 10) || null : null,
    networks,
    hasPriceData: networks.length > 0,
    adminName: f.admin_name || null,
    adminPhone: f.admin_phone_num || null,
    adminCity: f.admin_us_city || null,
    adminState: f.admin_us_state || null,
    adminZip: f.admin_us_zip || null,
  }
}

onMounted(() => {
  ensurePayorOptions()
  ensureReportingPlanFilters()
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

async function ensureReportingPlanFilters() {
  if (reportingPlanFilters.value) return reportingPlanFilters.value
  try {
    const res = await fetch('/api/v1/reporting-plans/filters')
    const json = await res.json()
    if (!res.ok) throw new Error(json?.error?.message ?? 'Failed to load filters')
    reportingPlanFilters.value = json?.data ?? null
    return reportingPlanFilters.value
  } catch {
    return null
  }
}

async function enrichWithReportingPlans() {
  const filters = await ensureReportingPlanFilters()
  if (!filters) return

  const eins = [...new Set(results.value.map(r => r.ein).filter(Boolean))]
  if (eins.length === 0) return

  const params = new URLSearchParams()
  for (const ein of eins) params.append('eins', ein)
  params.append('ingestor_ids', selectedPayorId.value)
  for (const t of filters.plan_id_types ?? []) params.append('plan_id_types', t)
  for (const m of filters.plan_market_types ?? []) params.append('plan_market_types', m)

  // BCBSIL/BCBSTX don't require plan_market_types — add a dummy if missing so the
  // API doesn't reject with "missing_filters" for non-BCBS payors.
  if ((filters.plan_market_types ?? []).length === 0) {
    params.append('plan_market_types', 'group')
  }

  try {
    const res = await fetch(`/api/v1/reporting-plans?${params.toString()}`)
    if (!res.ok) return
    const json = await res.json()
    const plans = json?.data ?? []

    const planByEIN = new Map()
    for (const p of plans) {
      const key = p.plan_id.replace(/-/g, '')
      if (!planByEIN.has(key)) planByEIN.set(key, p)
    }

    // Mutate through results.value so Vue's reactivity picks up the changes
    for (let i = 0; i < results.value.length; i++) {
      const key = (results.value[i].ein ?? '').replace(/-/g, '')
      const plan = planByEIN.get(key)
      if (plan) {
        results.value[i] = {
          ...results.value[i],
          planName: plan.plan_name || null,
          planIdType: plan.plan_id_type || null,
          planMarketType: plan.plan_market_type || null
        }
      }
    }
  } catch {
    // non-critical: silently ignore
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

function onEmployerKeydown(event, employer) {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    selectEmployer(employer)
  }
}

async function selectEmployer(employer) {
  if (selectedEmployer.value?._key === employer._key) {
    selectedEmployer.value = null
    expandedReportingPlans.value = []
    reportingPlansError.value = null
    return
  }

  selectedEmployer.value = employer
  reportingPlansError.value = null
  expandedReportingPlans.value = []
  reportingPlansLoading.value = true

  try {
    const filters = await ensureReportingPlanFilters()
    const params = new URLSearchParams()
    params.set('eins', employer.ein)
    const effectivePayorId = selectedPayorId.value !== 'select_payor' ? selectedPayorId.value : null
    if (effectivePayorId) {
      params.append('ingestor_ids', effectivePayorId)
    } else {
      for (const id of filters?.ingestor_ids ?? []) params.append('ingestor_ids', id)
    }
    for (const t of filters?.plan_id_types ?? []) params.append('plan_id_types', t)
    for (const m of filters?.plan_market_types ?? []) params.append('plan_market_types', m)

    const res = await fetch(`/api/v1/reporting-plans?${params.toString()}`)
    const json = await res.json()
    if (!res.ok) {
      reportingPlansError.value = json?.error?.message ?? 'Failed to fetch reporting plans'
      return
    }
    expandedReportingPlans.value = json?.data ?? []
  } catch (fetchError) {
    reportingPlansError.value = fetchError?.message || 'Failed to fetch reporting plans'
  } finally {
    reportingPlansLoading.value = false
  }
}

function onPayorChange() {
  error.value = null
  results.value = []
  selectedEmployer.value = null
  expandedReportingPlans.value = []
  reportingPlansError.value = null

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
    if (fundingGenAssetOnly.value) {
      params.set('funding_gen_asset_ind', '1')
    }

    const res = await fetch(`/api/v1/form-5500?${params.toString()}`)
    const json = await res.json()
    if (!res.ok) {
      error.value = json?.error?.message ?? 'An error occurred'
      results.value = []
    } else {
      const mapped = (json.data ?? []).map((f, i) => mapFiling(f, i))
      results.value = mapped
      enrichWithReportingPlans()
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
