<script>
export default {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  data() {
    return {
      exchanges: [
        {text: 'Binance', value: 'binance', btn_class: 'btn btn-outline-warning'},
        {text: 'Deribit', value: 'deribit', btn_class: 'btn btn-outline-success'},
        {text: 'Kraken', value: 'kraken', btn_class: 'btn btn-outline-dark'},
        {text: 'BTSE', value: 'btse', btn_class: 'btn btn-outline-info'},
        {text: 'Bittrex', value: 'bittrex', btn_class: 'btn btn-outline-info'},
        {text: 'FTX', value: 'ftx', btn_class: 'btn btn-outline-primary'},
      ],
    }
  }
}
</script>

<template>
  <div class="input-group mb-3">
    <span class="input-group-text" id="basic-addon">Exchange</span>
    <span v-for="exchange in exchanges" :key="exchange.value">
      <input type="radio" class="btn-check" name="options-outlined" :value="exchange.value" :id="'checkbox-' + exchange.value" @input="$emit('update:modelValue', $event.target.value)"/>
      <label :class="exchange.btn_class" :for="'checkbox-' + exchange.value">{{exchange.text}}</label>
    </span>
  </div>
</template>