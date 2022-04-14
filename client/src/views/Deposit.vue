<template>
  <form class="deposit-form" @submit.prevent="onSubmit">
    <div class="input-group mb-3">
      <!-- {{template "exchangesRadioButtons"}} -->
      <span class="input-group-text" id="basic-addon">Exchange</span>
      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="ftx"
        id="ftx-btn-check-outlined"
        autocomplete="off"
        checked
        required />
      <label class="btn btn-outline-primary" for="ftx-btn-check-outlined"
        >FTX</label
      ><br />

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="binance"
        id="binance-btn-check-2-outlined"
        autocomplete="off" />
      <label class="btn btn-outline-warning" for="binance-btn-check-2-outlined"
        >Binance</label
      ><br />

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="deribit"
        id="deribit-success-outlined"
        autocomplete="off" />
      <label class="btn btn-outline-success" for="deribit-success-outlined"
        >Deribit</label
      >

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="bitmex"
        id="bitmex-danger-outlined"
        autocomplete="off" />
      <label class="btn btn-outline-danger" for="bitmex-danger-outlined"
        >Bitmex</label
      >

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="huobi"
        id="huobi-info-outlined"
        autocomplete="off" />
      <label class="btn btn-outline-info" for="huobi-info-outlined"
        >Huobi</label
      >

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="bitfinex"
        id="finex-success-outlined"
        autocomplete="off" />
      <label class="btn btn-outline-success" for="finex-success-outlined"
        >Bitfinex</label
      >

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="btse"
        id="btse-info-outlined"
        autocomplete="off" />
      <label class="btn btn-outline-info" for="btse-info-outlined">BTSE</label>

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="kraken"
        id="kraken-dark-outlined"
        autocomplete="off" />
      <label class="btn btn-outline-dark" for="kraken-dark-outlined"
        >Kraken</label
      >

      <input
        type="radio"
        class="btn-check"
        name="options-outlined"
        v-model="exchangeName"
        value="bittrex"
        id="bittrex-info-outlined"
        autocomplete="off" />
    </div>

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
    <div class="input-group mb-3">
      <div class="form-check form-check-inline">
        <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          v-model="chain"
          id="inlineRadio1"
          value="default"
          checked />
        <label class="form-check-label" for="inlineRadio1">Default</label>
      </div>
      <div class="form-check form-check-inline">
        <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          v-model="chain"
          id="inlineRadio2"
          value="erc20" />
        <label class="form-check-label" for="inlineRadio2">ERC20</label>
      </div>
      <div class="form-check form-check-inline">
        <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          v-model="chain"
          id="inlineRadio3"
          value="trx" />
        <label class="form-check-label" for="inlineRadio2">TRX</label>
      </div>

      <div class="form-check form-check-inline">
        <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          v-model="chain"
          id="inlineRadio4"
          value="sol" />
        <label class="form-check-label" for="inlineRadio3">SOL</label>
      </div>

      <div class="form-check form-check-inline">
        <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          v-model="chain"
          id="inlineRadio5"
          value="BNB" />
        <label class="form-check-label" for="inlineRadio3">BNB</label>
      </div>
    </div>

    <button class="btn btn-primary" type="submit" :disabled="loading">
      <span
        v-if="loading"
        class="spinner-border spinner-border-sm"
        role="status"
        aria-hidden="true"></span>
      <span v-if="loading">Loading...</span><span v-else>Deposit</span>
    </button>
  </form>
  <br />

  <div v-if="loading" class="spinner-border text-primary" role="status">
    <span class="sr-only"></span>
  </div>

  <div v-if="error" class="error">{{ error }}</div>

  <div v-if="address" class="content">
    <p>Address: {{ address }}</p>
    <p>Balance: {{ balance }} {{ symbol }} (${{ balanceUSD.toFixed(2) }})</p>
  </div>
</template>

<script>
import vSelect from 'vue-select'

export default {
  name: 'Deposit',
  components: {
    vSelect,
  },
  data() {
    return {
      loading: false,
      assetData: '',
      error: '',
      options: [
        { text: 'FTX', value: 'ftx' },
        { text: 'Binance', value: 'binance' },
        { text: 'Deribit', value: 'deribit' },
        { text: 'Kraken', value: 'kraken' },
        { text: 'BTSE', value: 'btse' },
      ],
      asset: { code: 'USDT' },
      assets: [],
      exchangeInput: '',
      exchangeName: 'ftx',
      address: '',
      balance: '',
      balanceUSD: '',
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
          data: { assets },
        } = await this.$api.get(`assets/${this.exchangeName}`)

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
          `deposit/${this.exchangeName}/${this.asset.code}/${this.chain}`
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
      this.balance = data.balance
      this.price = data.price
      this.balanceUSD = data.balance * data.price
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
