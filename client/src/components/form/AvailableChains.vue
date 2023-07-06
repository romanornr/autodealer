<template>
<ExchangeInput></ExchangeInput>
<ChainInput></ChainInput>
</template>


<script>

import vSelect from 'vue-select';
import ChainInput from "@/components/form/ChainInput.vue";
import ExchangeInput from "@/components/form/ExchangeInput.vue";

export default {
  props: ['modelValue'],
  emits: ['update:modelValue'],
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
      chains: [],
      exchangeInput: '',
      exchangeName: 'binance',
      address: '',
      balance: '',
      balanceUSD: '',
      price: 0,
      chain: 'default',
    }
  },
  methods: {
    async fetchAssets() {
      this.assets = [] // clear the array

      try {
        const {
          data: {assets},
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
    async fetchChains() {
      this.chains = [] // clear the array
      try {
        const {
          data: {chains},
        } = await this.$api.get(`transfer/chains/${this.exchangeName}/${this.asset}`)
        this.chains = chains
      } catch (error) {
        console.log(error)
      }
    },
  },
  watch: {
    exchangeName: {
      async handler() {
        await this.fetchAssets()
      },
    },
    assetName: {
      async handler() {
        await this.fetchChains()
      },
    },
  },
}



// import vSelect from 'vue-select'
// import ChainInput from "@/components/form/ChainInput.vue";
// import ExchangeInput from "@/components/form/ExchangeInput.vue";
// import AvailableChains from "@/components/form/AvailableChains.vue";
//
// export default {
//   name: 'Deposit',
//   components: {
//     ExchangeInput,
//     ChainInput,
//     vSelect,
//   },
//   data() {
//     return {
//       loading: false,
//       assetData: '',
//       error: '',
//       options: [
//         {text: 'Binance', value: 'binance'},
//         {text: 'FTX', value: 'ftx'},
//         {text: 'Deribit', value: 'deribit'},
//         {text: 'Kraken', value: 'kraken'},
//         {text: 'BTSE', value: 'btse'},
//       ],
//       asset: {code: 'USDT'},
//       assets: [],
//       exchangeInput: '',
//       exchangeName: 'binance',
//       address: '',
//       balance: '',
//       balanceUSD: '',
//       price: 0,
//       chain: 'default',
//     }
//   },
//   async created() {
//     await this.fetchAssets()
//   },
//   methods: {
//     async fetchAssets() {
//       this.assets = [] // clear the array
//
//       try {
//         const {
//           data: {assets},
//         } = await this.$api.get(`assets/${this.exchangeName}`)
//
//         this.assets = assets
//       } catch (error) {
//         console.log(error)
//       }
//     },
//     async onSubmit() {
//       // reset previous result
//       this.error = ''
//
//       // reset address and balance for new result
//       this.address = ''
//       this.balance = ''
//
//       // disable the form
//       this.loading = true
//
//       try {
//         const response = await this.$api.get(
//             `deposit/${this.exchangeName}/${this.asset.code}/${this.chain}`
//         )
//
//         this.processData(response)
//       } catch (error) {
//         console.log(error)
//         this.error = error
//       } finally {
//         this.loading = false
//       }
//     },
//     processData(result) {
//       const data = result.data
//       this.assetData = data
//       this.address = data.address['Address']
//       this.symbol = data.code
//       this.balance = data.balance
//       this.price = data.price
//       this.balanceUSD = data.balance * data.price
//       // this.exchangeName = data.exchange
//     },
//   },
//   watch: {
//     exchangeName: {
//       async handler() {
//         await this.fetchAssets()
//       },
//     },
//   },
// }
</script>