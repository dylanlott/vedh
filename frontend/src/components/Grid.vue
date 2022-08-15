<template>
  <div class="grid-container">
    <div v-for="index in 10" :key="index" :class="`row row-${index}`">
      <div
        v-for="index in itemPerRow"
        :key="index"
        :class="`item item-${index}`"
      >
        {{ index }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "Grid",
  data() {
    return {
      itemPerRow: 30,
      itemWidth: "40px",
      itemHeight: "50px",
    };
  },
};
</script>

<style style="scss">
.grid-container {
  display: flex;
  flex-direction: column;
  position: absolute;
  top: 0;
  left: 0;
}

.row {
  flex: 1;
  display: flex;
  border-right: 1px solid #eee;
}

.row:last-child {
  border-bottom: 1px solid #eee;
}

.item {
  width: 40px;
  height: 50px;
  flex: 1;
  border-top: 1px solid #eee;
  border-left: 1px solid #eee;
  color: #eee;
}
</style>

