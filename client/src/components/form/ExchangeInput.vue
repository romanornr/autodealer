<script>
export default {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  data() {
    return {
      exchanges: [
        {text: 'Binance', value: 'binance'},
        {text: 'FTX', value: 'ftx'},
        {text: 'Deribit', value: 'deribit'},
        {text: 'Kraken', value: 'kraken'},
        {text: 'BTSE', value: 'btse'},
        {text: 'Bittrex', value: 'bittrex'},
      ],
    }
  }
}
</script>

<template>
  <span v-for="exchange in exchanges">
     <input type="radio" class="btn-check" name="options-outlined" :value="exchange.value" :id="'checkbox-' + exchange.value" @input="$emit('update:modelValue', $event.target.value)"/>
      <label class="btn btn-outline-warning" :for="'checkbox-' + exchange.value">{{exchange.text}}</label>
  </span>
</template>