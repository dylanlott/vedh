<template>
  <div class="grid-container">
  </div>
</template>

<script>
export default {
  name: "Grid",
  data() {
    return {
    };
  },
};
</script>

<style style="scss">
.grid-container {
  display: flex;
  flex-direction: column;
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

