<script>
export default {
  props: ['modelValue'],
  emits: ['update:modelValue']
}
</script>

<template>
  <div class="input-group mb-3">
    <div class="form-check form-check-inline">
      <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          @input="$emit('update:modelValue', $event.target.value)"
          id="inlineRadio1"
          value="default"
          checked/>
      <label class="form-check-label" for="inlineRadio1">Default</label>
    </div>
    <div class="form-check form-check-inline">
      <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          @input="$emit('update:modelValue', $event.target.value)"
          id="inlineRadio2"
          value="erc20"/>
      <label class="form-check-label" for="inlineRadio2">ERC20</label>
    </div>
    <div class="form-check form-check-inline">
      <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          @input="$emit('update:modelValue', $event.target.value)"
          id="inlineRadio3"
          value="trx"/>
      <label class="form-check-label" for="inlineRadio2">TRX</label>
    </div>

    <div class="form-check form-check-inline">
      <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          @input="$emit('update:modelValue', $event.target.value)"
          id="inlineRadio4"
          value="sol"/>
      <label class="form-check-label" for="inlineRadio3">SOL</label>
    </div>

    <div class="form-check form-check-inline">
      <input
          class="form-check-input"
          type="radio"
          name="inlineRadioOptions"
          @input="$emit('update:modelValue', $event.target.value)"
          id="inlineRadio5"
          value="BNB"/>
      <label class="form-check-label" for="inlineRadio3">BNB</label>
    </div>
  </div>
</template>