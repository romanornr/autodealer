<template>
  <div class="container-fluid">
    <div class="row">
      <Header />
    </div>
    <div class="row">
      <Navigation />
      <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
        <div
          class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
          <h1 class="h2">{{ pageTitle }}</h1>
          <div class="btn-toolbar mb-2 mb-md-0">
            <div class="btn-group me-2">
              <button type="button" class="btn btn-sm btn-outline-secondary">
                Share
              </button>
              <button type="button" class="btn btn-sm btn-outline-secondary">
                Export
              </button>
            </div>
            <button
              type="button"
              class="btn btn-sm btn-outline-secondary dropdown-toggle">
              <span data-feather="calendar"></span>
              This week
            </button>
          </div>
        </div>
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script>
import { RouterView } from 'vue-router'

import Header from './components/layout/Header.vue'
import Navigation from '@/components/layout/Navigation.vue'

export default {
  name: 'App',
  components: {
    Navigation,
    Header,
    RouterView,
  },
  computed: {
    pageTitle() {
      return this.$route.meta.title
    },
  },
}
</script>

<style>
/* TODO */
/* @import '@/assets/base.css'; */
</style>
