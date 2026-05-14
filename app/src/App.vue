<template>
  <div class="min-h-screen flex flex-col">
    <nav class="border-b border-slate-800 px-6 py-4">
      <div class="max-w-6xl mx-auto flex items-center justify-between">
        <RouterLink to="/" class="flex items-center gap-2 text-brand-500 font-semibold text-lg hover:text-brand-400 transition-colors">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" />
          </svg>
          NetExplorer
        </RouterLink>
        <div class="flex items-center gap-6 text-sm text-slate-400">
          <RouterLink to="/" class="hover:text-slate-100 transition-colors" active-class="text-slate-100">Home</RouterLink>
          <RouterLink to="/search" class="hover:text-slate-100 transition-colors" active-class="text-slate-100">Search</RouterLink>
        </div>
      </div>
    </nav>

    <main class="flex-1">
      <RouterView />
    </main>

    <footer class="border-t border-slate-800 px-6 py-4 text-center text-xs text-slate-600">
      Built at Serif Health's San Diego Hackathon 🌊 · Powered by Transparency in Coverage data
    </footer>
  </div>
</template>

<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>
