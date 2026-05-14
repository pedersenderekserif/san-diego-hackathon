<template>
  <div class="max-w-6xl mx-auto px-6 py-20">
    <!-- Hero -->
    <div class="text-center mb-20">
      <div class="inline-flex items-center gap-2 bg-brand-500/10 text-brand-400 text-sm font-medium px-4 py-1.5 rounded-full border border-brand-500/20 mb-6">
        🌊 San Diego Hackathon · Serif Health
      </div>
      <h1 class="text-5xl md:text-6xl font-bold text-white mb-6 leading-tight">
        Who's <span class="text-brand-500">self-funding</span><br />their health benefits?
      </h1>
      <p class="text-xl text-slate-400 max-w-2xl mx-auto mb-10">
        Explore which employers are self-funding their health plans, which networks they've chosen, and whether those choices actually make sense — using real Transparency in Coverage data.
      </p>
      <div class="flex flex-col sm:flex-row gap-4 justify-center">
        <RouterLink
          to="/search"
          class="inline-flex items-center justify-center gap-2 bg-brand-500 hover:bg-brand-600 text-white font-semibold px-8 py-3 rounded-lg transition-colors"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          Search Employers
        </RouterLink>
        <a
          href="https://github.com/pedersenderekserif/san-diego-hackathon"
          target="_blank"
          rel="noopener"
          class="inline-flex items-center justify-center gap-2 bg-slate-800 hover:bg-slate-700 text-slate-100 font-semibold px-8 py-3 rounded-lg transition-colors"
        >
          View on GitHub
        </a>
      </div>
    </div>

    <!-- Stats row -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-20">
      <div v-for="stat in stats" :key="stat.label" class="bg-brand-950/40 border border-brand-900/20 rounded-xl p-6 text-center">
        <div class="text-3xl font-bold text-brand-400 mb-1">{{ stat.value }}</div>
        <div class="text-sm text-slate-500">{{ stat.label }}</div>
      </div>
    </div>

    <!-- How it works -->
    <div class="mb-20">
      <h2 class="text-2xl font-bold text-white mb-8 text-center">How it works</h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div v-for="step in steps" :key="step.title" class="bg-brand-950/40 border border-brand-900/20 rounded-xl p-6">
          <div class="text-3xl mb-4">{{ step.icon }}</div>
          <h3 class="text-lg font-semibold text-white mb-2">{{ step.title }}</h3>
          <p class="text-slate-400 text-sm leading-relaxed">{{ step.description }}</p>
        </div>
      </div>
    </div>

    <!-- CTA -->
    <div class="text-center bg-brand-950/40 border border-brand-900/20 rounded-2xl p-12">
      <h2 class="text-2xl font-bold text-white mb-3">Ready to dig in?</h2>
      <p class="text-slate-400 mb-6">Search by employer name or EIN to see their network choices and coverage data.</p>
      <RouterLink
        to="/search"
        class="inline-flex items-center gap-2 bg-brand-500 hover:bg-brand-600 text-white font-semibold px-8 py-3 rounded-lg transition-colors"
      >
        Start Searching →
      </RouterLink>
    </div>
  </div>
</template>

<script setup>
import { RouterLink } from 'vue-router'

const stats = [
  { value: '6M+', label: 'Form 5500 filings analyzed' },
  { value: '500+', label: 'Networks tracked' },
  { value: '100%', label: 'Public Transparency data' }
]

const steps = [
  {
    icon: '📋',
    title: 'Harvest EINs',
    description: 'We pull employer identification numbers from Form 5500 DOL filings to identify self-funded plans.'
  },
  {
    icon: '🔗',
    title: 'Cross-reference Networks',
    description: 'Each EIN is matched against reporting_plans to find which health networks the employer declared.'
  },
  {
    icon: '📊',
    title: 'Trace Price Data',
    description: 'We follow the MRF index to surface actual published pricing data and flag gaps in coverage.'
  }
]
</script>
