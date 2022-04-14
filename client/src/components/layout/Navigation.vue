<template>
  <nav
    id="sidebarMenu"
    class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
    <div class="position-sticky pt-3">
      <ul class="nav flex-column">
        <li class="nav-item">
          <RouterLink class="nav-link" to="/"> Dashboard </RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink class="nav-link" to="/trade"> Trade </RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink class="nav-link" to="/deposit"> Deposit </RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink class="nav-link" to="/withdraw">
            Transfer capital
          </RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink class="nav-link" to="/bank/transfer">
            International bank transfer
          </RouterLink>
        </li>

        <li class="nav-item">
          <RouterLink class="nav-link" to="/move">
            FTX MOVE Term Structure
          </RouterLink>
        </li>

        <li class="nav-item">
          <a class="nav-link" href="#"> FTX Move Delta Hedging </a>
        </li>

        <li class="nav-item">
          <a class="nav-link" href="#"> Reports </a>
        </li>
        <li class="nav-item">
          <a class="nav-link" href="#"> Integrations </a>
        </li>
      </ul>

      <h6
        class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted">
        <span>Saved reports</span>
        <a class="link-secondary" href="#" aria-label="Add a new report">
          <span data-feather="plus-circle"></span>
        </a>
      </h6>
      <ul class="nav flex-column mb-2">
        <li class="nav-item">
          <a class="nav-link" href="#">
            <span data-feather="file-text"></span>
            Current month
          </a>
        </li>
        <li class="nav-item">
          <a class="nav-link" href="#">
            <span data-feather="file-text"></span>
            Last quarter
          </a>
        </li>
        <!--            <li class="nav-item">-->
        <!--                <a class="nav-link" href="#">-->
        <!--                    <span data-feather="file-text"></span>-->
        <!--                    Social engagement-->
        <!--                </a>-->
        <!--            </li>-->
        <li class="nav-item">
          <a class="nav-link" href="#">
            <span data-feather="file-text"></span>
            Year-end sale
          </a>
        </li>
      </ul>
    </div>
  </nav>
</template>

<script>
import { RouterLink } from 'vue-router'

export default {
  name: 'Navigation',
  components: {
    RouterLink,
  },
}
</script>
