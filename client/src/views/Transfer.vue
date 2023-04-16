<template>
<!--  <div>Transfer</div>-->
  <form class="transfer-form" @submit.prevent="onSubmit">
    <ExchangeInput v-model="exchangeSource" label="exchangeSource" id="exchangeSource"/>
    <ExchangeInput v-model="exchangeDestination" label="exchangeDestination" id="exchangeDestination"/>
    <div class="row">
      <div class="col">
        <v-select
            v-model="asset"
            :options="assets"
            label="code"
            placeholder="BTC"></v-select>
      </div>
    </div>

    <!-- {{template "chain"}} -->
    <ChainInput v-model="chain"/>
    <button class="btn btn-primary" type="submit" :disabled="loading">
      <span
          v-if="loading"
          class="spinner-border spinner-border-sm"
          role="status"
          aria-hidden="true"></span>
      <span v-if="loading">Loading...</span><span v-else>Deposit</span>
    </button>
  </form>
  <br/>

  <div v-if="loading" class="spinner-border text-primary" role="status">
    <span class="sr-only"></span>
  </div>

    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="address" class="content">
      <p>Address: {{ address }}</p>
      <p>Balance: {{ exchangeSourceBalance }} {{ symbol }}</p>
      <!-- Show USD balance -->
      <!-- green if positive balance -->
      <div v-if="exchangeSourceBalanceUSD > 0">
        <p style="color:green;">USD: ${{ exchangeSourceBalanceUSD.toFixed(2) }}</p>
      </div>
      <!-- red if negative balance -->
      <div v-else>
        <p style="color:red;">USD: {{ exchangeSourceBalanceUSD.toFixed(2) }}</p>
      </div>
      <!-- end show USD balance -->
    </div>
  </template>

<!--<script>-->
<!--// TODO-->
<!--export default {-->
<!--  name: 'Transfer',-->
<!--}-->


<!--</script>-->

<script>
import vSelect from 'vue-select'
import ChainInput from "@/components/form/ChainInput.vue";
import ExchangeInput from "@/components/form/ExchangeInput.vue";

export default {
  name: 'Transfer',
  components: {
    ExchangeInput,
    ChainInput,
    vSelect,
  },
  data() {
    return {
      loading: false,
      assetData: '',
      error: '',
      options: [
        {text: 'Binance', value: 'binance'},
        {text: 'FTX', value: 'ftx'},
        {text: 'Deribit', value: 'deribit'},
        {text: 'Kraken', value: 'kraken'},
        {text: 'BTSE', value: 'btse'},
      ],
      asset: {code: 'USDT'},
      assets: [],
      //exchangeInput: '',
      exchangeSource: 'binance',
      exchangeDestination: '',
      address: '',
      exchangeSourceBalance: '',
      exchangeSourceBalanceUSD: '',
      price: 0,
      chain: 'default',
    }
  },
  async created() {
    await this.fetchAssets()
  },
  methods: {
    async fetchAssets() {
      this.assets = [] // clear the array

      try {
        const {
          data: {assets},
        } = await this.$api.get(`assets/${this.exchangeSource}`)

        this.assets = assets
      } catch (error) {
        console.log(error)
      }
    },
    async onSubmit() {
      // reset previous result
      this.error = ''

      // reset address and balance for new result
      this.address = ''
      this.balance = ''

      // disable the form
      this.loading = true

      try {
        const response = await this.$api.get(
            `deposit/${this.exchangeSource}/${this.asset.code}/${this.chain}`
        )

        this.processData(response)
      } catch (error) {
        console.log(error)
        this.error = error
      } finally {
        this.loading = false
      }
    },
    processData(result) {
      const data = result.data
      this.assetData = data
      this.address = data.address['Address']
      this.symbol = data.code
      this.exchangeSourceBalance = data.balance
      this.price = data.price
      this.exchangeSourceBalanceUSD = data.balance * data.price
      // this.exchangeName = data.exchange
    },
  },
  watch: {
    exchangeName: {
      async handler() {
        await this.fetchAssets()
      },
    },
  },
}
</script>
