<template>
  <div class="app-shell">
    <AppNav />
    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router';
import AppNav from './components/layout/AppNav.vue';
</script>
